package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	serverID := os.Getenv("SERVER_ID")
	port := "8080"

	// /health - for LB health probing
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Server-ID", serverID)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	// /chat - GPT-style chat endpoint
	// Reads X-Chat-ID header, supports X-Slow: true for HoL blocking tests
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		chatID := r.Header.Get("X-Chat-ID")
		if r.Header.Get("X-Slow") == "true" {
			time.Sleep(200 * time.Millisecond)
		}
		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"server":"%s","chat_id":"%s","status":"ok"}`, serverID, chatID)
	})

	// /payload - echoes body size back; used for multi-packet / large payload tests
	http.HandleFunc("/payload", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"server":"%s","bytes_received":%d}`, serverID, len(body))
	})

	// / - default catch-all (for stress-test.js baseline)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Server-ID", serverID)
		fmt.Fprintf(w, "ACK: %s", serverID)
	})

	fmt.Printf("Backend %s starting on :%s\n", serverID, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}