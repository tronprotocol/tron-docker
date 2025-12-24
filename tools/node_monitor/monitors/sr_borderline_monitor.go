package monitors

import (
	"context"
	"log"
	"time"
)

// Default threshold for 27th and 28th SR voteCount (100 million).
const defaultSRBorderlineThreshold int64 = 100_000_000

// SRBorderlineMonitor checks whether the 27th and 28th SR voteCounts
// fall below a configurable threshold and updates SRBorderlineLow metric.
type SRBorderlineMonitor struct {
	client    *TronClient
	interval  time.Duration
	nodeLabel string
	threshold int64
}

// NewSRBorderlineMonitor creates a new SRBorderlineMonitor.
// threshold <= 0 will fall back to defaultSRBorderlineThreshold.
func NewSRBorderlineMonitor(client *TronClient, interval time.Duration, nodeLabel string, threshold int64) *SRBorderlineMonitor {
	if threshold <= 0 {
		threshold = defaultSRBorderlineThreshold
	}
	return &SRBorderlineMonitor{
		client:    client,
		interval:  interval,
		nodeLabel: nodeLabel,
		threshold: threshold,
	}
}

// Start starts the SR borderline monitoring loop.
func (m *SRBorderlineMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	log.Println("SR borderline monitor started")

	// Initial check
	m.checkBorderline(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("SR borderline monitor stopped")
			return
		case <-ticker.C:
			m.checkBorderline(ctx)
		}
	}
}

// checkBorderline fetches the 27th and 28th SRs and updates SRBorderlineLow based on the voteCount gap between them.
func (m *SRBorderlineMonitor) checkBorderline(ctx context.Context) {
	// Fetch 27th and 28th SRs (offset 26, limit 2) ordered by voteCount.
	witnesses, err := m.client.GetPaginatedNowWitnessList(26, 2)
	if err != nil {
		// Skip check during maintenance period
		if IsMaintenanceError(err) {
			log.Printf("Skip SR borderline check for node %s due to maintenance period", m.nodeLabel)
			return
		}
		log.Printf("Failed to get borderline witnesses: %v", err)
		NodeMetrics.APICallErrors.WithLabelValues(m.nodeLabel, "getpaginatednowwitnesslist", "request_failed").Inc()
		return
	}

	if len(witnesses) < 2 {
		// Not enough witnesses to evaluate gap between 27th and 28th; treat as non-triggered.
		log.Printf("Not enough witnesses returned for SR gap check: got=%d, need=2", len(witnesses))
		SRSetMetrics.SRBorderlineLow.WithLabelValues(m.nodeLabel).Set(0)
		NodeMetrics.LastCheckTime.WithLabelValues(m.nodeLabel, "sr_borderline").Set(float64(time.Now().Unix()))
		return
	}

	v27 := witnesses[0].VoteCount
	v28 := witnesses[1].VoteCount
	gap := v27 - v28

	// Trigger condition: voteCount gap between 27th and 28th is below threshold.
	if gap < m.threshold {
		log.Printf("SR borderline gap detected: 27th=%d, 28th=%d, gap=%d, threshold=%d", v27, v28, gap, m.threshold)
		SRSetMetrics.SRBorderlineLow.WithLabelValues(m.nodeLabel).Set(1)
	} else {
		SRSetMetrics.SRBorderlineLow.WithLabelValues(m.nodeLabel).Set(0)
	}
	SRSetMetrics.SRVoteGap.WithLabelValues(m.nodeLabel).Set(float64(gap))

	NodeMetrics.LastCheckTime.WithLabelValues(m.nodeLabel, "sr_borderline").Set(float64(time.Now().Unix()))
}
