package monitors

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// EmptyBlockMetrics contains metrics related to empty block monitoring
	EmptyBlockMetrics = struct {
		EmptyBlockCount    *prometheus.CounterVec
		EmptyBlockDetected *prometheus.GaugeVec
		EmptyBlockInfo     *prometheus.GaugeVec
		LastBlockNumber    *prometheus.GaugeVec
		LastBlockTimestamp *prometheus.GaugeVec
	}{
		EmptyBlockCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tron_empty_blocks_total",
				Help: "Total number of empty blocks detected",
			},
			[]string{"node", "witness_address"},
		),
		EmptyBlockDetected: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tron_empty_block_detected",
				Help: "Whether an empty block was detected (1 if yes, 0 if no) for the latest block",
			},
			[]string{"node", "block_number", "witness_address"},
		),
		EmptyBlockInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tron_empty_block_info",
				Help: "Info about empty blocks including block hash",
			},
			[]string{"node", "block_number", "witness_address", "block_hash"},
		),
		LastBlockNumber: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tron_last_block_number",
				Help: "The number of the last checked block",
			},
			[]string{"node"},
		),
		LastBlockTimestamp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tron_last_block_timestamp",
				Help: "The timestamp of the last checked block",
			},
			[]string{"node"},
		),
	}

	// SRSetMetrics contains metrics related to SR set monitoring
	SRSetMetrics = struct {
		SRSetChanges    *prometheus.CounterVec
		SRCount         *prometheus.GaugeVec
		SRSetHash       *prometheus.GaugeVec
		SREnabled       *prometheus.GaugeVec
		SRTotalProduced *prometheus.GaugeVec
		SRTotalMissed   *prometheus.GaugeVec
	}{
		SRSetChanges: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tron_sr_set_changes_total",
				Help: "Total number of SR set changes detected",
			},
			[]string{"node", "change_type"}, // change_type: "added", "removed", "reordered"
		),
		SRCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tron_sr_count",
				Help: "Current number of Super Representatives",
			},
			[]string{"node"},
		),
		SRSetHash: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tron_sr_set_hash",
				Help: "Hash of the current SR set (for detecting changes)",
			},
			[]string{"node"},
		),
		SREnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tron_sr_enabled",
				Help: "Whether a Super Representative is enabled (1 if yes, 0 if no)",
			},
			[]string{"node", "address", "url"},
		),
		SRTotalProduced: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tron_sr_total_produced",
				Help: "Total blocks produced by a Super Representative",
			},
			[]string{"node", "address"},
		),
		SRTotalMissed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tron_sr_total_missed",
				Help: "Total blocks missed by a Super Representative",
			},
			[]string{"node", "address"},
		),
	}

	// NodeMetrics contains general node metrics
	NodeMetrics = struct {
		APICallErrors *prometheus.CounterVec
		LastCheckTime *prometheus.GaugeVec
	}{
		APICallErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tron_monitor_api_errors_total",
				Help: "Total number of API call errors",
			},
			[]string{"node", "api_endpoint", "error_type"},
		),
		LastCheckTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tron_monitor_last_check_timestamp",
				Help: "Timestamp of the last successful check",
			},
			[]string{"node", "monitor_type"}, // monitor_type: "empty_block", "sr_set"
		),
	}
)

// RegisterAllMetrics registers all metrics with the provided registry
func RegisterAllMetrics(registry *prometheus.Registry) {
	registry.MustRegister(
		EmptyBlockMetrics.EmptyBlockCount,
		EmptyBlockMetrics.EmptyBlockDetected,
		EmptyBlockMetrics.EmptyBlockInfo,
		EmptyBlockMetrics.LastBlockNumber,
		EmptyBlockMetrics.LastBlockTimestamp,
		SRSetMetrics.SRSetChanges,
		SRSetMetrics.SRCount,
		SRSetMetrics.SRSetHash,
		SRSetMetrics.SREnabled,
		SRSetMetrics.SRTotalProduced,
		SRSetMetrics.SRTotalMissed,
		NodeMetrics.APICallErrors,
		NodeMetrics.LastCheckTime,
	)
}
