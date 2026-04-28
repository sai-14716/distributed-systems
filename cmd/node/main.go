package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ds-loadbalancer/internal/raft"
)

func main() {
	id := getenvDefault("NODE_ID", "node-1")
	addr := getenvDefault("HTTP_ADDR", ":19090")
	selfURL := os.Getenv("SELF_URL")
	stateFile := getenvDefault("STATE_FILE", "/var/lib/raft/state.json")
	activityLogFile := getenvDefault("ACTIVITY_LOG_FILE", "/var/lib/raft/activity.log")
	dataplaneControlURL := getenvDefault("DATAPLANE_CONTROL_URL", "http://127.0.0.1:18080")
	peersEnv := os.Getenv("PEERS")
	peers := parsePeers(peersEnv, selfURL)

	setupActivityLogging(activityLogFile)

	node := raft.NewNode(id, addr, peers, stateFile)
	node.Start()

	node.OnConfigApplied(func(cfg raft.Config) {
		go func() {
			if err := pushDataplaneConfig(dataplaneControlURL, cfg, 30, 150*time.Millisecond); err != nil {
				log.Printf("failed to push config to dataplane: %v", err)
			}
		}()
	})

	// Push current committed config on startup to initialize the local dataplane process.
	go func(cfg raft.Config) {
		if err := pushDataplaneConfig(dataplaneControlURL, cfg, 60, 200*time.Millisecond); err != nil {
			log.Printf("initial dataplane config sync failed: %v", err)
		}
	}(node.State().Config)

	mux := http.NewServeMux()
	mux.HandleFunc("/raft/request-vote", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var args raft.RequestVoteArgs
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		reply := node.HandleRequestVote(args)
		writeJSON(w, reply)
	})

	mux.HandleFunc("/raft/append-entries", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var args raft.AppendEntriesArgs
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		reply := node.HandleAppendEntries(args)
		writeJSON(w, reply)
	})

	handleSubmit := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Read raw body so we can forward the exact request payload if this node isn't leader.
		raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var cmd raft.Command
		if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&cmd); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := raft.ValidateCommand(&cmd); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"accepted": false,
				"error":    err.Error(),
			})
			return
		}

		reply, submitErr := node.Submit(cmd)
		if submitErr == nil {
			writeJSON(w, reply)
			return
		}

		// Transparently forward to leader (DNS round-robin compatibility):
		// clients can hit any LB; non-leaders forward writes to the leader.
		if r.Header.Get("X-Raft-Forwarded") != "" {
			w.WriteHeader(http.StatusServiceUnavailable)
			writeJSON(w, map[string]any{
				"accepted": false,
				"error":    "submit reached non-leader after forward; leader unknown/unreachable",
				"leader":   reply.Leader,
			})
			return
		}

		leaderURL := resolveLeaderURL(selfURL, peers, reply.Leader)
		if leaderURL == "" {
			leaderURL = discoverLeaderURL(selfURL, peers)
		}
		if leaderURL == "" {
			w.WriteHeader(http.StatusServiceUnavailable)
			writeJSON(w, map[string]any{
				"accepted": false,
				"error":    "leader unknown; try again",
			})
			return
		}

		status, respBody, ferr := forwardSubmit(leaderURL, raw)
		if ferr != nil {
			w.WriteHeader(http.StatusBadGateway)
			writeJSON(w, map[string]any{
				"accepted": false,
				"error":    "failed to forward to leader",
				"leader":   leaderURL,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(respBody)
	}

	// Admin API endpoint for command submission.
	mux.HandleFunc("/admin/submit", handleSubmit)
	// Backward-compatible alias used by older clients.
	mux.HandleFunc("/client/submit", handleSubmit)

	mux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		state := node.State()
		writeJSON(w, state)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, map[string]any{"ok": true, "node": id})
	})

	server := &http.Server{Addr: addr, Handler: mux}
	log.Printf("raft node %s listening on %s, self_url=%q, peers=%v, state_file=%s", id, addr, selfURL, peers, stateFile)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func pushDataplaneConfig(baseURL string, cfg raft.Config, maxAttempts int, delay time.Duration) error {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if err := applyDataplaneConfig(baseURL, cfg); err == nil {
			log.Printf("pushed dataplane config: algorithm=%s probe_interval_ms=%d", cfg.Algorithm, cfg.ProbeIntervalMs)
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(delay)
	}
	return lastErr
}

func applyDataplaneConfig(baseURL string, cfg raft.Config) error {
	body, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	c := &http.Client{Timeout: 800 * time.Millisecond}
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/internal/config/apply", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return &httpStatusError{statusCode: resp.StatusCode, body: string(b)}
	}
	return nil
}

type httpStatusError struct {
	statusCode int
	body       string
}

func (e *httpStatusError) Error() string {
	if e.body == "" {
		return "dataplane apply failed with status " + http.StatusText(e.statusCode)
	}
	return "dataplane apply failed: status=" + http.StatusText(e.statusCode) + " body=" + e.body
}

func parsePeers(peers, selfURL string) []string {
	if peers == "" {
		return nil
	}
	parts := strings.Split(peers, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	selfURL = strings.TrimRight(strings.TrimSpace(selfURL), "/")
	for _, p := range parts {
		p = strings.TrimRight(strings.TrimSpace(p), "/")
		if p == "" || p == selfURL {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func setupActivityLogging(logFile string) {
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		log.Printf("activity log disabled (mkdir failed): %v", err)
		return
	}
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("activity log disabled (open failed): %v", err)
		return
	}
	log.SetOutput(io.MultiWriter(os.Stdout, f))
}

func resolveLeaderURL(selfURL string, peers []string, leaderID string) string {
	if leaderID == "" {
		return ""
	}
	// In our docker setup, leader IDs match the service names (node1..node5),
	// and peer URLs contain that name (e.g. http://node3:19090).
	if selfURL != "" && strings.Contains(selfURL, leaderID) {
		return selfURL
	}
	for _, p := range peers {
		if strings.Contains(p, leaderID) {
			return p
		}
	}
	return ""
}

func discoverLeaderURL(selfURL string, peers []string) string {
	targets := make([]string, 0, 1+len(peers))
	if selfURL != "" {
		targets = append(targets, selfURL)
	}
	targets = append(targets, peers...)

	c := &http.Client{Timeout: 400 * time.Millisecond}
	for _, t := range targets {
		resp, err := c.Get(t + "/state")
		if err != nil {
			continue
		}
		var st raft.StateView
		_ = json.NewDecoder(resp.Body).Decode(&st)
		_ = resp.Body.Close()
		if st.Role == raft.Leader {
			return t
		}
	}
	return ""
}

func forwardSubmit(leaderURL string, raw []byte) (int, []byte, error) {
	c := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest(http.MethodPost, leaderURL+"/admin/submit", bytes.NewReader(raw))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Raft-Forwarded", "1")
	resp, err := c.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, body, nil
}
