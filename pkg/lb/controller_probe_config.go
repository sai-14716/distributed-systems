package lb

import (
	_ "embed"
	"encoding/json"
	"log"
	"os"
)

//go:embed controller_probe_config.json
var embeddedControllerProbeConfig []byte

type ControllerProbeConfig struct {
	Assignments map[string][]string `json:"assignments"`
}

func DefaultControllerProbeAssignments() map[string][]string {
	cfg, err := parseControllerProbeConfig(embeddedControllerProbeConfig)
	if err != nil {
		log.Printf("[lb-load] failed to parse embedded controller probe config: %v", err)
		return nil
	}
	return cfg.Assignments
}

func LoadControllerProbeAssignments(path string) (map[string][]string, error) {
	if path == "" {
		return DefaultControllerProbeAssignments(), nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg, err := parseControllerProbeConfig(b)
	if err != nil {
		return nil, err
	}
	return cfg.Assignments, nil
}

func parseControllerProbeConfig(b []byte) (ControllerProbeConfig, error) {
	var cfg ControllerProbeConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
