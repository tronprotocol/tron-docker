package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/tronprotocol/tron-docker/tools/node_monitor/monitors"
)

var (
	// Command line arguments (nodes and global settings are configured via YAML)
	configPath = flag.String("config", "", "Path to YAML config file (required for node definitions)")
)

// nodeCfg represents a single node configuration used by the monitors.
type nodeCfg struct {
	Label string
	URL   string
}

func main() {
	flag.Parse()

	// Build node configuration list
	var nodes []nodeCfg

	// Config file is now mandatory for node definitions
	if *configPath == "" {
		log.Fatal("config file is required; please provide -config path/to/node_monitor.yml")
	}

	cfgNodes, cfgMetricsAddr, cfgInterval := loadConfig(*configPath)
	if len(cfgNodes) == 0 {
		log.Fatalf("no nodes defined in config file %q", *configPath)
	}
	nodes = cfgNodes

	// Derive metrics address / interval from config (with sane defaults)
	metricsAddr := cfgMetricsAddr
	if metricsAddr == "" {
		metricsAddr = "0.0.0.0:9098"
	}
	interval := cfgInterval
	if interval <= 0 {
		interval = 10 * time.Second
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create Prometheus registry
	registry := prometheus.NewRegistry()

	// Register default Go and process metrics
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Register custom metrics
	monitors.RegisterAllMetrics(registry)

	// Start Prometheus metrics HTTP server
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	metricsServer := &http.Server{
		Addr: metricsAddr,
	}

	go func() {
		log.Printf("Starting Prometheus metrics server on %s", metricsAddr)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// Create monitors for each configured node
	for _, n := range nodes {
		log.Printf("Starting monitors for node %s (%s)", n.Label, n.URL)
		client := monitors.NewTronClient(n.URL)

		// Empty block monitor
		emptyBlockMonitor := monitors.NewEmptyBlockMonitor(client, interval, n.Label)
		go emptyBlockMonitor.Start(ctx)

		// SR set monitor
		srSetMonitor := monitors.NewSRSetMonitor(client, interval, n.Label)
		go srSetMonitor.Start(ctx)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Node monitor service started. Press Ctrl+C to stop...")
	<-sigChan

	log.Println("Shutting down service...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down metrics server: %v", err)
	}

	cancel()
	log.Println("Service stopped")
}
