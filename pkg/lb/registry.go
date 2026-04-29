package lb

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ---------- Health / Stats ----------

type BackendStats struct {
	ActiveRequests int32
	TotalRequests  int64

	// These fields are read concurrently by the dataplane selection algorithms.
	healthy    uint32 // 0/1
	weightBits uint64 // float64 bits
	cpuBucket  int32  // 0..20 (5% buckets)
}

func (s *BackendStats) SetHealthy(v bool) {
	if v {
		atomic.StoreUint32(&s.healthy, 1)
	} else {
		atomic.StoreUint32(&s.healthy, 0)
	}
}

func (s *BackendStats) IsHealthy() bool {
	return atomic.LoadUint32(&s.healthy) == 1
}

func (s *BackendStats) SetWeight(w float64) {
	atomic.StoreUint64(&s.weightBits, math.Float64bits(w))
}

func (s *BackendStats) Weight() float64 {
	return math.Float64frombits(atomic.LoadUint64(&s.weightBits))
}

func (s *BackendStats) SetCPUBucket(b int) {
	atomic.StoreInt32(&s.cpuBucket, int32(b))
}

func (s *BackendStats) CPUBucket() int {
	return int(atomic.LoadInt32(&s.cpuBucket))
}

// ---------- Backend ----------

type Backend struct {
	ID    string
	URL   *url.URL
	Proxy *httputil.ReverseProxy
	Stats BackendStats
}

// ---------- Sticky Rules ----------

// StickyRule narrows which backend subset handles a matching request.
// PathPrefix="" or Role="" act as wildcards.
// NOTE: Replication of StickyRules across LBs is owned by the Raft/Controller team.
type StickyRule struct {
	PathPrefix string
	Role       string
	Subset     []string // backend IDs
}

// ---------- Registry (Shared Memory) ----------

// Registry is the shared memory between the Controller and Forwarder.
type Registry struct {
	Backends    []*Backend
	Epoch       uint64
	StickyRules []StickyRule
	mu          sync.RWMutex
}

func NewRegistry() *Registry {
	endpoints := []string{"/chat", "/payload", "/encrypt", "/"}
	allBackends := []string{
		"backend-1", "backend-2", "backend-3", "backend-4", "backend-5",
		"backend-6", "backend-7", "backend-8", "backend-9", "backend-10",
	}

	var rules []StickyRule
	for _, ep := range endpoints {
		shuffled := append([]string{}, allBackends...)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		// Assign a random subset of 3 to 7 backends to each endpoint
		subsetSize := rand.Intn(5) + 3
		rules = append(rules, StickyRule{
			PathPrefix: ep,
			Role:       "",
			Subset:     shuffled[:subsetSize],
		})
	}

	return &Registry{
		Backends:    make([]*Backend, 0),
		StickyRules: rules,
	}
}

func (r *Registry) AddBackend(id string, u *url.URL) {
	r.mu.Lock()
	defer r.mu.Unlock()
	proxy := httputil.NewSingleHostReverseProxy(u)
	// Forward X-Server-ID from backend through the proxy to the client
	proxy.ModifyResponse = func(resp *http.Response) error {
		if v := resp.Header.Get("X-Server-ID"); v != "" {
			resp.Header.Set("X-Server-ID", v)
		}
		nodeID := resp.Request.Header.Get("X-LB-Node")
		if p := resp.Header.Get("X-Packet-Path"); p != "" && nodeID != "" {
			resp.Header.Set("X-Packet-Path", p+" -> LB("+nodeID+")")
		}
		return nil
	}
	b := &Backend{
		ID:    id,
		URL:   u,
		Proxy: proxy,
		Stats: BackendStats{},
	}
	b.Stats.SetHealthy(true)
	b.Stats.SetWeight(1.0)
	b.Stats.SetCPUBucket(0)
	r.Backends = append(r.Backends, b)
	atomic.AddUint64(&r.Epoch, 1)
}

func (r *Registry) GetHealthyBackends() []*Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Backend
	for _, b := range r.Backends {
		if b.Stats.IsHealthy() {
			out = append(out, b)
		}
	}
	return out
}

// MatchSubset returns the filtered healthy backends for a matching StickyRule,
// or nil (meaning: use all healthy backends).
func (r *Registry) MatchSubset(path, role string) []*Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rule := range r.StickyRules {
		pathMatch := rule.PathPrefix == "" || len(path) >= len(rule.PathPrefix) && path[:len(rule.PathPrefix)] == rule.PathPrefix
		roleMatch := rule.Role == "" || rule.Role == role
		if pathMatch && roleMatch {
			var subset []*Backend
			for _, b := range r.Backends {
				if b.Stats.IsHealthy() {
					for _, id := range rule.Subset {
						if b.ID == id {
							subset = append(subset, b)
							break
						}
					}
				}
			}
			if len(subset) > 0 {
				return subset
			}
		}
	}
	return nil
}

func (r *Registry) GetEpoch() uint64 {
	return atomic.LoadUint64(&r.Epoch)
}

// SimulateBackendFailure marks a backend unhealthy from the Controller side.
// In production, the Raft/Controller team provides real health signals.
func (r *Registry) SimulateBackendFailure(id string, healthy bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range r.Backends {
		if b.ID == id {
			b.Stats.SetHealthy(healthy)
			log.Printf("[Controller] Backend %s IsHealthy set to %v", id, healthy)
		}
	}
	atomic.AddUint64(&r.Epoch, 1)
}

// StartController mocks the Controller process.
// Updates weights (N(0,1)) and prints shared-memory state every second.
func (r *Registry) StartController() {
	selfID := os.Getenv("NODE_ID")
	assignments := DefaultControllerProbeAssignments()
	r.StartControllerForNode(selfID, assignments)
}

func (r *Registry) StartControllerForNode(selfID string, assignments map[string][]string) {
	go func() {
		client := &http.Client{Timeout: 500 * time.Millisecond}
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()

		assigned := backendSetForNode(selfID, assignments)
		if len(assigned) > 0 {
			log.Printf("[Controller] node=%s probing assigned backend subset=%v", selfID, sortedBackendIDs(assigned))
		} else {
			log.Printf("[Controller] node=%s has no probe subset; legacy controller health probes disabled", selfID)
		}

		for range t.C {
			backends := r.GetBackendsByID(assigned)

			for _, b := range backends {
				var active, total int64
				healthy := false

				resp, err := client.Get(b.URL.String() + "/health")
				if err == nil {
					if resp.StatusCode == 200 {
						var res struct {
							Status         string `json:"status"`
							ActiveRequests int64  `json:"active_requests"`
							TotalRequests  int64  `json:"total_requests"`
						}
						if err := json.NewDecoder(resp.Body).Decode(&res); err == nil {
							healthy = (res.Status == "OK")
							active = res.ActiveRequests
							total = res.TotalRequests
						}
					}
					_ = resp.Body.Close()
				}

				b.Stats.SetHealthy(healthy)
				if healthy {
					atomic.StoreInt32(&b.Stats.ActiveRequests, int32(active))
					atomic.StoreInt64(&b.Stats.TotalRequests, total)
				}

				// Random weight for the baseline (non-gossip) controller mode.
				w := math.Abs(rand.NormFloat64())
				if w < 0.1 {
					w = 0.1
				}
				b.Stats.SetWeight(w)

				log.Printf("[Controller] -> %s Healthy: %v, ActiveReqs: %d, TotalReqs: %d, Weight: %.2f",
					b.ID, b.Stats.IsHealthy(),
					atomic.LoadInt32(&b.Stats.ActiveRequests),
					atomic.LoadInt64(&b.Stats.TotalRequests),
					b.Stats.Weight(),
				)
			}

			atomic.AddUint64(&r.Epoch, 1)
		}
	}()
}

func (r *Registry) GetBackends() []*Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Backend, 0, len(r.Backends))
	out = append(out, r.Backends...)
	return out
}

func (r *Registry) GetBackendsByID(ids map[string]struct{}) []*Backend {
	if len(ids) == 0 {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Backend, 0, len(ids))
	for _, b := range r.Backends {
		if _, ok := ids[b.ID]; ok {
			out = append(out, b)
		}
	}
	return out
}

func backendSetForNode(nodeID string, assignments map[string][]string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, id := range assignments[nodeID] {
		if id != "" {
			out[id] = struct{}{}
		}
	}
	return out
}

func sortedBackendIDs(ids map[string]struct{}) []string {
	out := make([]string, 0, len(ids))
	for id := range ids {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
