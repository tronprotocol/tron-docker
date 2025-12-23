package monitors

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

// SRSetMonitor monitors changes in the Super Representative set
type SRSetMonitor struct {
	client          *TronClient
	interval        time.Duration
	lastSRSetHash   string
	lastSRAddresses []string
	nodeLabel       string
}

// NewSRSetMonitor creates a new SR set monitor
func NewSRSetMonitor(client *TronClient, interval time.Duration, nodeLabel string) *SRSetMonitor {
	return &SRSetMonitor{
		client:    client,
		interval:  interval,
		nodeLabel: nodeLabel,
	}
}

// Start starts the SR set monitoring loop
func (m *SRSetMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	log.Println("SR set monitor started")

	// Do an initial check to establish baseline
	m.checkSRSet(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("SR set monitor stopped")
			return
		case <-ticker.C:
			m.checkSRSet(ctx)
		}
	}
}

// checkSRSet checks for changes in the SR set
func (m *SRSetMonitor) checkSRSet(ctx context.Context) {
	witnesses, err := m.client.ListWitnesses()
	if err != nil {
		log.Printf("Failed to get witness list: %v", err)
		NodeMetrics.APICallErrors.WithLabelValues(m.nodeLabel, "listwitnesses", "request_failed").Inc()
		return
	}

	// Only consider the first 27 witnesses as active SRs
	const activeSRCount = 27
	var activeWitnesses []Witness
	if len(witnesses) > activeSRCount {
		activeWitnesses = witnesses[:activeSRCount]
	} else {
		activeWitnesses = witnesses
	}

	// Create list of addresses for active SRs only (preserve API order, which reflects ranking by voteCount)
	currentAddresses := make([]string, len(activeWitnesses))
	for i, w := range activeWitnesses {
		currentAddresses[i] = w.Address
	}

	// Compute hash of current active SR set
	currentHash := m.computeSRSetHash(activeWitnesses)

	// Update metrics with active SR count only
	SRSetMetrics.SRCount.WithLabelValues(m.nodeLabel).Set(float64(len(activeWitnesses)))
	SRSetMetrics.SRSetHash.WithLabelValues(m.nodeLabel).Set(float64(m.hashToFloat(currentHash)))

	// Update SR metrics for active SRs only
	for _, witness := range activeWitnesses {
		SRSetMetrics.SREnabled.WithLabelValues(m.nodeLabel, witness.Address, witness.URL).Set(1)
	}

	// Check for changes if we have a previous state
	if m.lastSRSetHash != "" && currentHash != m.lastSRSetHash {
		m.detectChanges(activeWitnesses, currentAddresses)
	}

	m.lastSRSetHash = currentHash
	m.lastSRAddresses = currentAddresses
	NodeMetrics.LastCheckTime.WithLabelValues(m.nodeLabel, "sr_set").Set(float64(time.Now().Unix()))
}

// detectChanges detects what changed in the SR set
func (m *SRSetMonitor) detectChanges(currentWitnesses []Witness, currentAddresses []string) {
	// Create maps for easier lookup
	lastSet := make(map[string]bool)
	for _, addr := range m.lastSRAddresses {
		lastSet[addr] = true
	}

	currentSet := make(map[string]bool)
	for _, addr := range currentAddresses {
		currentSet[addr] = true
	}

	// Detect added SRs
	for _, addr := range currentAddresses {
		if !lastSet[addr] {
			log.Printf("SR added: %s", addr)
			SRSetMetrics.SRSetChanges.WithLabelValues(m.nodeLabel, "added").Inc()
		}
	}

	// Detect removed SRs
	for _, addr := range m.lastSRAddresses {
		if !currentSet[addr] {
			log.Printf("SR removed: %s", addr)
			SRSetMetrics.SRSetChanges.WithLabelValues(m.nodeLabel, "removed").Inc()
		}
	}

	// TODO: track SR ranking reordering (within the common address set) separately in the future
	// Currently we only expose:
	// - added: SR entered the top-N active set
	// - removed: SR left the top-N active set
}

// computeSRSetHash computes a hash of the SR set
// The hash includes addresses and vote counts to detect any changes
func (m *SRSetMonitor) computeSRSetHash(witnesses []Witness) string {
	h := sha256.New()
	for _, w := range witnesses {
		h.Write([]byte(fmt.Sprintf("%s:%d", w.Address, w.VoteCount)))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// hashToFloat converts a hash string to a float64 for Prometheus metrics
// This is a simple conversion for the hash metric
func (m *SRSetMonitor) hashToFloat(hash string) float64 {
	// Take first 8 bytes and convert to float64
	if len(hash) >= 16 {
		val, err := hex.DecodeString(hash[:16])
		if err == nil && len(val) >= 8 {
			var result float64
			for i := 0; i < 8 && i < len(val); i++ {
				result = result*256 + float64(val[i])
			}
			return result
		}
	}
	return 0
}
