package raft

import "testing"

func TestValidateCommandRejectsUnknownAlgorithm(t *testing.T) {
	cmd := &Command{
		Type: "set_config",
		Data: map[string]any{
			"algorithm": "kuchbhi",
		},
	}
	if err := ValidateCommand(cmd); err == nil {
		t.Fatalf("expected error for unknown algorithm")
	}
}

func TestValidateCommandNormalizesAliases(t *testing.T) {
	cmd := &Command{
		Type: "set_config",
		Data: map[string]any{
			"algorithm":         "weighted_round_robin",
			"probe_interval_ms": float64(1000),
		},
	}
	if err := ValidateCommand(cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cmd.Data["algorithm"]; got != "wrr" {
		t.Fatalf("expected normalized algorithm 'wrr', got %v", got)
	}
}

func TestValidateCommandRejectsOutOfRangeValues(t *testing.T) {
	cases := []Command{
		{Type: "set_config", Data: map[string]any{"probe_interval_ms": float64(10)}},
	}
	for i := range cases {
		cmd := cases[i]
		if err := ValidateCommand(&cmd); err == nil {
			t.Fatalf("case %d: expected validation error", i)
		}
	}
}
