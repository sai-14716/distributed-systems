package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

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

	// /health - for LB health probing
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ensureTraceID(r, w)
		w.Header().Set("X-Server-ID", serverID)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
		logBackendRequest(serverID, r, http.StatusOK, "health=true")
	})

	// /chat - GPT-style chat endpoint
	// Reads X-Chat-ID header, supports X-Slow: true for HoL blocking tests
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)
		chatID := r.Header.Get("X-Chat-ID")
		if r.Header.Get("X-Slow") == "true" {
			time.Sleep(200 * time.Millisecond)
		}
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
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
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
		w.Header().Set("X-Server-ID", serverID)
		fmt.Fprintf(w, "ACK: %s", serverID)
		logBackendRequest(serverID, r, http.StatusOK, "endpoint=default")
	})

	fmt.Printf("Backend %s starting on :%s\n", serverID, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
