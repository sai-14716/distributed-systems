package lb

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type Forwarder struct {
	Registry *Registry
	algo     atomic.Value // stores algorithmHolder
}

type algorithmHolder struct {
	algo Algorithm
}

type raftStateView struct {
	Role     string `json:"role"`
	LeaderID string `json:"leader_id"`
}

func NewForwarder(r *Registry, algo Algorithm) *Forwarder {
	f := &Forwarder{Registry: r}
	f.SetAlgorithm(algo)
	return f
}

func (f *Forwarder) SetAlgorithm(algo Algorithm) {
	if algo == nil {
		return
	}
	f.algo.Store(algorithmHolder{algo: algo})
}

func (f *Forwarder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	nodeID := getenvOr("NODE_ID", "unknown-node")
	role, leaderID := getLocalRaftRoleLeader()
	traceID := req.Header.Get("X-Trace-ID")
	if traceID == "" {
		traceID = time.Now().Format("20060102T150405.000000000")
	}
	req.Header.Set("X-Trace-ID", traceID)
	req.Header.Set("X-LB-Node", nodeID)
	if p := req.Header.Get("X-Packet-Path"); p != "" {
		req.Header.Set("X-Packet-Path", p+" -> LB("+nodeID+")")
	} else {
		req.Header.Set("X-Packet-Path", "LB("+nodeID+")")
	}
	w.Header().Set("X-Trace-ID", traceID)
	w.Header().Set("X-LB-Node", nodeID)

	holder, _ := f.algo.Load().(algorithmHolder)
	a := holder.algo
	if a == nil {
		http.Error(w, "503 Service Unavailable - No Algorithm", http.StatusServiceUnavailable)
		log.Printf("[lb-data] trace=%s node=%s role=%s leader=%s method=%s path=%s status=503 reason=no_algorithm", traceID, nodeID, role, leaderID, req.Method, req.URL.Path)
		return
	}

	pool := f.Registry.MatchSubset(req.URL.Path, req.Header.Get("X-Role"))
	poolSource := "sticky_subset"
	if pool == nil {
		pool = f.Registry.GetHealthyBackends()
		poolSource = "all_healthy"
	}
	poolIDs := backendIDs(pool)
	algoName := algorithmName(a)
	key := sessionKey(req)

	target := a.NextBackend(f.Registry, req)

	if target == nil {
		http.Error(w, "503 Service Unavailable - No Healthy Backends", http.StatusServiceUnavailable)
		log.Printf("[lb-data] trace=%s node=%s role=%s leader=%s method=%s path=%s status=503 reason=no_healthy_backends algo=%s pool_source=%s pool=%s session_key=%s", traceID, nodeID, role, leaderID, req.Method, req.URL.Path, algoName, poolSource, poolIDs, key)
		return
	}
	w.Header().Set("X-LB-Backend", target.ID)

	// Removed Increment/Decrement active since backends now natively track requests
	// Propagate real client IP for downstream L7 inspection
	req.Header.Set("X-Forwarded-For", req.RemoteAddr)

	target.Proxy.ServeHTTP(w, req)
	log.Printf("[lb-data] trace=%s node=%s role=%s leader=%s method=%s path=%s algo=%s pool_source=%s pool=%s session_key=%s chosen_backend=%s upstream=%s latency_ms=%d", traceID, nodeID, role, leaderID, req.Method, req.URL.Path, algoName, poolSource, poolIDs, key, target.ID, target.URL.String(), time.Since(start).Milliseconds())
}

func algorithmName(a Algorithm) string {
	t := fmt.Sprintf("%T", a)
	t = strings.TrimPrefix(t, "*")
	if idx := strings.LastIndex(t, "."); idx >= 0 {
		return t[idx+1:]
	}
	return t
}

func backendIDs(backends []*Backend) string {
	if len(backends) == 0 {
		return "none"
	}
	ids := make([]string, 0, len(backends))
	for _, b := range backends {
		if b != nil {
			ids = append(ids, b.ID)
		}
	}
	if len(ids) == 0 {
		return "none"
	}
	return strings.Join(ids, ",")
}

func getenvOr(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getLocalRaftRoleLeader() (string, string) {
	client := &http.Client{Timeout: 150 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:19090/state")
	if err != nil {
		return "unknown", "unknown"
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return "unknown", "unknown"
	}
	var st raftStateView
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		return "unknown", "unknown"
	}
	role := st.Role
	if role == "" {
		role = "unknown"
	}
	leader := st.LeaderID
	if leader == "" {
		leader = "unknown"
	}
	return role, leader
}
