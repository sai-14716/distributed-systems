package main

import (
	"fmt"
	"io"
<<<<<<< Updated upstream
	"log"
	"net/http"
	"os"
	"time"
)

=======
	"crypto/aes"
	"crypto/cipher"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"log"
	"math/rand"
	"net/http"
	// "os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)


>>>>>>> Stashed changes
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

<<<<<<< Updated upstream
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
=======
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

	// serverID := os.Getenv("SERVER_ID")
	//Testing
	serverID := "test-backend"
	port := "8088"

	// /health - for LB health probing
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Slow") == "true" {
		delay := time.Duration(100/runtime.NumCPU()) * time.Millisecond
		time.Sleep(delay)
	}

	ensureTraceID(r, w)
	w.Header().Set("X-Server-ID", serverID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// get rolling averages
	cpuAvg, netSent, netRecv, diskRead, diskWrite := getAverages()

	// memory snapshot (current)
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
>>>>>>> Stashed changes

	// /chat - GPT-style chat endpoint
	// Reads X-Chat-ID header, supports X-Slow: true for HoL blocking tests
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)
		chatID := r.Header.Get("X-Chat-ID")
<<<<<<< Updated upstream
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
=======

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
>>>>>>> Stashed changes
		)
		logBackendRequest(serverID, r, http.StatusOK, "endpoint=chat")
	})

	// /payload - echoes body size back; used for multi-packet / large payload tests
	http.HandleFunc("/payload", func(w http.ResponseWriter, r *http.Request) {
<<<<<<< Updated upstream
=======
		if r.Header.Get("X-Slow") == "true" {
			time.Sleep(200 * time.Millisecond)
		}
>>>>>>> Stashed changes
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
<<<<<<< Updated upstream
		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"server":"%s","bytes_received":%d,"ack":"ack from %s for trace %s"}`,
			serverID,
			len(body),
			serverID,
			traceID,
=======

		customMessage := fmt.Sprintf("Hey! You have reached this service %s at this time %s. Request payload: %s",
			serverID, time.Now().Format(time.RFC3339), string(body))

		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"server":"%s","message":%q,"size_received":"%s"}`,
			serverID,
			customMessage,
			formatBytes(uint64(len(body))),
>>>>>>> Stashed changes
		)
		logBackendRequest(serverID, r, http.StatusOK, fmt.Sprintf("endpoint=payload bytes=%d", len(body)))
	})

<<<<<<< Updated upstream
=======
	// /encrypt - aes-128 encryption of payload
	http.HandleFunc("/encrypt", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Slow") == "true" {
			time.Sleep(150 * time.Millisecond)
		}
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		key := []byte("thisis16byteskey") // Fixed AES-128 key
		block, err := aes.NewCipher(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ciphertext := make([]byte, aes.BlockSize+len(body))
		iv := ciphertext[:aes.BlockSize]
		
		// Use crypto_rand to securely generate a random IV
		if _, err := io.ReadFull(crypto_rand.Reader, iv); err != nil {
			http.Error(w, "Failed to generate random IV: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		stream := cipher.NewCFBEncrypter(block, iv)
		stream.XORKeyStream(ciphertext[aes.BlockSize:], body)

		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"server":"%s","plain_text":"%s","encrypted_hex":"%s"}`, serverID, string(body), hex.EncodeToString(ciphertext))
		logBackendRequest(serverID, r, http.StatusOK, "endpoint=encrypt")
	})

>>>>>>> Stashed changes
	// / - default catch-all (for stress-test.js baseline)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		traceID := ensureTraceID(r, w)
		r.Header.Set("X-Trace-ID", traceID)
		w.Header().Set("X-Server-ID", serverID)
<<<<<<< Updated upstream
		fmt.Fprintf(w, "ACK: %s", serverID)
=======
		fmt.Fprintf(w, "Welcome to the load balancer! (Handled by: %s)", serverID)
>>>>>>> Stashed changes
		logBackendRequest(serverID, r, http.StatusOK, "endpoint=default")
	})

	fmt.Printf("Backend %s starting on :%s\n", serverID, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
