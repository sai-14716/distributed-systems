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
		targetsArg      = flag.String("targets", "", "comma-separated list of node base URLs")
		configPath      = flag.String("config", getenvDefault("CLUSTER_CONFIG", "/app/cluster_config.yaml"), "cluster topology config")
		algorithm       = flag.String("algorithm", "round_robin", "algorithm name")
		probeIntervalMs = flag.Int("probe-interval-ms", 1000, "probe interval ms")
		deadline        = flag.Duration("timeout", 10*time.Second, "overall timeout")
	)
	flag.Parse()

	targets, targetByID, err := resolveTargets(*targetsArg, *configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: *deadline}

	leader := waitForLeader(client, targets, targetByID, *deadline)
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

type clusterConfig struct {
	Laptops map[string]string `json:"laptops"`
	Nodes   map[string]struct {
		Laptop string `json:"laptop"`
		Port   int    `json:"port"`
	} `json:"nodes"`
}

func resolveTargets(targetsArg, configPath string) ([]string, map[string]string, error) {
	if targetsArg == "" {
		targetsArg = os.Getenv("TARGETS")
	}
	if targetsArg != "" {
		targets := parseTargets(targetsArg)
		if len(targets) == 0 {
			return nil, nil, fmt.Errorf("no targets provided")
		}
		return targets, map[string]string{}, nil
	}

	targets, targetByID, err := targetsFromConfig(configPath)
	if err != nil {
		return nil, nil, err
	}
	if len(targets) == 0 {
		return nil, nil, fmt.Errorf("no nodes configured in %s", configPath)
	}
	return targets, targetByID, nil
}

func targetsFromConfig(path string) ([]string, map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read cluster config %s: %w", path, err)
	}
	defer f.Close()

	var cfg clusterConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to parse cluster config %s: %w", path, err)
	}

	targets := make([]string, 0, len(cfg.Nodes))
	targetByID := make(map[string]string, len(cfg.Nodes))
	for nodeID, info := range cfg.Nodes {
		ip, ok := cfg.Laptops[info.Laptop]
		if !ok || ip == "" {
			return nil, nil, fmt.Errorf("node %s references unknown laptop %q", nodeID, info.Laptop)
		}
		url := fmt.Sprintf("http://%s:%d", ip, info.Port)
		targets = append(targets, url)
		targetByID[nodeID] = url
	}
	return targets, targetByID, nil
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

func waitForLeader(client *http.Client, targets []string, targetByID map[string]string, timeout time.Duration) string {
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
				if leaderURL := targetByID[state.LeaderID]; leaderURL != "" {
					return leaderURL
				}
				if leaderURL := leaderURLFromID(targets, state.LeaderID); leaderURL != "" {
					return leaderURL
				}
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
		reachable := 0
		for _, t := range targets {
			state, err := getState(client, t)
			if err != nil {
				continue
			}
			reachable++
			if state.Config != cfg {
				all = false
			}
		}
		if reachable > 0 && all {
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
