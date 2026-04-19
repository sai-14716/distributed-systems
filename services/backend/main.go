package main

import (
	"fmt"
	"io"
	"crypto/aes"
	"crypto/cipher"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

var (
	backendActiveRequests int32
	backendTotalRequests  int64
)

func trackRequests(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&backendActiveRequests, 1)
		atomic.AddInt64(&backendTotalRequests, 1)
		defer atomic.AddInt32(&backendActiveRequests, -1)
		h(w, r)
	}
}

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

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func main() {
	serverID := os.Getenv("SERVER_ID")
	if serverID == "" {
		serverID = "test-backend"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ---------------- HEALTH ----------------
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Slow") == "true" {
			delay := time.Duration(100/runtime.NumCPU()) * time.Millisecond
			time.Sleep(delay)
		}

		ensureTraceID(r, w)
		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")

		cpuPercents, err := cpu.Percent(0, false)
		cpuPct := 0.0
		if err == nil && len(cpuPercents) > 0 {
			cpuPct = cpuPercents[0]
		}

		fmt.Fprintf(w, `{
			"status":"OK",
			"server":"%s",
			"active_requests": %d,
			"total_requests": %d,
			"cpu_percent": %.2f
		}`,
			serverID,
			atomic.LoadInt32(&backendActiveRequests),
			atomic.LoadInt64(&backendTotalRequests),
			cpuPct,
		)

		logBackendRequest(serverID, r, http.StatusOK, "health=true heartbeat")
	})

	// ---------------- CHAT ----------------
	http.HandleFunc("/chat", trackRequests(func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)

		chatID := r.Header.Get("X-Chat-ID")

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		words := strings.Fields(string(body))

		if r.Header.Get("X-Slow") == "true" {
			delay := 2000 * time.Millisecond
			if delay == 0 {
				delay = 2000 * time.Millisecond
			}
			end := time.Now().Add(delay)
			for time.Now().Before(end) {
				x := 14716.14716
				for i := 0; i < 10000; i++ {
					x = x/2.332114554858
				}
			}
		}

		rand.Shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })
		rearranged := strings.Join(words, " ")

		if rearranged == "" {
			rearranged = "[no words received to rearrange]"
		}

		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprintf(w, `{"server":"%s","chat_id":"%s","status":"ok","rearranged":%q}`,
			serverID,
			chatID,
			rearranged,
		)

		logBackendRequest(serverID, r, http.StatusOK, "endpoint=chat")
	}))

	// ---------------- PAYLOAD ----------------
	http.HandleFunc("/payload", trackRequests(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Slow") == "true" {
			end := time.Now().Add(200 * time.Millisecond)
			for time.Now().Before(end) {
				x := 1.0000001
				for i := 0; i < 1000; i++ {
					x *= x
				}
			}
		}

		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		customMessage := fmt.Sprintf(
			"Hey! You have reached this service %s at this time %s. Request payload: %s",
			serverID, time.Now().Format(time.RFC3339), string(body),
		)

		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprintf(w, `{"server":"%s","message":%q,"size_received":"%s"}`,
			serverID,
			customMessage,
			formatBytes(uint64(len(body))),
		)

		logBackendRequest(serverID, r, http.StatusOK,
			fmt.Sprintf("endpoint=payload bytes=%d", len(body)))
	}))

	// ---------------- ENCRYPT ----------------
	http.HandleFunc("/encrypt", trackRequests(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Slow") == "true" {
			end := time.Now().Add(150 * time.Millisecond)
			for time.Now().Before(end) {
				x := 1.0000001
				for i := 0; i < 1000; i++ {
					x *= x
				}
			}
		}

		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		key := []byte("thisis16byteskey")

		block, err := aes.NewCipher(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ciphertext := make([]byte, aes.BlockSize+len(body))
		iv := ciphertext[:aes.BlockSize]

		if _, err := io.ReadFull(crypto_rand.Reader, iv); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stream := cipher.NewCFBEncrypter(block, iv)
		stream.XORKeyStream(ciphertext[aes.BlockSize:], body)

		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprintf(w, `{"server":"%s","plain_text":"%s","encrypted_hex":"%s"}`,
			serverID,
			string(body),
			hex.EncodeToString(ciphertext),
		)

		logBackendRequest(serverID, r, http.StatusOK, "endpoint=encrypt")
	}))

	// ---------------- DEFAULT ----------------
	http.HandleFunc("/", trackRequests(func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)

		w.Header().Set("X-Server-ID", serverID)

		fmt.Fprintf(w, "Welcome to the load balancer! (Handled by: %s)", serverID)

		logBackendRequest(serverID, r, http.StatusOK, "endpoint=default")
	}))

	fmt.Printf("Backend %s starting on :%s\n", serverID, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}