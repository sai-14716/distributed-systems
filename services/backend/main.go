package main

import (
	"crypto/aes"
	"crypto/cipher"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
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
	serverID              string
)

func trackRequests(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&backendActiveRequests, 1)
		atomic.AddInt64(&backendTotalRequests, 1)
		defer atomic.AddInt32(&backendActiveRequests, -1)

		if p := r.Header.Get("X-Packet-Path"); p != "" {
			w.Header().Set("X-Packet-Path", p+" -> Backend("+serverID+")")
		} else {
			w.Header().Set("X-Packet-Path", "Backend("+serverID+")")
		}

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

func sendHTML(w http.ResponseWriter, title, content string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<html><head><title>%s</title></head><body style=\"font-family:sans-serif;padding:20px;\"><h2>%s</h2>%s</body></html>", title, title, content)
}

// helper to read message from query (?q=...) with fallback to body
func readMessage(r *http.Request) string {
	msg := r.URL.Query().Get("q")
	if msg == "" {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		msg = string(body)
	}
	return msg
}

func main() {
	serverID = os.Getenv("SERVER_ID")
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

		cpuPercents, err := cpu.Percent(0, false)
		cpuPct := 0.0
		if err == nil && len(cpuPercents) > 0 {
			cpuPct = cpuPercents[0]
		}

		content := fmt.Sprintf("<p><strong>Status:</strong> OK</p><p><strong>Server:</strong> %s</p><p><strong>Active Requests:</strong> %d</p><p><strong>Total Req:</strong> %d</p><p><strong>CPU:</strong> %.2f%%</p>",
			serverID, atomic.LoadInt32(&backendActiveRequests), atomic.LoadInt64(&backendTotalRequests), cpuPct)
		sendHTML(w, "Health Check", content)

		logBackendRequest(serverID, r, http.StatusOK, "health=true heartbeat")
	})

	// ---------------- CHAT ----------------
	http.HandleFunc("/chat", trackRequests(func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)

		chatID := r.Header.Get("X-Chat-ID")

		msg := readMessage(r)
		words := strings.Fields(msg)

		if r.Header.Get("X-Slow") == "true" {
			delay := 2000 * time.Millisecond
			end := time.Now().Add(delay)
			for time.Now().Before(end) {
				x := 14716.14716
				for i := 0; i < 10000; i++ {
					x = x / 2.332114554858
				}
			}
		}

		rand.Shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })
		rearranged := strings.Join(words, " ")

		if rearranged == "" {
			rearranged = "[no words received to rearrange]"
		}

		w.Header().Set("X-Server-ID", serverID)

		content := fmt.Sprintf("<p><strong>Server:</strong> %s</p><p><strong>Chat ID:</strong> %s</p><p><strong>Rearranged words:</strong></p><pre style=\"background:#eee;padding:10px;border-radius:4px;\">%s</pre>",
			serverID, chatID, rearranged)
		sendHTML(w, "Chat", content)

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

		msg := readMessage(r)

		customMessage := fmt.Sprintf(
			"Hey! You have reached this service %s at this time %s. Request payload: %s",
			serverID, time.Now().Format(time.RFC3339), msg,
		)

		w.Header().Set("X-Server-ID", serverID)

		content := fmt.Sprintf("<p><strong>Server:</strong> %s</p><p><strong>Size Received:</strong> %s</p><p><strong>Message:</strong></p><pre style=\"background:#eee;padding:10px;border-radius:4px;\">%s</pre>",
			serverID, formatBytes(uint64(len(msg))), customMessage)
		sendHTML(w, "Payload", content)

		logBackendRequest(serverID, r, http.StatusOK,
			fmt.Sprintf("endpoint=payload bytes=%d", len(msg)))
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

		msg := readMessage(r)

		key := []byte("thisis16byteskey")

		block, err := aes.NewCipher(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		plaintext := []byte(msg)
		ciphertext := make([]byte, aes.BlockSize+len(plaintext))
		iv := ciphertext[:aes.BlockSize]

		if _, err := io.ReadFull(crypto_rand.Reader, iv); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stream := cipher.NewCFBEncrypter(block, iv)
		stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

		w.Header().Set("X-Server-ID", serverID)

		content := fmt.Sprintf("<p><strong>Server:</strong> %s</p><p><strong>Plain text:</strong> %s</p><p><strong>Encrypted (Hex):</strong></p><pre style=\"background:#eee;padding:10px;border-radius:4px;word-break:break-all;\">%s</pre>",
			serverID, msg, hex.EncodeToString(ciphertext))
		sendHTML(w, "Encrypt", content)

		logBackendRequest(serverID, r, http.StatusOK, "endpoint=encrypt")
	}))

	// ---------------- DEFAULT ----------------
	http.HandleFunc("/", trackRequests(func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)

		w.Header().Set("X-Server-ID", serverID)

		content := fmt.Sprintf("<p>Welcome to the load balancer!</p><p>Handled by: <strong>%s</strong></p><p>Trace: <code>%s</code></p>", serverID, traceID)
		sendHTML(w, "Backend Node", content)

		logBackendRequest(serverID, r, http.StatusOK, "endpoint=default")
	}))

	fmt.Printf("Backend %s starting on :%s\n", serverID, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
