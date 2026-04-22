package main

import (
	"net/http"
	"strconv"
	"time"
)

func workDuration(r *http.Request, defMs int) time.Duration {
	ms := defMs
	// Optional per-request override (used for experiments / tests).
	if v := r.Header.Get("X-Work-Ms"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			ms = n
		}
	}
	if ms > 5000 {
		ms = 5000
	}
	return time.Duration(ms) * time.Millisecond
}

// burnCPU consumes CPU until the duration elapses (no sleeps).
func burnCPU(d time.Duration) {
	if d <= 0 {
		return
	}
	deadline := time.Now().Add(d)

	// Simple integer mixing loop; fast and predictable enough to drive CPU load.
	var x uint64 = 0x9e3779b97f4a7c15
	for time.Now().Before(deadline) {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
		x *= 0x2545F4914F6CDD1D
	}
	_ = x
}
