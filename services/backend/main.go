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
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
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
	go collectMetrics()

	serverID := "test-backend"
	port := "8088"

	// ---------------- HEALTH ----------------
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Slow") == "true" {
			delay := time.Duration(100/runtime.NumCPU()) * time.Millisecond
			time.Sleep(delay)
		}

		ensureTraceID(r, w)
		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")

		cpuAvg, netSent, netRecv, diskRead, diskWrite := getAverages()
		vmStat, _ := mem.VirtualMemory()

		fmt.Fprintf(w, `{
			"status":"OK",
			"server":"%s",
			"cpu_avg_%ds": "%.2f%%",
			"memory_usage": "%.2f%%",
			"ram_used": "%s",
			"goroutines": %d,
			"network_sent_per_%ds": "%s",
			"network_recv_per_%ds": "%s",
			"disk_read_per_%ds": "%s",
			"disk_write_per_%ds": "%s"
		}`,
			serverID,
			HEART_BEAT,
			cpuAvg,
			vmStat.UsedPercent,
			formatBytes(vmStat.Used),
			runtime.NumGoroutine(),
			HEART_BEAT, formatBytes(netSent),
			HEART_BEAT, formatBytes(netRecv),
			HEART_BEAT, formatBytes(diskRead),
			HEART_BEAT, formatBytes(diskWrite),
		)

		logBackendRequest(serverID, r, http.StatusOK, "health=true heartbeat")
	})

	// ---------------- CHAT ----------------
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)

		chatID := r.Header.Get("X-Chat-ID")

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		words := strings.Fields(string(body))

		if r.Header.Get("X-Slow") == "true" {
			delay := time.Duration(len(words)*50) * time.Millisecond
			if delay == 0 {
				delay = 100 * time.Millisecond
			}
			time.Sleep(delay)
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
	})

	// ---------------- PAYLOAD ----------------
	http.HandleFunc("/payload", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Slow") == "true" {
			time.Sleep(200 * time.Millisecond)
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
	})

	// ---------------- ENCRYPT ----------------
	http.HandleFunc("/encrypt", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Slow") == "true" {
			time.Sleep(150 * time.Millisecond)
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
	})

	// ---------------- DEFAULT ----------------
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)

		w.Header().Set("X-Server-ID", serverID)

		fmt.Fprintf(w, "Welcome to the load balancer! (Handled by: %s)", serverID)

		logBackendRequest(serverID, r, http.StatusOK, "endpoint=default")
	})

	fmt.Printf("Backend %s starting on :%s\n", serverID, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}