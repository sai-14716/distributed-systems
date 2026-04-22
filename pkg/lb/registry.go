package lb

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	return &Registry{
		Backends: make([]*Backend, 0),
		StickyRules: []StickyRule{
			// Example: /chat requests from any role go to backend-1 or backend-2
			{PathPrefix: "/chat", Role: "", Subset: []string{"backend-1", "backend-2"}},
			// /payload goes to backend-3 (heavier compute)
			{PathPrefix: "/payload", Role: "", Subset: []string{"backend-3"}},
		},
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
	go func() {
		client := &http.Client{Timeout: 500 * time.Millisecond}
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()

		for range t.C {
			backends := r.GetBackends()

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
