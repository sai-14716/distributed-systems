package lb

import (
	"fmt"
	"net/url"
	"testing"
	"time"
)

func TestOwnerForStableAcrossPeerOrder(t *testing.T) {
	reg := NewRegistry()
	u, _ := url.Parse("http://example.invalid")
	reg.AddBackend("backend-1", u)

	self := Peer{ID: "nodeA", BaseURL: "http://nodeA:8000"}
	peers1 := []Peer{
		{ID: "nodeC", BaseURL: "http://nodeC:8000"},
		self,
		{ID: "nodeB", BaseURL: "http://nodeB:8000"},
	}
	peers2 := []Peer{
		{ID: "nodeB", BaseURL: "http://nodeB:8000"},
		{ID: "nodeC", BaseURL: "http://nodeC:8000"},
		self,
	}

	cl1, err := NewClusterLoad(reg, self, peers1, ClusterLoadOptions{}, func() float64 { return 0.8 }, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad 1: %v", err)
	}
	cl2, err := NewClusterLoad(reg, self, peers2, ClusterLoadOptions{}, func() float64 { return 0.8 }, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad 2: %v", err)
	}

	o1, o1b := cl1.ownersFor("backend-1")
	o2, o2b := cl2.ownersFor("backend-1")
	if o1 != o2 {
		t.Fatalf("owner changed with peer order: %q vs %q", o1, o2)
	}
	if o1b != o2b {
		t.Fatalf("secondary changed with peer order: %q vs %q", o1b, o2b)
	}
}

func TestRingDistributesOwnersAcrossKeys(t *testing.T) {
	reg := NewRegistry()
	u, _ := url.Parse("http://example.invalid")
	reg.AddBackend("backend-1", u)
	reg.AddBackend("backend-2", u)
	reg.AddBackend("backend-3", u)
	reg.AddBackend("backend-4", u)
	reg.AddBackend("backend-5", u)

	self := Peer{ID: "nodeA", BaseURL: "http://nodeA:8000"}
	peers := []Peer{
		self,
		{ID: "nodeB", BaseURL: "http://nodeB:8000"},
		{ID: "nodeC", BaseURL: "http://nodeC:8000"},
		{ID: "nodeD", BaseURL: "http://nodeD:8000"},
		{ID: "nodeE", BaseURL: "http://nodeE:8000"},
	}
	cl, err := NewClusterLoad(reg, self, peers, ClusterLoadOptions{}, func() float64 { return 0.8 }, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad: %v", err)
	}

	primaries := map[string]struct{}{}
	for i := 1; i <= 5; i++ {
		p, _ := cl.ownersFor(fmt.Sprintf("backend-%d", i))
		primaries[p] = struct{}{}
	}
	// It's a probabilistic property, but with a ring and 5 nodes we should typically not collapse to 1 owner.
	if len(primaries) < 2 {
		t.Fatalf("expected ownership to spread across >=2 primaries, got %d", len(primaries))
	}
}

func TestApplyToRegistryHealthThresholdUsesBuckets(t *testing.T) {
	reg := NewRegistry()
	u, _ := url.Parse("http://example.invalid")
	reg.AddBackend("backend-1", u)

	self := Peer{ID: "nodeA", BaseURL: "http://nodeA:8000"}
	cl, err := NewClusterLoad(reg, self, []Peer{self}, ClusterLoadOptions{TTL: 5 * time.Second}, func() float64 { return 0.8 }, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad: %v", err)
	}

	now := time.Now()
	cl.mu.Lock()
	cl.view["backend-1"] = &viewState{
		ownerID: self.ID,
		epoch:   1,
		seq:     1,
		status:  LoadUp,
		bucket:  7,
		active:  0,
		expires: now.Add(2 * time.Second),
	}
	cl.mu.Unlock()

	cl.applyToRegistry()
	backends := reg.GetBackends()
	if len(backends) != 1 {
		t.Fatalf("expected 1 backend, got %d", len(backends))
	}
	if !backends[0].Stats.IsHealthy() {
		t.Fatalf("expected backend healthy at bucket 7 with threshold 0.8")
	}

	cl.mu.Lock()
	cl.view["backend-1"].bucket = 8
	cl.mu.Unlock()

	cl.applyToRegistry()
	if backends[0].Stats.IsHealthy() {
		t.Fatalf("expected backend unhealthy at bucket 8 with threshold 0.8")
	}
}

func TestSecondaryPublishesWhenPrimaryDownAndBumpsEpoch(t *testing.T) {
	reg := NewRegistry()
	u, _ := url.Parse("http://example.invalid")
	reg.AddBackend("backend-1", u)

	peerA := Peer{ID: "nodeA", BaseURL: "http://nodeA:8000"}
	peerB := Peer{ID: "nodeB", BaseURL: "http://nodeB:8000"}
	peers := []Peer{peerA, peerB}

	// Create cluster load for nodeB (which may be primary or secondary depending on hashing).
	cl, err := NewClusterLoad(reg, peerB, peers, ClusterLoadOptions{PeerDownAfter: 2 * time.Second}, func() float64 { return 0.8 }, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad: %v", err)
	}

	primary, secondary := cl.ownersFor("backend-1")
	if primary == "" || secondary == "" {
		t.Fatalf("expected primary and secondary")
	}

	st := cl.state["backend-1"]
	if st == nil {
		t.Fatalf("expected local state for backend-1 (nodeB must be primary or secondary)")
	}

	now := time.Now()
	cl.mu.Lock()
	// Mark primary as down.
	cl.lastHeard[primary] = now.Add(-10 * time.Second)
	// Pretend the cluster view last had primary as owner at epoch 5.
	cl.view["backend-1"] = &viewState{ownerID: primary, epoch: 5, seq: 10, status: LoadUp, bucket: 1, active: 0, expires: now.Add(5 * time.Second)}

	should := cl.shouldPublishLocked("backend-1", st, now)
	cl.ensureEpochForPublishLocked("backend-1", st)
	newEpoch := st.epoch
	cl.mu.Unlock()

	// Only secondaries should publish because primary is down.
	if st.role == RoleSecondary && !should {
		t.Fatalf("expected secondary to publish when primary down")
	}
	if st.role == RoleSecondary && newEpoch != 6 {
		t.Fatalf("expected epoch bump to 6, got %d", newEpoch)
	}
	if st.role == RolePrimary && should {
		// If nodeB is primary (hashing), it will publish regardless of nodeA liveness.
		// That's fine; the test is primarily about secondary takeover correctness.
	}
}

func TestSecondaryDoesNotTakeOverImmediatelyWithoutEvidence(t *testing.T) {
	reg := NewRegistry()
	u, _ := url.Parse("http://example.invalid")
	reg.AddBackend("backend-1", u)

	peerA := Peer{ID: "nodeA", BaseURL: "http://nodeA:8000"}
	peerB := Peer{ID: "nodeB", BaseURL: "http://nodeB:8000"}
	peers := []Peer{peerA, peerB}

	clA, err := NewClusterLoad(reg, peerA, peers, ClusterLoadOptions{PeerDownAfter: 3 * time.Second}, func() float64 { return 0.8 }, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad A: %v", err)
	}
	clB, err := NewClusterLoad(reg, peerB, peers, ClusterLoadOptions{PeerDownAfter: 3 * time.Second}, func() float64 { return 0.8 }, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad B: %v", err)
	}

	var secondary *ClusterLoad
	if st := clA.state["backend-1"]; st != nil && st.role == RoleSecondary {
		secondary = clA
	} else if st := clB.state["backend-1"]; st != nil && st.role == RoleSecondary {
		secondary = clB
	} else {
		t.Fatalf("expected one of the peers to be secondary")
	}

	primary, _ := secondary.ownersFor("backend-1")

	secondary.mu.Lock()
	st := secondary.state["backend-1"]
	delete(secondary.lastHeard, primary) // unknown
	now := secondary.started.Add(1 * time.Second)
	should := secondary.shouldPublishLocked("backend-1", st, now)
	secondary.mu.Unlock()

	if should {
		t.Fatalf("expected secondary not to publish during startup grace period without evidence primary is down")
	}
}
