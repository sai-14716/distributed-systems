package main

import (
	"net/http"
)

type Forwarder struct {
	Registry *Registry
	Algo     Algorithm
}

func NewForwarder(r *Registry, algo Algorithm) *Forwarder {
	return &Forwarder{Registry: r, Algo: algo}
}

func (f *Forwarder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	target := f.Algo.NextBackend(f.Registry, req)

	if target == nil {
		http.Error(w, "503 Service Unavailable - No Healthy Backends", http.StatusServiceUnavailable)
		return
	}

	f.Registry.IncrementActive(target)
	defer f.Registry.DecrementActive(target)

	// Propagate real client IP for downstream L7 inspection
	req.Header.Set("X-Forwarded-For", req.RemoteAddr)

	target.Proxy.ServeHTTP(w, req)
}
