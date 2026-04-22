package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"ds-loadbalancer/internal/raft"
	"ds-loadbalancer/pkg/lb"
)

// adminMux wraps the forwarder and adds /admin/* control endpoints.
type adminMux struct {
	forwarder *lb.Forwarder
	registry  *lb.Registry
	state     *runtimeState
	load      *lb.ClusterLoad
}

type runtimeState struct {
	mu     sync.RWMutex
	config raft.Config
}

func (s *runtimeState) setConfig(cfg raft.Config) {
	s.mu.Lock()
	s.config = cfg
	s.mu.Unlock()
}

func (s *runtimeState) getConfig() raft.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (a *adminMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	if traceID == "" {
		traceID = fmt.Sprintf("lbctl-%d", time.Now().UnixNano())
	}
	w.Header().Set("X-Trace-ID", traceID)

	if a.load != nil {
		if r.URL.Path == "/internal/lb/health" {
			a.load.HandleHealth(w, r)
			return
		}
		if r.URL.Path == "/internal/lb/gossip" {
			a.load.HandleGossip(w, r)
			return
		}
	}

	if r.URL.Path == "/admin/status" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"config": a.state.getConfig(),
		})
		return
	}

	if r.URL.Path == "/admin/load-view" && a.load != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(a.load.Snapshot())
		return
	}

	a.forwarder.ServeHTTP(w, r)
}

func buildAlgorithm(name string) lb.Algorithm {
	switch strings.ToLower(name) {
	case "least-req", "least_requests", "leastrequests":
		log.Println("Using LeastRequests algorithm")
		return &lb.LeastRequests{}
	case "least_req":
		log.Println("Using LeastRequests algorithm")
		return &lb.LeastRequests{}
	case "least-load", "leastload", "cpu":
		log.Println("Using LeastLoad algorithm")
		return &lb.LeastLoad{}
	case "wrr", "weighted_round_robin", "weightedroundrobin":
		log.Println("Using WeightedRoundRobin algorithm")
		return &lb.WeightedRoundRobin{}
	case "maglev":
		log.Println("Using Maglev Consistent Hashing algorithm")
		return lb.NewMaglev()
	default:
		log.Println("Using RoundRobin algorithm")
		return &lb.RoundRobin{}
	}
}

func startControlAPI(addr string, forwarder *lb.Forwarder, state *runtimeState) {
	mux := http.NewServeMux()

	mux.HandleFunc("/internal/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})

	mux.HandleFunc("/internal/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(w).Encode(state.getConfig())
	})

	mux.HandleFunc("/internal/config/apply", func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = fmt.Sprintf("lbctl-%d", time.Now().UnixNano())
		}
		w.Header().Set("X-Trace-ID", traceID)

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Printf("[lb-control] trace=%s endpoint=/internal/config/apply status=405", traceID)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("[lb-control] trace=%s endpoint=/internal/config/apply status=400 error=read_body", traceID)
			return
		}
		var cfg raft.Config
		if err := json.Unmarshal(body, &cfg); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("[lb-control] trace=%s endpoint=/internal/config/apply status=400 error=invalid_json", traceID)
			return
		}
		if cfg.Algorithm == "" {
			cfg.Algorithm = state.getConfig().Algorithm
		}

		forwarder.SetAlgorithm(buildAlgorithm(cfg.Algorithm))
		state.setConfig(cfg)
		log.Printf("applied dataplane config: algorithm=%s probe_interval_ms=%d health_threshold=%.3f", cfg.Algorithm, cfg.ProbeIntervalMs, cfg.HealthThreshold)
		log.Printf("[lb-control] trace=%s endpoint=/internal/config/apply status=200 algorithm=%s probe_interval_ms=%d health_threshold=%.3f", traceID, cfg.Algorithm, cfg.ProbeIntervalMs, cfg.HealthThreshold)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "config": cfg})
	})

	server := &http.Server{Addr: addr, Handler: mux}
	log.Printf("LB control API listening on %s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	registry := lb.NewRegistry()

	backendEnv := os.Getenv("BACKENDS")
	if backendEnv == "" {
		backendEnv = "http://backend-1:8080,http://backend-2:8080,http://backend-3:8080,http://backend-4:8080,http://backend-5:8080,http://backend-6:8080,http://backend-7:8080,http://backend-8:8080,http://backend-9:8080,http://backend-10:8080"
	}
	targets := strings.Split(backendEnv, ",")

	for i, t := range targets {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		u, err := url.Parse(t)
		if err != nil {
			log.Fatalf("invalid backend url %q: %v", t, err)
		}
		backendID := fmt.Sprintf("backend-%d", i+1)
		log.Printf("Registering target: %s (%s)", t, backendID)
		registry.AddBackend(backendID, u)
	}

	initialCfg := raft.Config{
		Algorithm:       getenvDefault("ALGO", "round_robin"),
		ProbeIntervalMs: 1000,
		HealthThreshold: 0.8,
	}

	forwarder := lb.NewForwarder(registry, buildAlgorithm(initialCfg.Algorithm))
	state := &runtimeState{config: initialCfg}

	selfPeer, peers := deriveLBPeers()
	load, err := lb.NewClusterLoad(
		registry,
		selfPeer,
		peers,
		lb.ClusterLoadOptions{TTL: 5 * time.Second, RefreshEvery: 5 * time.Second},
		func() float64 { return state.getConfig().HealthThreshold },
		func() time.Duration {
			ms := state.getConfig().ProbeIntervalMs
			if ms <= 0 {
				ms = 1000
			}
			return time.Duration(ms) * time.Millisecond
		},
	)
	if err != nil {
		log.Fatalf("load monitor init failed: %v", err)
	}
	load.Start()

	mux := &adminMux{forwarder: forwarder, registry: registry, state: state, load: load}

	controlAddr := getenvDefault("LB_CONTROL_ADDR", "127.0.0.1:18080")
	go startControlAPI(controlAddr, forwarder, state)

	addr := getenvDefault("LB_ADDR", ":8080")
	log.Printf("Load Balancer starting on %s ...", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func deriveLBPeers() (lb.Peer, []lb.Peer) {
	// LBs run colocated with raft nodes; we reuse NODE_ID/SELF_URL/PEERS to build an all-to-all LB peer list.
	selfID := getenvDefault("NODE_ID", "lb-unknown")
	lbAddr := getenvDefault("LB_ADDR", ":8000")
	lbPort := portFromAddr(lbAddr, "8000")

	selfURL := os.Getenv("SELF_URL")
	selfBase := "http://127.0.0.1:" + lbPort
	if selfURL != "" {
		if u, err := url.Parse(selfURL); err == nil && u.Hostname() != "" {
			scheme := u.Scheme
			if scheme == "" {
				scheme = "http"
			}
			selfBase = scheme + "://" + u.Hostname() + ":" + lbPort
		}
	}

	self := lb.Peer{ID: selfID, BaseURL: selfBase}

	peersEnv := os.Getenv("PEERS")
	out := []lb.Peer{self}
	if peersEnv == "" {
		return self, out
	}
	for _, raw := range strings.Split(peersEnv, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || u.Hostname() == "" {
			continue
		}
		id := u.Hostname()
		if id == selfID {
			continue
		}
		scheme := u.Scheme
		if scheme == "" {
			scheme = "http"
		}
		out = append(out, lb.Peer{ID: id, BaseURL: scheme + "://" + u.Hostname() + ":" + lbPort})
	}
	return self, out
}

func portFromAddr(addr, fallback string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return fallback
	}
	// Handles ":8000", "0.0.0.0:8000", "127.0.0.1:8000".
	if idx := strings.LastIndex(addr, ":"); idx >= 0 && idx+1 < len(addr) {
		return addr[idx+1:]
	}
	return fallback
}
