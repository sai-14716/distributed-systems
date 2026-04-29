package lb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"
)

type LoadStatus string

const (
	LoadUp   LoadStatus = "up"
	LoadDown LoadStatus = "down"
)

type Peer struct {
	ID      string // stable ID used for ownership hashing (e.g. node1)
	BaseURL string // e.g. http://node1:8000
}

type LoadUpdate struct {
	BackendID      string     `json:"backend_id"`
	OwnerID        string     `json:"owner_id"`
	Epoch          uint64     `json:"epoch"`
	Seq            uint64     `json:"seq"`
	Status         LoadStatus `json:"status"`
	CPUBucket10    int        `json:"cpu_bucket"` // 0..20 (5% buckets)
	ActiveRequests int32      `json:"active_requests"`
}

type GossipMessage struct {
	From    string       `json:"from"`
	Updates []LoadUpdate `json:"updates"`
}

type ClusterLoadOptions struct {
	// TTL bounds how long a load entry is considered fresh after receipt.
	TTL time.Duration

	// RefreshEvery ensures a stable backend remains "fresh" cluster-wide.
	// Even if bucket/status don't change, owners will re-send within this period.
	RefreshEvery time.Duration

	// GossipEvery controls how often we attempt to send deltas to every peer.
	GossipEvery time.Duration

	// ProbeTimeout caps each backend poll.
	ProbeTimeout time.Duration

	// GossipTimeout caps each LB-to-LB gossip send.
	GossipTimeout time.Duration

	// BackendLoadPath is appended to each backend URL when polling.
	BackendLoadPath string

	// PeerCheckEvery controls how often we probe other LBs for liveness.
	PeerCheckEvery time.Duration

	// PeerDownAfter is the local duration after which a peer is treated as down if unheard from.
	PeerDownAfter time.Duration

	// ProbeAssignments maps LB node IDs to the backend IDs they should probe.
	// When empty, the embedded controller_probe_config.json is used.
	ProbeAssignments map[string][]string
}

func (o *ClusterLoadOptions) withDefaults() ClusterLoadOptions {
	out := *o
	if out.TTL <= 0 {
		out.TTL = 5 * time.Second
	}
	if out.RefreshEvery <= 0 {
		out.RefreshEvery = 5 * time.Second
	}
	if out.GossipEvery <= 0 {
		out.GossipEvery = 1 * time.Second
	}
	if out.ProbeTimeout <= 0 {
		out.ProbeTimeout = 500 * time.Millisecond
	}
	if out.GossipTimeout <= 0 {
		out.GossipTimeout = 400 * time.Millisecond
	}
	if out.BackendLoadPath == "" {
		out.BackendLoadPath = "/internal/load"
	}
	if out.PeerCheckEvery <= 0 {
		out.PeerCheckEvery = 1 * time.Second
	}
	if out.PeerDownAfter <= 0 {
		out.PeerDownAfter = 3 * time.Second
	}
	return out
}

// ClusterLoad is an all-to-all, delta-based dissemination mechanism:
// - Each backend is "owned" by exactly one LB (rendezvous hashing).
// - The owner polls the backend and broadcasts updates to every other LB.
// - Updates are only sent on 10% bucket/status change, plus periodic refreshes (<= TTL).
type ClusterLoad struct {
	registry         *Registry
	self             Peer
	peers            []Peer // includes self
	opts             ClusterLoadOptions
	started          time.Time
	ring             *hashRing
	ownerByBackend   map[string]string
	probersByBackend map[string]map[string]struct{}

	// getProbeEvery returns how often owners should poll their backends.
	getProbeEvery func() time.Duration

	httpProbe  *http.Client
	httpGossip *http.Client

	mu sync.Mutex

	// state includes only backends owned by this LB.
	state map[string]*backendState // backend ID -> state

	// reports keeps the latest known report per backend and prober.
	reports map[string]map[string]*viewState // backend ID -> prober ID -> state

	// view is the aggregated cluster view used for routing.
	view map[string]*viewState // backend ID -> aggregated state

	// lastSent tracks per-peer what we've last successfully sent, so we only send deltas.
	lastSent map[string]map[string]sentState // peerID -> backendID -> state

	// lastHeard tracks peer liveness using local time, updated by gossip and health checks.
	lastHeard map[string]time.Time // peerID -> last heard time

	// lastRecv summarizes last gossip receipt per peer for debugging/visualization.
	lastRecv map[string]recvSummary // peerID -> summary

	stop chan struct{}
}

type recvSummary struct {
	At      time.Time `json:"at"`
	Updates int       `json:"updates"`
	Applied int       `json:"applied"`
}

type backendState struct {
	backend *Backend

	// Latest observed probe result.
	lastBucket int
	lastActive int32
	lastStatus LoadStatus

	lastProbe  time.Time
	probeTotal uint64
	probeFail  uint64

	// Last published values for change detection (bucket/status is the "significant" part).
	pubBucket int
	pubActive int32
	pubStatus LoadStatus

	epoch uint64
	seq   uint64

	// For change detection (bucket/status) and for refresh.
	lastPublished time.Time
}

type sentState struct {
	lastSent   time.Time
	ownerID    string
	lastEpoch  uint64
	lastBucket int
	lastStatus LoadStatus
	lastSeq    uint64
}

type viewState struct {
	ownerID string
	epoch   uint64
	seq     uint64
	status  LoadStatus
	bucket  int
	active  int32
	expires time.Time
}

func NewClusterLoad(registry *Registry, self Peer, peers []Peer, opts ClusterLoadOptions, getProbeEvery func() time.Duration) (*ClusterLoad, error) {
	opts = opts.withDefaults()
	if self.ID == "" || self.BaseURL == "" {
		return nil, fmt.Errorf("cluster load requires self peer ID and BaseURL")
	}
	if len(peers) == 0 {
		peers = []Peer{self}
	}

	// Normalize peers: ensure self is present exactly once and IDs are unique.
	byID := map[string]Peer{}
	for _, p := range peers {
		if p.ID == "" || p.BaseURL == "" {
			continue
		}
		byID[p.ID] = p
	}
	byID[self.ID] = self
	norm := make([]Peer, 0, len(byID))
	for _, p := range byID {
		norm = append(norm, p)
	}
	sort.Slice(norm, func(i, j int) bool { return norm[i].ID < norm[j].ID })

	ring, err := newHashRing(norm, 50)
	if err != nil {
		return nil, err
	}
	ownerByBackend, probersByBackend := probeMapsFromAssignments(opts.ProbeAssignments, norm)
	if len(ownerByBackend) == 0 && opts.ProbeAssignments == nil {
		ownerByBackend, probersByBackend = probeMapsFromAssignments(DefaultControllerProbeAssignments(), norm)
	}

	cl := &ClusterLoad{
		registry:         registry,
		self:             self,
		peers:            norm,
		opts:             opts,
		started:          time.Now(),
		ring:             ring,
		ownerByBackend:   ownerByBackend,
		probersByBackend: probersByBackend,
		getProbeEvery:    getProbeEvery,
		httpProbe:        &http.Client{Timeout: opts.ProbeTimeout},
		httpGossip:       &http.Client{Timeout: opts.GossipTimeout},
		state:            map[string]*backendState{},
		reports:          map[string]map[string]*viewState{},
		view:             map[string]*viewState{},
		lastSent:         map[string]map[string]sentState{},
		lastHeard:        map[string]time.Time{},
		lastRecv:         map[string]recvSummary{},
		stop:             make(chan struct{}),
	}

	for _, p := range cl.peers {
		if p.ID == cl.self.ID {
			continue
		}
		cl.lastSent[p.ID] = map[string]sentState{}
	}

	cl.initOwnership()
	return cl, nil
}

func (c *ClusterLoad) initOwnership() {
	now := time.Now()
	backends := c.registry.GetBackends()
	for _, b := range backends {
		if !c.shouldProbeBackend(b.ID, c.self.ID) {
			continue
		}

		c.state[b.ID] = &backendState{
			backend:       b,
			lastBucket:    0,
			lastActive:    0,
			lastStatus:    LoadUp,
			lastProbe:     time.Time{},
			probeTotal:    0,
			probeFail:     0,
			pubBucket:     -1,
			pubActive:     0,
			pubStatus:     "",
			epoch:         1,
			seq:           1,
			lastPublished: time.Time{},
		}
		c.setReportLocked(b.ID, c.self.ID, &viewState{ownerID: c.self.ID, epoch: 0, seq: 0, status: LoadDown, bucket: 10, active: 0, expires: now.Add(-time.Second)})
		c.view[b.ID] = c.aggregateBackendLocked(b.ID, now)
	}
	c.lastHeard[c.self.ID] = now
}

func (c *ClusterLoad) Start() {
	// Probe loop per locally-owned backend.
	for backendID := range c.state {
		id := backendID
		go c.runProbeLoop(id)
	}
	go c.runPeerHealthLoop()
	go c.runGossipLoop()
	go c.runRegistryApplyLoop()
}

func (c *ClusterLoad) Stop() {
	close(c.stop)
}

func (c *ClusterLoad) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "id": c.self.ID})
}

type LoadViewSnapshot struct {
	Self      string                 `json:"self"`
	Peers     []Peer                 `json:"peers"`
	LastHeard map[string]time.Time   `json:"last_heard"`
	LastRecv  map[string]recvSummary `json:"last_recv"`
	Backends  map[string]any         `json:"backends"`
	Now       time.Time              `json:"now"`
}

// Snapshot returns a view of what this LB currently believes about cluster load.
// Intended for visualization/debugging (not used on the data path).
func (c *ClusterLoad) Snapshot() LoadViewSnapshot {
	now := time.Now()

	backendIDs := make([]string, 0)
	for _, b := range c.registry.GetBackends() {
		if b != nil {
			backendIDs = append(backendIDs, b.ID)
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	lh := make(map[string]time.Time, len(c.lastHeard))
	for k, v := range c.lastHeard {
		lh[k] = v
	}

	lr := make(map[string]recvSummary, len(c.lastRecv))
	for k, v := range c.lastRecv {
		lr[k] = v
	}

	backends := make(map[string]any, len(backendIDs))
	for _, id := range backendIDs {
		v := c.aggregateBackendLocked(id, now)
		probers := c.probersForBackend(id)
		entry := map[string]any{
			"owner":   c.ownerFor(id),
			"probers": probers,
			"local_role": func() string {
				if c.shouldProbeBackend(id, c.self.ID) {
					return "prober"
				}
				return "none"
			}(),
		}
		if owner := c.ownerFor(id); owner != "" {
			entry["owner_alive"] = c.isPeerAliveLocked(owner, now)
		}
		if st, ok := c.state[id]; ok && st != nil {
			entry["local_probe"] = map[string]any{
				"at":          st.lastProbe,
				"total":       st.probeTotal,
				"fail":        st.probeFail,
				"last_bucket": st.lastBucket,
				"last_status": st.lastStatus,
			}
			entry["local_publish"] = map[string]any{
				"epoch": st.epoch,
				"seq":   st.seq,
				"at":    st.lastPublished,
			}
		} else {
			entry["local_probe"] = nil
			entry["local_publish"] = nil
		}
		if v != nil {
			entry["view"] = map[string]any{
				"owner_id": v.ownerID,
				"epoch":    v.epoch,
				"seq":      v.seq,
				"status":   v.status,
				"bucket":   v.bucket,
				"active":   v.active,
				"expires":  v.expires,
			}
		} else {
			entry["view"] = nil
		}
		reportViews := map[string]any{}
		for proberID, report := range c.reports[id] {
			if report == nil {
				continue
			}
			reportViews[proberID] = map[string]any{
				"epoch":   report.epoch,
				"seq":     report.seq,
				"status":  report.status,
				"bucket":  report.bucket,
				"active":  report.active,
				"expires": report.expires,
			}
		}
		entry["reports"] = reportViews
		backends[id] = entry
	}

	return LoadViewSnapshot{
		Self:      c.self.ID,
		Peers:     append([]Peer(nil), c.peers...),
		LastHeard: lh,
		LastRecv:  lr,
		Backends:  backends,
		Now:       now,
	}
}

func (c *ClusterLoad) HandleGossip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var msg GossipMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	if msg.From != "" {
		c.lastHeard[msg.From] = now
	}

	applied := 0
	for _, u := range msg.Updates {
		if u.BackendID == "" {
			continue
		}
		if !c.shouldProbeBackend(u.BackendID, u.OwnerID) {
			continue
		}
		if u.CPUBucket10 < 0 {
			u.CPUBucket10 = 0
		}
		if u.CPUBucket10 > 20 {
			u.CPUBucket10 = 20
		}

		cur := c.reportForLocked(u.BackendID, u.OwnerID)
		if cur == nil || u.Epoch > cur.epoch || (u.Epoch == cur.epoch && u.Seq > cur.seq) {
			c.setReportLocked(u.BackendID, u.OwnerID, &viewState{
				ownerID: u.OwnerID,
				epoch:   u.Epoch,
				seq:     u.Seq,
				status:  u.Status,
				bucket:  u.CPUBucket10,
				active:  u.ActiveRequests,
				expires: now.Add(c.opts.TTL),
			})
			c.view[u.BackendID] = c.aggregateBackendLocked(u.BackendID, now)
			applied++
			continue
		}
		// Heartbeat/refresh: same seq should still extend TTL (bounded by owner sending).
		if cur.ownerID == u.OwnerID && cur.epoch == u.Epoch && cur.seq == u.Seq {
			cur.expires = now.Add(c.opts.TTL)
			c.view[u.BackendID] = c.aggregateBackendLocked(u.BackendID, now)
		}
	}

	if msg.From != "" {
		c.lastRecv[msg.From] = recvSummary{At: now, Updates: len(msg.Updates), Applied: applied}
	}
	log.Printf("[lb-gossip] event=recv self=%s from=%s updates=%d applied=%d", c.self.ID, msg.From, len(msg.Updates), applied)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (c *ClusterLoad) runPeerHealthLoop() {
	t := time.NewTicker(c.opts.PeerCheckEvery)
	defer t.Stop()

	for {
		select {
		case <-c.stop:
			return
		case <-t.C:
			for _, p := range c.peers {
				if p.ID == c.self.ID {
					continue
				}
				base, err := url.Parse(p.BaseURL)
				if err != nil {
					continue
				}
				base.Path = "/internal/lb/health"
				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, base.String(), nil)
				if err != nil {
					continue
				}
				resp, err := c.httpGossip.Do(req)
				if err != nil {
					continue
				}
				_ = resp.Body.Close()
				if resp.StatusCode/100 != 2 {
					continue
				}
				now := time.Now()
				c.mu.Lock()
				c.lastHeard[p.ID] = now
				c.mu.Unlock()
			}
		}
	}
}

func (c *ClusterLoad) isPeerAliveLocked(peerID string, now time.Time) bool {
	if peerID == "" {
		return false
	}
	if peerID == c.self.ID {
		return true
	}
	t, ok := c.lastHeard[peerID]
	if !ok || t.IsZero() {
		// Avoid immediate false-positive takeovers at startup: assume peers are alive until we
		// either hear from them or a short grace period passes.
		return now.Sub(c.started) <= c.opts.PeerDownAfter
	}
	return now.Sub(t) <= c.opts.PeerDownAfter
}

func (c *ClusterLoad) shouldPublishLocked(backendID string, st *backendState, now time.Time) bool {
	_ = st
	_ = now
	return c.shouldProbeBackend(backendID, c.self.ID)
}

func (c *ClusterLoad) ensureEpochForPublishLocked(backendID string, st *backendState) {
	_ = backendID
	_ = st
}

func (c *ClusterLoad) runProbeLoop(backendID string) {
	for {
		select {
		case <-c.stop:
			return
		default:
		}

		probeEvery := 1 * time.Second
		if c.getProbeEvery != nil {
			if d := c.getProbeEvery(); d > 0 {
				probeEvery = d
			}
		}

		c.probeOnce(backendID)
		time.Sleep(probeEvery)
	}
}

type backendLoadResp struct {
	OK             bool    `json:"ok"`
	CPUPct         float64 `json:"cpu_pct"`
	CPUBucket      int     `json:"cpu_bucket"`
	ActiveRequests int32   `json:"active_requests"`
}

func (c *ClusterLoad) probeOnce(backendID string) {
	c.mu.Lock()
	st := c.state[backendID]
	c.mu.Unlock()
	if st == nil || st.backend == nil || st.backend.URL == nil {
		return
	}

	u := *st.backend.URL
	u.Path = c.opts.BackendLoadPath
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), nil)
	if err != nil {
		return
	}
	resp, err := c.httpProbe.Do(req)
	if err != nil {
		c.recordProbe(backendID, 20, 0, LoadDown, false)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		c.recordProbe(backendID, 20, 0, LoadDown, false)
		return
	}
	var lr backendLoadResp
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		c.recordProbe(backendID, 20, 0, LoadDown, false)
		return
	}

	bucket := lr.CPUBucket
	if bucket < 0 {
		bucket = 0
	}
	if bucket > 10 {
		bucket = 10
	}
	status := LoadUp
	if !lr.OK {
		status = LoadDown
	}
	c.recordProbe(backendID, bucket, lr.ActiveRequests, status, true)
}

func (c *ClusterLoad) recordProbe(backendID string, bucket int, active int32, status LoadStatus, probeOK bool) {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	st := c.state[backendID]
	if st == nil {
		return
	}

	st.lastProbe = now
	st.probeTotal++
	if !probeOK {
		st.probeFail++
	}

	changed := bucket != st.lastBucket || status != st.lastStatus
	st.lastBucket = bucket
	st.lastActive = active
	st.lastStatus = status

	// Each configured prober contributes one report; the aggregate view is used for routing.
	if c.shouldPublishLocked(backendID, st, now) {
		c.ensureEpochForPublishLocked(backendID, st)
		c.setReportLocked(backendID, c.self.ID, &viewState{
			ownerID: c.self.ID,
			epoch:   st.epoch,
			seq:     st.seq,
			status:  status,
			bucket:  bucket,
			active:  active,
			expires: now.Add(c.opts.TTL),
		})
		c.view[backendID] = c.aggregateBackendLocked(backendID, now)
	}
	if changed {
		log.Printf("[lb-load] event=probe self=%s backend=%s bucket=%d status=%s active=%d", c.self.ID, backendID, bucket, status, active)
	}
}

func (c *ClusterLoad) runGossipLoop() {
	t := time.NewTicker(c.opts.GossipEvery)
	defer t.Stop()

	for {
		select {
		case <-c.stop:
			return
		case <-t.C:
			c.gossipOnce()
		}
	}
}

func (c *ClusterLoad) gossipOnce() {
	type pubSnap struct {
		backendID string
		ownerID   string
		epoch     uint64
		seq       uint64
		bucket    int
		status    LoadStatus
		active    int32
	}

	now := time.Now()

	c.mu.Lock()
	published := make([]pubSnap, 0, len(c.state))
	for id, st := range c.state {
		if !c.shouldPublishLocked(id, st, now) {
			continue
		}
		c.ensureEpochForPublishLocked(id, st)

		if st.pubBucket != st.lastBucket || st.pubStatus != st.lastStatus {
			st.seq++
			st.pubBucket = st.lastBucket
			st.pubStatus = st.lastStatus
			st.pubActive = st.lastActive
			st.lastPublished = now
			log.Printf("[lb-load] event=publish self=%s backend=%s epoch=%d seq=%d bucket=%d status=%s", c.self.ID, id, st.epoch, st.seq, st.lastBucket, st.lastStatus)
		} else if st.lastPublished.IsZero() {
			// Ensure refresh scheduling works even if we haven't observed a "change".
			st.lastPublished = now
		}

		published = append(published, pubSnap{
			backendID: id,
			ownerID:   c.self.ID,
			epoch:     st.epoch,
			seq:       st.seq,
			bucket:    st.lastBucket,
			status:    st.lastStatus,
			active:    st.lastActive,
		})
	}

	for id, byProber := range c.reports {
		for proberID, v := range byProber {
			if v == nil || proberID == c.self.ID || !v.expires.After(now) {
				continue
			}
			if !c.shouldProbeBackend(id, proberID) {
				continue
			}
			published = append(published, pubSnap{
				backendID: id,
				ownerID:   proberID,
				epoch:     v.epoch,
				seq:       v.seq,
				bucket:    v.bucket,
				status:    v.status,
				active:    v.active,
			})
		}
	}

	peers := make([]Peer, 0, len(c.peers))
	for _, p := range c.peers {
		if p.ID != c.self.ID {
			peers = append(peers, p)
		}
	}
	c.mu.Unlock()

	for _, p := range peers {
		updates := make([]LoadUpdate, 0, 16)

		c.mu.Lock()
		sentMap := c.lastSent[p.ID]
		if sentMap == nil {
			sentMap = map[string]sentState{}
			c.lastSent[p.ID] = sentMap
		}
		for _, st := range published {
			key := sentKey(st.backendID, st.ownerID)
			prev := sentMap[key]
			needsDelta := prev.lastSeq == 0 ||
				prev.ownerID != st.ownerID ||
				prev.lastEpoch != st.epoch ||
				prev.lastBucket != st.bucket ||
				prev.lastStatus != st.status ||
				(st.epoch == prev.lastEpoch && st.seq > prev.lastSeq)
			needsRefresh := prev.lastSeq == 0 || now.Sub(prev.lastSent) >= c.opts.RefreshEvery
			if !needsDelta && !needsRefresh {
				continue
			}
			updates = append(updates, LoadUpdate{
				BackendID:      st.backendID,
				OwnerID:        st.ownerID,
				Epoch:          st.epoch,
				Seq:            st.seq,
				Status:         st.status,
				CPUBucket10:    st.bucket,
				ActiveRequests: st.active,
			})
		}
		c.mu.Unlock()

		if len(updates) == 0 {
			continue
		}

		if err := c.sendGossip(p, updates); err != nil {
			log.Printf("[lb-gossip] event=send self=%s peer=%s updates=%d err=%v", c.self.ID, p.ID, len(updates), err)
			continue
		}
		log.Printf("[lb-gossip] event=send self=%s peer=%s updates=%d ok=true", c.self.ID, p.ID, len(updates))

		// Mark sent.
		c.mu.Lock()
		sentMap = c.lastSent[p.ID]
		for _, u := range updates {
			sentMap[sentKey(u.BackendID, u.OwnerID)] = sentState{
				lastSent:   now,
				ownerID:    u.OwnerID,
				lastEpoch:  u.Epoch,
				lastBucket: u.CPUBucket10,
				lastStatus: u.Status,
				lastSeq:    u.Seq,
			}
		}
		c.mu.Unlock()
	}
}

func (c *ClusterLoad) sendGossip(peer Peer, updates []LoadUpdate) error {
	base, err := url.Parse(peer.BaseURL)
	if err != nil {
		return err
	}
	base.Path = "/internal/lb/gossip"
	msg := GossipMessage{From: c.self.ID, Updates: updates}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, base.String(), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpGossip.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("status=%d", resp.StatusCode)
	}
	return nil
}

func (c *ClusterLoad) runRegistryApplyLoop() {
	t := time.NewTicker(250 * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-c.stop:
			return
		case <-t.C:
			c.applyToRegistry()
		}
	}
}

func (c *ClusterLoad) applyToRegistry() {
	now := time.Now()

	backends := c.registry.GetBackends()
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, b := range backends {
		st := c.aggregateBackendLocked(b.ID, now)
		c.view[b.ID] = st
		healthy := false
		bucket := 10

		if st != nil {
			bucket = st.bucket
			if st.expires.After(now) && st.status == LoadUp {
				healthy = true
			}
		}

		b.Stats.SetHealthy(healthy)
		b.Stats.SetCPUBucket(bucket)

		// Weight is used by WRR; keep it monotone w.r.t. load.
		w := 1.0 - float64(bucket)*0.1
		if w < 0.1 {
			w = 0.1
		}
		b.Stats.SetWeight(w)
	}
}

func (c *ClusterLoad) ownerFor(backendID string) string {
	if c != nil {
		if owner := c.ownerByBackend[backendID]; owner != "" {
			return owner
		}
	}
	if c == nil || c.ring == nil {
		return ""
	}
	return c.ring.owner(backendID)
}

func (c *ClusterLoad) shouldProbeBackend(backendID, peerID string) bool {
	if c == nil || backendID == "" || peerID == "" {
		return false
	}
	if probers := c.probersByBackend[backendID]; len(probers) > 0 {
		_, ok := probers[peerID]
		return ok
	}
	return c.ownerFor(backendID) == peerID
}

func (c *ClusterLoad) probersForBackend(backendID string) []string {
	if c == nil {
		return nil
	}
	if probers := c.probersByBackend[backendID]; len(probers) > 0 {
		out := make([]string, 0, len(probers))
		for id := range probers {
			out = append(out, id)
		}
		sort.Strings(out)
		return out
	}
	if owner := c.ownerFor(backendID); owner != "" {
		return []string{owner}
	}
	return nil
}

func (c *ClusterLoad) reportForLocked(backendID, proberID string) *viewState {
	if c == nil || c.reports == nil {
		return nil
	}
	return c.reports[backendID][proberID]
}

func (c *ClusterLoad) setReportLocked(backendID, proberID string, report *viewState) {
	if c.reports[backendID] == nil {
		c.reports[backendID] = map[string]*viewState{}
	}
	c.reports[backendID][proberID] = report
}

func (c *ClusterLoad) aggregateBackendLocked(backendID string, now time.Time) *viewState {
	reports := c.reports[backendID]
	if len(reports) == 0 {
		return nil
	}

	var bestUp *viewState
	var bestDown *viewState
	for proberID, report := range reports {
		if report == nil || !report.expires.After(now) || !c.isPeerAliveLocked(proberID, now) {
			continue
		}
		if report.status == LoadUp {
			if bestUp == nil || report.bucket < bestUp.bucket || (report.bucket == bestUp.bucket && report.expires.After(bestUp.expires)) {
				bestUp = report
			}
			continue
		}
		if bestDown == nil || report.expires.After(bestDown.expires) {
			bestDown = report
		}
	}
	if bestUp != nil {
		return cloneViewState(bestUp)
	}
	if bestDown != nil {
		return cloneViewState(bestDown)
	}
	return nil
}

func cloneViewState(v *viewState) *viewState {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func sentKey(backendID, ownerID string) string {
	return backendID + "\x00" + ownerID
}

func probeMapsFromAssignments(assignments map[string][]string, peers []Peer) (map[string]string, map[string]map[string]struct{}) {
	if len(assignments) == 0 {
		return nil, nil
	}

	peerIDs := make(map[string]struct{}, len(peers))
	for _, p := range peers {
		if p.ID != "" {
			peerIDs[p.ID] = struct{}{}
		}
	}

	probersByBackend := map[string]map[string]struct{}{}
	for ownerID, backendIDs := range assignments {
		if _, ok := peerIDs[ownerID]; !ok {
			continue
		}
		for _, backendID := range backendIDs {
			if backendID == "" {
				continue
			}
			if probersByBackend[backendID] == nil {
				probersByBackend[backendID] = map[string]struct{}{}
			}
			probersByBackend[backendID][ownerID] = struct{}{}
		}
	}
	if len(probersByBackend) == 0 {
		return nil, nil
	}

	primaryByBackend := map[string]string{}
	for backendID, probers := range probersByBackend {
		ids := make([]string, 0, len(probers))
		for id := range probers {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		primaryByBackend[backendID] = ids[0]
	}
	return primaryByBackend, probersByBackend
}
