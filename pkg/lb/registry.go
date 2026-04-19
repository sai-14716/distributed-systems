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
	IsHealthy      bool
	Weight         float64
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
// ClientBackendMap provides session stickiness for WRR and LR.
// NOTE: Cross-LB replication of ClientBackendMap is owned by Raft/Controller team.
type Registry struct {
	Backends         []*Backend
	Epoch            uint64
	StickyRules      []StickyRule
	ClientBackendMap map[string]string // session_key → backend ID
	mu               sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		Backends:         make([]*Backend, 0),
		ClientBackendMap: make(map[string]string),
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
		return nil
	}
	b := &Backend{
		ID:  id,
		URL: u,
		Proxy: proxy,
		Stats: BackendStats{IsHealthy: true, Weight: 1.0},
	}
	r.Backends = append(r.Backends, b)
	atomic.AddUint64(&r.Epoch, 1)
}

func (r *Registry) GetHealthyBackends() []*Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Backend
	for _, b := range r.Backends {
		if b.Stats.IsHealthy {
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
				if b.Stats.IsHealthy {
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

// StickyGet returns the previously mapped backend for a session key (WRR/LR).
func (r *Registry) StickyGet(key string) *Backend {
	r.mu.RLock()
	id, ok := r.ClientBackendMap[key]
	r.mu.RUnlock()
	if !ok {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.Backends {
		if b.ID == id && b.Stats.IsHealthy {
			return b
		}
	}
	return nil // was sticky but backend is now unhealthy — caller should remap
}

// StickySet records the session → backend mapping.
func (r *Registry) StickySet(key, backendID string) {
	r.mu.Lock()
	r.ClientBackendMap[key] = backendID
	r.mu.Unlock()
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
			b.Stats.IsHealthy = healthy
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
		for {
			time.Sleep(1 * time.Second)
			
			r.mu.RLock()
			var backends []*Backend
			backends = append(backends, r.Backends...)
			r.mu.RUnlock()

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
					resp.Body.Close()
				}

				r.mu.Lock()
				b.Stats.IsHealthy = healthy
				if healthy {
					atomic.StoreInt32(&b.Stats.ActiveRequests, int32(active))
					atomic.StoreInt64(&b.Stats.TotalRequests, total)
				}
				b.Stats.Weight = math.Abs(rand.NormFloat64())
				log.Printf("[Controller] -> %s Healthy: %v, ActiveReqs: %d, TotalReqs: %d, Weight: %.2f",
					b.ID, b.Stats.IsHealthy,
					atomic.LoadInt32(&b.Stats.ActiveRequests),
					atomic.LoadInt64(&b.Stats.TotalRequests),
					b.Stats.Weight)
				r.mu.Unlock()
			}
			atomic.AddUint64(&r.Epoch, 1)
		}
	}()
}
