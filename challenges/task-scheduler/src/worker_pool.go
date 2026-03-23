package taskscheduler

import (
	"context"
	"errors"
)

var (
	ErrPoolNotRunning = errors.New("worker pool is not running")
	ErrPoolFull       = errors.New("worker pool task queue is full")
)

type workItem struct {
	task    *Task
	handler TaskHandler
}

type WorkerPool struct {
	size int
}

func NewWorkerPool(size int) *WorkerPool {
	return nil
}

func (wp *WorkerPool) Start(ctx context.Context) {
}

func (wp *WorkerPool) Submit(task *Task, handler TaskHandler) error {
	return ErrPoolNotRunning
}

func (wp *WorkerPool) Shutdown(ctx context.Context) error {
	return nil
}

func (wp *WorkerPool) ActiveWorkers() int {
	return 0
}

func (wp *WorkerPool) ProcessedCount() int64 {
	return 0
}
