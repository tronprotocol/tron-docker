package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// fileConfig represents the YAML configuration file structure.
//
// Example:
// metrics_addr: "0.0.0.0:9090"
// interval: "10s"
// nodes:
//   - label: "tron-node1"
//     url: "http://tron-node1:8090"
//   - label: "tron-node2"
//     url: "http://tron-node2:8090"
type fileConfig struct {
	MetricsAddr           string        `yaml:"metrics_addr"`
	Interval              string        `yaml:"interval"`
	SRBorderlineThreshold int64         `yaml:"sr_borderline_threshold"`
	Nodes                 []fileCfgNode `yaml:"nodes"`
}

type fileCfgNode struct {
	Label string `yaml:"label"`
	URL   string `yaml:"url"`
}

// loadConfig reads and parses the YAML config file, returning node configurations,
// optional metrics address override, optional interval override, and optional SR borderline threshold override.
func loadConfig(path string) ([]nodeCfg, string, time.Duration, int64) {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("failed to read config file %q: %w", path, err))
	}

	var cfg fileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Errorf("failed to parse config file %q: %w", path, err))
	}

	var nodes []nodeCfg
	for _, n := range cfg.Nodes {
		if n.Label == "" || n.URL == "" {
			continue
		}
		nodes = append(nodes, nodeCfg{
			Label: n.Label,
			URL:   n.URL,
		})
	}

	var interval time.Duration
	if cfg.Interval != "" {
		d, err := time.ParseDuration(cfg.Interval)
		if err != nil {
			panic(fmt.Errorf("invalid interval %q in config file %q: %w", cfg.Interval, path, err))
		}
		interval = d
	}

	return nodes, cfg.MetricsAddr, interval, cfg.SRBorderlineThreshold
}
