package batch

import (
	"context"
	"sync"
	"time"
)

// BatchConfig defines batching behavior.
type BatchConfig struct {
	BatchSize    int           // Max requests in one batch
	BatchTimeout time.Duration // Max time to wait before flushing
	MaxPending   int           // Max pending batches
}

// BatchProcessor processes batches of items.
type BatchProcessor struct {
	config    BatchConfig
	processor func([]interface{}) error
	queue     chan interface{}
	batches   chan []interface{}
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.Mutex
	stats     BatchStats
}

// BatchStats tracks batching statistics.
type BatchStats struct {
	ProcessedBatches int64
	ProcessedItems   int64
	FailedBatches    int64
	FailedItems      int64
	AvgBatchSize     float64
}

// NewBatchProcessor creates a new batch processor.
func NewBatchProcessor(config BatchConfig, processor func([]interface{}) error) *BatchProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	return &BatchProcessor{
		config:    config,
		processor: processor,
		queue:     make(chan interface{}, config.MaxPending*10),
		batches:   make(chan []interface{}, 10),
		ctx:       ctx,
		cancel:    cancel,
		wg:        sync.WaitGroup{},
		stats:     BatchStats{},
	}
}

// Start begins processing.
func (bp *BatchProcessor) Start() {
	bp.wg.Add(1)
	go bp.batchWorker()
	go bp.flushWorker()
}

// Stop gracefully shuts down processor.
func (bp *BatchProcessor) Stop() {
	bp.cancel()
	bp.wg.Wait()
}

// Add adds an item to be processed.
func (bp *BatchProcessor) Add(item interface{}) {
	select {
	case bp.queue <- item:
	case <-bp.ctx.Done():
	}
}

// batchWorker collects items into batches.
func (bp *BatchProcessor) batchWorker() {
	defer bp.wg.Done()

	currentBatch := make([]interface{}, 0, bp.config.BatchSize)
	batchTimer := time.NewTimer(bp.config.BatchTimeout)

	for {
		select {
		case item := <-bp.queue:
			currentBatch = append(currentBatch, item)

			if len(currentBatch) >= bp.config.BatchSize {
				batchTimer.Stop()
				bp.mu.Lock()
				bp.batches <- currentBatch
				currentBatch = make([]interface{}, 0, bp.config.BatchSize)
				bp.mu.Unlock()
				batchTimer.Reset(bp.config.BatchTimeout)
			}
		case <-batchTimer.C:
			if len(currentBatch) > 0 {
				batchTimer.Stop()
				bp.mu.Lock()
				bp.batches <- currentBatch
				currentBatch = make([]interface{}, 0, bp.config.BatchSize)
				bp.mu.Unlock()
				batchTimer.Reset(bp.config.BatchTimeout)
			}
		case <-bp.ctx.Done():
			if len(currentBatch) > 0 {
				bp.batches <- currentBatch
			}
			return
		}
	}
}

// flushWorker processes batches.
func (bp *BatchProcessor) flushWorker() {
	defer bp.wg.Done()

	for {
		select {
		case batch := <-bp.batches:
			if err := bp.processor(batch); err != nil {
				bp.mu.Lock()
				bp.stats.FailedBatches++
				bp.stats.FailedItems += int64(len(batch))
				bp.mu.Unlock()
			} else {
				bp.mu.Lock()
				bp.stats.ProcessedBatches++
				bp.stats.ProcessedItems += int64(len(batch))
				totalBatches := bp.stats.ProcessedBatches + bp.stats.FailedBatches
				if totalBatches > 0 {
					bp.stats.AvgBatchSize = float64(bp.stats.ProcessedItems) / float64(totalBatches)
				}
				bp.mu.Unlock()
			}
		case <-bp.ctx.Done():
			return
		}
	}
}

// GetStats returns current statistics.
func (bp *BatchProcessor) GetStats() BatchStats {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.stats
}
