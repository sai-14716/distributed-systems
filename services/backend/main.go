package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var activeRequests int32
var inflightSem chan struct{}

func ensureTraceID(r *http.Request, w http.ResponseWriter) string {
	traceID := r.Header.Get("X-Trace-ID")
	if traceID == "" {
		traceID = fmt.Sprintf("backend-%d", time.Now().UnixNano())
	}
	w.Header().Set("X-Trace-ID", traceID)
	return traceID
}

func logBackendRequest(serverID string, r *http.Request, status int, extra string) {
	log.Printf("[backend] trace=%s server=%s method=%s path=%s status=%d session=%s chat=%s from=%s %s",
		r.Header.Get("X-Trace-ID"),
		serverID,
		r.Method,
		r.URL.Path,
		status,
		r.Header.Get("X-Session-ID"),
		r.Header.Get("X-Chat-ID"),
		r.Header.Get("X-Forwarded-For"),
		extra,
	)
}

func main() {
	serverID := os.Getenv("SERVER_ID")
	port := "8080"

	workDefaultMs := getenvInt("WORK_MS", 50)
	maxInflight := getenvInt("MAX_INFLIGHT", 100)
	cpuSampleMs := getenvInt("CPU_SAMPLE_MS", 500)
	cpuAlpha := getenvFloat("CPU_EWMA_ALPHA", 0.2)

	sampler := newCPUSampler(time.Duration(cpuSampleMs)*time.Millisecond, cpuAlpha)
	if maxInflight < 1 {
		maxInflight = 1
	}
	inflightSem = make(chan struct{}, maxInflight)

	// /health - for LB health probing
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ensureTraceID(r, w)
		w.Header().Set("X-Server-ID", serverID)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
		logBackendRequest(serverID, r, http.StatusOK, "health=true")
	})

	// /internal/load - for LB load polling (10% CPU buckets; no global-clock semantics required)
	http.HandleFunc("/internal/load", func(w http.ResponseWriter, r *http.Request) {
		ensureTraceID(r, w)
		cpuPct := sampler.CPUPct()
		bucket := cpuBucket10(cpuPct)
		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":              true,
			"server_id":       serverID,
			"cpu_pct":         cpuPct,
			"cpu_bucket":      bucket,
			"active_requests": atomic.LoadInt32(&activeRequests),
		})
		logBackendRequest(serverID, r, http.StatusOK, fmt.Sprintf("endpoint=internal/load cpu_pct=%.1f bucket=%d", cpuPct, bucket))
	})

	// /chat - GPT-style chat endpoint
	// Reads X-Chat-ID header, supports X-Slow: true for HoL blocking tests
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)
		if !tryAcquireInflight(w, r, serverID) {
			return
		}
		defer releaseInflight()
		atomic.AddInt32(&activeRequests, 1)
		defer atomic.AddInt32(&activeRequests, -1)

		chatID := r.Header.Get("X-Chat-ID")
		if r.Header.Get("X-Slow") == "true" {
			time.Sleep(200 * time.Millisecond)
		}
		burnCPU(workDuration(r, workDefaultMs))
		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"server":"%s","chat_id":"%s","status":"ok","ack":"ack from %s for trace %s"}`,
			serverID,
			chatID,
			serverID,
			traceID,
		)
		logBackendRequest(serverID, r, http.StatusOK, "endpoint=chat")
	})

	// /payload - echoes body size back; used for multi-packet / large payload tests
	http.HandleFunc("/payload", func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)
		if !tryAcquireInflight(w, r, serverID) {
			return
		}
		defer releaseInflight()
		atomic.AddInt32(&activeRequests, 1)
		defer atomic.AddInt32(&activeRequests, -1)

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		burnCPU(workDuration(r, workDefaultMs))
		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"server":"%s","bytes_received":%d,"ack":"ack from %s for trace %s"}`,
			serverID,
			len(body),
			serverID,
			traceID,
		)
		logBackendRequest(serverID, r, http.StatusOK, fmt.Sprintf("endpoint=payload bytes=%d", len(body)))
	})

	// / - default catch-all (for stress-test.js baseline)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)
		if !tryAcquireInflight(w, r, serverID) {
			return
		}
		defer releaseInflight()
		atomic.AddInt32(&activeRequests, 1)
		defer atomic.AddInt32(&activeRequests, -1)

		burnCPU(workDuration(r, workDefaultMs))
		w.Header().Set("X-Server-ID", serverID)
		fmt.Fprintf(w, "ACK: %s", serverID)
		logBackendRequest(serverID, r, http.StatusOK, "endpoint=default")
	})

	fmt.Printf("Backend %s starting on :%s\n", serverID, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func tryAcquireInflight(w http.ResponseWriter, r *http.Request, serverID string) bool {
	select {
	case inflightSem <- struct{}{}:
		return true
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "busy")
		logBackendRequest(serverID, r, http.StatusServiceUnavailable, "busy=true")
		return false
	}
}

func releaseInflight() {
	select {
	case <-inflightSem:
	default:
	}
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func getenvFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f <= 0 || f > 1 {
		return def
	}
	return f
}
