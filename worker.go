package cruiseprocessor

import (
	"context"
)

const (
	defaultPoolSize  = 1000
	defaultWorkerNum = 2
)

type placeHolder struct{}
type PlaceHolder placeHolder
type task func() error
type Result interface{}

type WorkerPool struct {
	pool      chan task
	workerNum int

	closeCh chan placeHolder
	errCh   chan error
	resCh   chan Result
}

func NewWorkerPool(workNum int) *WorkerPool {
	return &WorkerPool{
		workerNum: workNum,
		pool:      make(chan task, defaultPoolSize),
		closeCh:   make(chan placeHolder),
		errCh:     make(chan error),
	}
}

func (wp *WorkerPool) Run(ctx context.Context) (err error) {
	// run all worker
	for i := 0; i < wp.workerNum; i++ {
		go wp.work()
	}

	select {
	case <-ctx.Done():
	case err = <-wp.errCh:
	}
	close(wp.closeCh)
	return
}

func (wp *WorkerPool) ResultChan() <-chan Result {
	return wp.resCh
}

func (wp *WorkerPool) Execute(t func() error) {
	wp.pool <- t
}

func (wp *WorkerPool) work() {
	for {
		select {
		case t := <-wp.pool:
			// TODO: consider how to wrap errors
			if err := t(); err != nil {
				wp.errCh <- err
			}

		case <-wp.closeCh:
			return
		}
	}
}
