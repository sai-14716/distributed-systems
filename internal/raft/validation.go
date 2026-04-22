package raft

import (
	"fmt"
	"math"
	"strings"
)

// ValidateCommand validates and normalizes externally submitted commands.
// It mutates cmd in place when normalizing aliases.
func ValidateCommand(cmd *Command) error {
	if cmd == nil {
		return fmt.Errorf("command is required")
	}
	if cmd.Type != "set_config" {
		return fmt.Errorf("unsupported command type: %q", cmd.Type)
	}
	if cmd.Data == nil {
		return fmt.Errorf("command data is required")
	}

	if raw, ok := cmd.Data["algorithm"]; ok {
		algo, ok := raw.(string)
		if !ok {
			return fmt.Errorf("algorithm must be a string")
		}
		canon, ok := normalizeAlgorithm(algo)
		if !ok {
			return fmt.Errorf("invalid algorithm %q (allowed: round_robin, wrr, least-req, least-load, maglev)", algo)
		}
		cmd.Data["algorithm"] = canon
	}

	if raw, ok := cmd.Data["probe_interval_ms"]; ok {
		v, ok := raw.(float64)
		if !ok || math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("probe_interval_ms must be a finite number")
		}
		if v != math.Trunc(v) {
			return fmt.Errorf("probe_interval_ms must be an integer")
		}
		iv := int(v)
		if iv < 50 || iv > 60000 {
			return fmt.Errorf("probe_interval_ms out of range: %d (allowed: 50..60000)", iv)
		}
	}

	if raw, ok := cmd.Data["health_threshold"]; ok {
		v, ok := raw.(float64)
		if !ok || math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("health_threshold must be a finite number")
		}
		if v <= 0 || v > 1 {
			return fmt.Errorf("health_threshold out of range: %.4f (allowed: 0 < x <= 1)", v)
		}
	}

	return nil
}

func normalizeAlgorithm(v string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "round_robin", "roundrobin", "rr":
		return "round_robin", true
	case "wrr", "weighted_round_robin", "weightedroundrobin":
		return "wrr", true
	case "least-req", "least_requests", "leastrequests":
		return "least-req", true
	case "least-load", "leastload", "cpu":
		return "least-load", true
	case "maglev":
		return "maglev", true
	default:
		return "", false
	}
}
