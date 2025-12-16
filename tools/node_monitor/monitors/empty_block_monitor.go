package monitors

import (
	"context"
	"fmt"
	"log"
	"time"
)

// EmptyBlockMonitor monitors for empty blocks
type EmptyBlockMonitor struct {
	client       *TronClient
	interval     time.Duration
	lastBlockNum int64
	nodeLabel    string
}

// NewEmptyBlockMonitor creates a new empty block monitor
func NewEmptyBlockMonitor(client *TronClient, interval time.Duration, nodeLabel string) *EmptyBlockMonitor {
	return &EmptyBlockMonitor{
		client:    client,
		interval:  interval,
		nodeLabel: nodeLabel,
	}
}

// Start starts the empty block monitoring loop
func (m *EmptyBlockMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	log.Println("Empty block monitor started")

	// Do an initial check
	m.checkBlock(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Empty block monitor stopped")
			return
		case <-ticker.C:
			m.checkBlock(ctx)
		}
	}
}

// checkBlock checks the latest block for emptiness
func (m *EmptyBlockMonitor) checkBlock(ctx context.Context) {
	block, err := m.client.GetNowBlock()
	if err != nil {
		log.Printf("Failed to get latest block: %v", err)
		NodeMetrics.APICallErrors.WithLabelValues(m.nodeLabel, "getnowblock", "request_failed").Inc()
		return
	}

	blockNum := block.BlockHeader.RawData.Number
	witnessAddress := block.BlockHeader.RawData.WitnessAddress

	// Update last block metrics
	EmptyBlockMetrics.LastBlockNumber.WithLabelValues(m.nodeLabel).Set(float64(blockNum))
	EmptyBlockMetrics.LastBlockTimestamp.WithLabelValues(m.nodeLabel).Set(float64(block.BlockHeader.RawData.Timestamp))

	// Check if this is a new block
	if blockNum <= m.lastBlockNum {
		return
	}

	// Check if block is empty (no transactions or only witness transactions)
	isEmpty := m.isBlockEmpty(block)

	if isEmpty {
		log.Printf("Empty block detected: block_num=%d, witness=%s", blockNum, witnessAddress)
		EmptyBlockMetrics.EmptyBlockCount.WithLabelValues(m.nodeLabel, witnessAddress).Inc()
		EmptyBlockMetrics.EmptyBlockDetected.WithLabelValues(
			m.nodeLabel,
			fmt.Sprintf("%d", blockNum),
			witnessAddress,
		).Set(1)

		// expose block hash for empty blocks
		blockHash := computeBlockHash(block)
		EmptyBlockMetrics.EmptyBlockInfo.WithLabelValues(
			m.nodeLabel,
			fmt.Sprintf("%d", blockNum),
			witnessAddress,
			blockHash,
		).Set(1)
	} else {
		// Reset the detection flag for this block
		EmptyBlockMetrics.EmptyBlockDetected.WithLabelValues(
			m.nodeLabel,
			fmt.Sprintf("%d", blockNum),
			witnessAddress,
		).Set(0)
	}

	m.lastBlockNum = blockNum
	NodeMetrics.LastCheckTime.WithLabelValues(m.nodeLabel, "empty_block").Set(float64(time.Now().Unix()))
}

// isBlockEmpty checks if a block is empty
// A block is considered empty if it has no transactions
func (m *EmptyBlockMonitor) isBlockEmpty(block *Block) bool {
	// A block is empty if it contains no transactions
	// Note: In Tron, even empty blocks might have witness transactions,
	// but for monitoring purposes, we consider a block empty if transactions array is empty
	return len(block.Transactions) == 0
}

// computeBlockHash returns the hash identifier of the block
// For now we simply use the BlockID field from the API response
func computeBlockHash(block *Block) string {
	return block.BlockID
}
