package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// TaskFunc defines the logic for processing a single item
type TaskFunc[T any] func(ctx context.Context, item T) error

// ListFunc defines how to fetch the batch of work
type ListFunc[T any] func(ctx context.Context) ([]T, error)

// ErrorFunc handles errors that occur during task execution
type ErrorFunc[T any] func(item T, err error)

// MaintenanceRunner executes scheduled maintenance tasks with concurrent processing.
//
// It retrieves a batch of items at regular intervals and processes them concurrently
// using a semaphore pattern to limit the number of goroutines.
//
// Graceful Shutdown:
//   - Context cancellation stops the ticker loop immediately
//   - In-flight tasks are allowed to complete via WaitGroup synchronization
//   - No new tasks are started after context cancellation
type MaintenanceRunner[T any] struct {
	interval    time.Duration
	concurrency int
	retriever   ListFunc[T]
	executor    TaskFunc[T]
	onError     ErrorFunc[T]
	logger      *slog.Logger
}

// NewMaintenanceRunner creates a new maintenance runner with the specified configuration.
//
// Parameters:
//   - interval: How often to run the maintenance task
//   - concurrency: Maximum number of concurrent workers
//   - retriever: Callback to fetch the list of items to process
//   - executor: Callback to execute the task for each item
//   - onError: Optional callback for handling executor errors (can be nil)
//   - logger: Logger for diagnostic output (can be nil)
func NewMaintenanceRunner[T any](
	interval time.Duration,
	concurrency int,
	retriever ListFunc[T],
	executor TaskFunc[T],
	onError ErrorFunc[T],
	logger *slog.Logger,
) *MaintenanceRunner[T] {
	return &MaintenanceRunner[T]{
		interval:    interval,
		concurrency: concurrency,
		retriever:   retriever,
		executor:    executor,
		onError:     onError,
		logger:      logger,
	}
}

// Start begins the maintenance runner. It will execute the retriever callback
// at the specified interval and process results concurrently until the context is cancelled.
func (r *MaintenanceRunner[T]) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.run(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// run fetches items via the retriever and processes them concurrently with a limit.
func (r *MaintenanceRunner[T]) run(ctx context.Context) {
	items, err := r.retriever(ctx)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("retriever failed", "err", err)
		}
		return
	}

	if len(items) == 0 {
		return
	}

	if r.logger != nil {
		r.logger.Debug("processing items", "count", len(items))
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, r.concurrency)

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{}

		go func(it T) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := r.executor(ctx, it); err != nil && r.onError != nil {
				r.onError(it, err)
			}
		}(item)
	}
	wg.Wait()
}
