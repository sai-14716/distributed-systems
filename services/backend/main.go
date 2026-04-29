package main

import (
	"crypto/aes"
	"crypto/cipher"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
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

	workDefaultMs := getenvInt("WORK_MS", 50)
	maxInflight := getenvInt("MAX_INFLIGHT", 100)
	cpuSampleMs := getenvInt("CPU_SAMPLE_MS", 250)
	cpuAlpha := getenvFloat("CPU_EWMA_ALPHA", 0.2)

	sampler := newCPUSampler(time.Duration(cpuSampleMs)*time.Millisecond, cpuAlpha)
	if maxInflight < 1 {
		maxInflight = 1
	}
	inflightSem = make(chan struct{}, maxInflight)

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

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "OK", "active_requests": %d, "total_requests": %d, "cpu": %.2f}`,
			atomic.LoadInt32(&backendActiveRequests), atomic.LoadInt64(&backendTotalRequests), cpuPct)

		logBackendRequest(serverID, r, http.StatusOK, "health=true heartbeat")
	})

	// /internal/load - for LB load polling (10% CPU buckets; no global-clock semantics required)
	http.HandleFunc("/internal/load", func(w http.ResponseWriter, r *http.Request) {
		ensureTraceID(r, w)
		cpuPct := sampler.CPUPct()
		bucket := cpuBucket5(cpuPct)
		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":              true,
			"server_id":       serverID,
			"cpu_pct":         cpuPct,
			"cpu_bucket":      bucket,
			"active_requests": atomic.LoadInt32(&activeRequests),
		})
		logBackendRequest(serverID, r, http.StatusOK, fmt.Sprintf("endpoint=internal/load cpu_pct=%.1f bucket=%d step=5%%", cpuPct, bucket))
	})

	// ---------------- CHAT ----------------
	http.HandleFunc("/chat", trackRequests(func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)
		if !tryAcquireInflight(w, r, serverID) {
			return
		}
		defer releaseInflight()
		atomic.AddInt32(&activeRequests, 1)
		defer atomic.AddInt32(&activeRequests, -1)

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

		burnCPU(workDuration(r, workDefaultMs))
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
		if !tryAcquireInflight(w, r, serverID) {
			return
		}
		defer releaseInflight()
		atomic.AddInt32(&activeRequests, 1)
		defer atomic.AddInt32(&activeRequests, -1)

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

		burnCPU(workDuration(r, workDefaultMs))
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
		if !tryAcquireInflight(w, r, serverID) {
			return
		}
		defer releaseInflight()
		atomic.AddInt32(&activeRequests, 1)
		defer atomic.AddInt32(&activeRequests, -1)

		burnCPU(workDuration(r, workDefaultMs))

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
