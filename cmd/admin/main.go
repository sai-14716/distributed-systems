package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"ds-loadbalancer/internal/raft"
)

func main() {
	var (
		targetsArg      = flag.String("targets", getenvDefault("TARGETS", ""), "comma-separated list of node base URLs")
		algorithm       = flag.String("algorithm", "round_robin", "algorithm name")
		probeIntervalMs = flag.Int("probe-interval-ms", 1000, "probe interval ms")
		deadline        = flag.Duration("timeout", 10*time.Second, "overall timeout")
	)
	flag.Parse()

	if *targetsArg == "" {
		fmt.Println("no targets provided")
		os.Exit(1)
	}
	targets := parseTargets(*targetsArg)

	client := &http.Client{Timeout: 800 * time.Millisecond}

	leader := waitForLeader(client, targets, *deadline)
	if leader == "" {
		fmt.Println("failed to detect leader")
		os.Exit(1)
	}

	cmd := raft.Command{
		Type: "set_config",
		Data: map[string]any{
			"algorithm":         *algorithm,
			"probe_interval_ms": *probeIntervalMs,
		},
	}

	if err := submitCommand(client, leader, cmd); err != nil {
		fmt.Println("submit failed:", err)
		os.Exit(1)
	}

	ok := waitForConfig(client, targets, raft.Config{
		Algorithm:       *algorithm,
		ProbeIntervalMs: *probeIntervalMs,
	}, *deadline)
	if !ok {
		fmt.Println("config did not converge on all nodes")
		os.Exit(1)
	}

	fmt.Println("config update committed and applied on all nodes")
}

func parseTargets(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func waitForLeader(client *http.Client, targets []string, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, t := range targets {
			state, err := getState(client, t)
			if err != nil {
				continue
			}
			if state.Role == raft.Leader {
				return t
			}
			if state.LeaderID != "" {
				return leaderURLFromID(targets, state.LeaderID)
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return ""
}

func submitCommand(client *http.Client, leader string, cmd raft.Command) error {
	body, _ := json.Marshal(cmd)
	req, err := http.NewRequest(http.MethodPost, leader+"/admin/submit", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var reply raft.SubmitReply
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return err
	}
	if !reply.Accepted {
		return fmt.Errorf("not accepted, leader=%s", reply.Leader)
	}
	return nil
}

func waitForConfig(client *http.Client, targets []string, cfg raft.Config, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		all := true
		for _, t := range targets {
			state, err := getState(client, t)
			if err != nil {
				all = false
				continue
			}
			if state.Config != cfg {
				all = false
			}
		}
		if all {
			return true
		}
		time.Sleep(300 * time.Millisecond)
	}
	return false
}

func getState(client *http.Client, target string) (raft.StateView, error) {
	resp, err := client.Get(target + "/state")
	if err != nil {
		return raft.StateView{}, err
	}
	defer resp.Body.Close()
	var state raft.StateView
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return raft.StateView{}, err
	}
	return state, nil
}

func leaderURLFromID(targets []string, leaderID string) string {
	for _, t := range targets {
		if strings.Contains(t, leaderID) {
			return t
		}
	}
	return ""
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
