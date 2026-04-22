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

	cl1, err := NewClusterLoad(reg, self, peers1, ClusterLoadOptions{}, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad 1: %v", err)
	}
	cl2, err := NewClusterLoad(reg, self, peers2, ClusterLoadOptions{}, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad 2: %v", err)
	}

	o1 := cl1.ownerFor("backend-1")
	o2 := cl2.ownerFor("backend-1")
	if o1 != o2 {
		t.Fatalf("owner changed with peer order: %q vs %q", o1, o2)
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
	cl, err := NewClusterLoad(reg, self, peers, ClusterLoadOptions{}, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad: %v", err)
	}

	primaries := map[string]struct{}{}
	for i := 1; i <= 5; i++ {
		p := cl.ownerFor(fmt.Sprintf("backend-%d", i))
		primaries[p] = struct{}{}
	}
	// It's a probabilistic property, but with a ring and 5 nodes we should typically not collapse to 1 owner.
	if len(primaries) < 2 {
		t.Fatalf("expected ownership to spread across >=2 primaries, got %d", len(primaries))
	}
}

func TestApplyToRegistryUsesFreshUpStatus(t *testing.T) {
	reg := NewRegistry()
	u, _ := url.Parse("http://example.invalid")
	reg.AddBackend("backend-1", u)

	self := Peer{ID: "nodeA", BaseURL: "http://nodeA:8000"}
	cl, err := NewClusterLoad(reg, self, []Peer{self}, ClusterLoadOptions{TTL: 5 * time.Second}, func() time.Duration { return time.Second })
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
		t.Fatalf("expected backend healthy when status is up and fresh")
	}

	cl.mu.Lock()
	cl.view["backend-1"].status = LoadDown
	cl.mu.Unlock()

	cl.applyToRegistry()
	if backends[0].Stats.IsHealthy() {
		t.Fatalf("expected backend unhealthy when status is down")
	}
}

func TestOnlyOwnerTracksAndPublishes(t *testing.T) {
	reg := NewRegistry()
	u, _ := url.Parse("http://example.invalid")
	reg.AddBackend("backend-1", u)

	peerA := Peer{ID: "nodeA", BaseURL: "http://nodeA:8000"}
	peerB := Peer{ID: "nodeB", BaseURL: "http://nodeB:8000"}
	peers := []Peer{peerA, peerB}

	clA, err := NewClusterLoad(reg, peerA, peers, ClusterLoadOptions{PeerDownAfter: 2 * time.Second}, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad A: %v", err)
	}
	clB, err := NewClusterLoad(reg, peerB, peers, ClusterLoadOptions{PeerDownAfter: 2 * time.Second}, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad B: %v", err)
	}

	ownerID := clA.ownerFor("backend-1")
	if ownerID == "" {
		t.Fatalf("expected owner for backend-1")
	}

	ownerCL := clA
	nonOwnerCL := clB
	if ownerID == peerB.ID {
		ownerCL = clB
		nonOwnerCL = clA
	}

	if ownerCL.state["backend-1"] == nil {
		t.Fatalf("expected owner to track backend-1")
	}
	if nonOwnerCL.state["backend-1"] != nil {
		t.Fatalf("expected non-owner not to track backend-1")
	}

	now := time.Now()
	ownerCL.mu.Lock()
	st := ownerCL.state["backend-1"]
	ownerCL.view["backend-1"] = &viewState{ownerID: "other-owner", epoch: 5, seq: 10, status: LoadUp, bucket: 1, active: 0, expires: now.Add(5 * time.Second)}

	should := ownerCL.shouldPublishLocked("backend-1", st, now)
	ownerCL.ensureEpochForPublishLocked("backend-1", st)
	newEpoch := st.epoch
	ownerCL.mu.Unlock()

	if !should {
		t.Fatalf("expected owner to publish")
	}
	if newEpoch != 6 {
		t.Fatalf("expected epoch bump to 6, got %d", newEpoch)
	}
}
