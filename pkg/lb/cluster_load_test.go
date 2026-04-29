package lb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestProbeAssignmentsAllowOverlappingProbers(t *testing.T) {
	reg := NewRegistry()
	u, _ := url.Parse("http://example.invalid")
	for i := 1; i <= 10; i++ {
		reg.AddBackend(fmt.Sprintf("backend-%d", i), u)
	}

	peers := []Peer{
		{ID: "node1", BaseURL: "http://node1:8000"},
		{ID: "node2", BaseURL: "http://node2:8000"},
		{ID: "node3", BaseURL: "http://node3:8000"},
		{ID: "node4", BaseURL: "http://node4:8000"},
		{ID: "node5", BaseURL: "http://node5:8000"},
	}
	cl, err := NewClusterLoad(reg, peers[0], peers, ClusterLoadOptions{}, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad: %v", err)
	}

	want := map[string][]string{
		"backend-1":  {"node3", "node4"},
		"backend-2":  {"node1", "node5"},
		"backend-3":  {"node1"},
		"backend-4":  {"node4"},
		"backend-5":  {"node2", "node3"},
		"backend-6":  {"node3"},
		"backend-7":  {"node1", "node2"},
		"backend-8":  {"node2"},
		"backend-9":  {"node4", "node5"},
		"backend-10": {"node5"},
	}
	for backendID, proberIDs := range want {
		got := cl.probersForBackend(backendID)
		if !sameStrings(got, proberIDs) {
			t.Fatalf("probersForBackend(%s)=%v, want %v", backendID, got, proberIDs)
		}
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
	cl.setReportLocked("backend-1", self.ID, &viewState{
		ownerID: self.ID,
		epoch:   1,
		seq:     1,
		status:  LoadUp,
		bucket:  7,
		active:  0,
		expires: now.Add(2 * time.Second),
	})
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
	cl.reports["backend-1"][self.ID].status = LoadDown
	cl.mu.Unlock()

	cl.applyToRegistry()
	if backends[0].Stats.IsHealthy() {
		t.Fatalf("expected backend unhealthy when status is down")
	}
}

func TestOnlyAssignedProberTracksAndPublishes(t *testing.T) {
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

	proberID := clA.ownerFor("backend-1")
	if proberID == "" {
		t.Fatalf("expected prober for backend-1")
	}

	proberCL := clA
	nonProberCL := clB
	if proberID == peerB.ID {
		proberCL = clB
		nonProberCL = clA
	}

	if proberCL.state["backend-1"] == nil {
		t.Fatalf("expected prober to track backend-1")
	}
	if nonProberCL.state["backend-1"] != nil {
		t.Fatalf("expected non-prober not to track backend-1")
	}

	now := time.Now()
	proberCL.mu.Lock()
	st := proberCL.state["backend-1"]

	should := proberCL.shouldPublishLocked("backend-1", st, now)
	newEpoch := st.epoch
	proberCL.mu.Unlock()

	if !should {
		t.Fatalf("expected prober to publish")
	}
	if newEpoch != 1 {
		t.Fatalf("expected prober epoch to remain independent, got %d", newEpoch)
	}
}

func TestHandleGossipAcceptsRelayedOwnerUpdate(t *testing.T) {
	reg := NewRegistry()
	u, _ := url.Parse("http://example.invalid")
	reg.AddBackend("backend-1", u)

	peers := []Peer{
		{ID: "node1", BaseURL: "http://node1:8000"},
		{ID: "node2", BaseURL: "http://node2:8000"},
		{ID: "node3", BaseURL: "http://node3:8000"},
	}
	receiver, err := NewClusterLoad(reg, peers[0], peers, ClusterLoadOptions{TTL: 5 * time.Second}, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad: %v", err)
	}

	ownerID := receiver.ownerFor("backend-1")
	if ownerID == "" || ownerID == peers[0].ID || ownerID == peers[1].ID {
		t.Fatalf("test expected backend-1 owner to be a third peer, got %q", ownerID)
	}

	msg := GossipMessage{
		From: peers[1].ID,
		Updates: []LoadUpdate{{
			BackendID:      "backend-1",
			OwnerID:        ownerID,
			Epoch:          7,
			Seq:            3,
			Status:         LoadUp,
			CPUBucket10:    4,
			ActiveRequests: 2,
		}},
	}
	body, _ := json.Marshal(msg)
	req := httptest.NewRequest(http.MethodPost, "/internal/lb/gossip", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	receiver.HandleGossip(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	v := receiver.view["backend-1"]
	if v == nil {
		t.Fatalf("expected relayed update to populate view")
	}
	if v.ownerID != ownerID || v.epoch != 7 || v.seq != 3 || v.status != LoadUp || v.bucket != 4 || v.active != 2 {
		t.Fatalf("unexpected view after relayed update: %+v", *v)
	}
}

func TestGossipOnceRelaysFreshViewEntries(t *testing.T) {
	reg := NewRegistry()
	u, _ := url.Parse("http://example.invalid")
	reg.AddBackend("backend-1", u)

	receiverPeer := Peer{ID: "node1", BaseURL: ""}
	relayPeer := Peer{ID: "node2", BaseURL: "http://relay.invalid"}
	ownerPeer := Peer{ID: "node3", BaseURL: ""}

	var got GossipMessage
	receiverServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/lb/gossip" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode gossip: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer receiverServer.Close()
	receiverPeer.BaseURL = receiverServer.URL

	ownerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer ownerServer.Close()
	ownerPeer.BaseURL = ownerServer.URL

	peers := []Peer{receiverPeer, relayPeer, ownerPeer}
	relay, err := NewClusterLoad(reg, relayPeer, peers, ClusterLoadOptions{TTL: 5 * time.Second, RefreshEvery: time.Hour}, func() time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("NewClusterLoad: %v", err)
	}

	ownerID := relay.ownerFor("backend-1")
	if ownerID == "" || ownerID == relayPeer.ID {
		t.Fatalf("test expected backend-1 to be owned by another peer, got %q", ownerID)
	}
	relay.mu.Lock()
	relay.setReportLocked("backend-1", ownerID, &viewState{
		ownerID: ownerID,
		epoch:   9,
		seq:     11,
		status:  LoadUp,
		bucket:  3,
		active:  5,
		expires: time.Now().Add(5 * time.Second),
	})
	relay.mu.Unlock()

	relay.gossipOnce()

	if got.From != relayPeer.ID {
		t.Fatalf("expected relay sender %q, got %q", relayPeer.ID, got.From)
	}
	if len(got.Updates) != 1 {
		t.Fatalf("expected 1 relayed update, got %d: %+v", len(got.Updates), got.Updates)
	}
	u0 := got.Updates[0]
	if u0.OwnerID != ownerID || u0.BackendID != "backend-1" || u0.Epoch != 9 || u0.Seq != 11 || u0.CPUBucket10 != 3 || u0.ActiveRequests != 5 {
		t.Fatalf("unexpected relayed update: %+v", u0)
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
