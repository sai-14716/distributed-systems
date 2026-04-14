package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

// adminMux wraps the forwarder and adds /admin/* control endpoints.
type adminMux struct {
	forwarder *Forwarder
	registry  *Registry
}

func (a *adminMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/fail-backend" && r.Method == http.MethodPost {
		var req struct {
			Backend string `json:"backend"`
			Healthy bool   `json:"healthy"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		a.registry.SimulateBackendFailure(req.Backend, req.Healthy)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"backend":"%s","healthy":%v}`, req.Backend, req.Healthy)
		return
	}

	if r.URL.Path == "/admin/status" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	a.forwarder.ServeHTTP(w, r)
}

func main() {
	registry := NewRegistry()

	targets := []string{
		"http://backend-1:8080",
		"http://backend-2:8080",
		"http://backend-3:8080",
	}

	for i, t := range targets {
		u, _ := url.Parse(t)
		backendID := fmt.Sprintf("backend-%d", i+1)
		log.Printf("Registering target: %s (%s)", t, backendID)
		registry.AddBackend(backendID, u)
	}

	registry.StartController()

	algoType := os.Getenv("ALGO")
	var algo Algorithm
	switch algoType {
	case "least-req":
		log.Println("Using LeastRequests algorithm")
		algo = &LeastRequests{}
	case "wrr":
		log.Println("Using WeightedRoundRobin algorithm")
		algo = &WeightedRoundRobin{}
	case "maglev":
		log.Println("Using Maglev Consistent Hashing algorithm")
		algo = NewMaglev()
	default:
		log.Println("Using RoundRobin algorithm")
		algo = &RoundRobin{}
	}

	forwarder := NewForwarder(registry, algo)

	mux := &adminMux{forwarder: forwarder, registry: registry}

	log.Println("Load Balancer starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}