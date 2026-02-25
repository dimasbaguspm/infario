package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/dimasbaguspm/infario/internal/platform/engine"
)

// DeploymentTask represents an async deployment processing task.
type DeploymentTask struct {
	ID          string
	ProjectID   string
	Hash        string
	FileReader  io.ReadCloser
	OnComplete  func(status string, err error) // Callback to update deployment status
}

// DeploymentWorkerPool manages async deployment processing with a fixed worker count.
type DeploymentWorkerPool struct {
	fileEngine *engine.FileEngine
	taskQueue  chan DeploymentTask
	wg         sync.WaitGroup
	stopCh     chan struct{}
}

// NewDeploymentWorkerPool creates a new deployment worker pool.
func NewDeploymentWorkerPool(fileEngine *engine.FileEngine, numWorkers int) *DeploymentWorkerPool {
	return &DeploymentWorkerPool{
		fileEngine: fileEngine,
		taskQueue:  make(chan DeploymentTask, numWorkers*2), // buffer for burst enqueuing
		stopCh:     make(chan struct{}),
	}
}

// Start begins the worker goroutines.
func (p *DeploymentWorkerPool) Start(ctx context.Context) {
	for i := 0; i < cap(p.taskQueue)/2; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

// Enqueue adds a task to the processing queue.
func (p *DeploymentWorkerPool) Enqueue(task DeploymentTask) {
	select {
	case p.taskQueue <- task:
	case <-p.stopCh:
		slog.Warn("worker pool stopped, task dropped", "deployment_id", task.ID)
	}
}

// Shutdown gracefully stops the worker pool and waits for pending tasks.
func (p *DeploymentWorkerPool) Shutdown(ctx context.Context) error {
	close(p.stopCh)

	// Use a goroutine to close the task queue once all workers finish
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(p.taskQueue)
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout")
	}
}

// worker processes tasks from the queue.
func (p *DeploymentWorkerPool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}
			p.processDeployment(ctx, task)
		case <-p.stopCh:
			return
		}
	}
}

// processDeployment handles the actual file extraction and status callback.
func (p *DeploymentWorkerPool) processDeployment(ctx context.Context, task DeploymentTask) {
	defer task.FileReader.Close()

	// Construct storage path: "deployments/{projectID}/{hash}"
	storagePath := fmt.Sprintf("deployments/%s/%s", task.ProjectID, task.Hash)

	slog.InfoContext(ctx, "processing deployment", "deployment_id", task.ID, "hash", task.Hash)

	// Store file to filesystem (extracts zip content)
	_, err := p.fileEngine.Store(ctx, storagePath, task.FileReader)
	if err != nil {
		slog.ErrorContext(ctx, "failed to store deployment", "deployment_id", task.ID, "error", err)
		task.OnComplete("error", fmt.Errorf("failed to store deployment: %w", err))
		return
	}

	slog.InfoContext(ctx, "deployment stored successfully", "deployment_id", task.ID)
	task.OnComplete("ready", nil)
}
