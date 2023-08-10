package cruiseprocessor

import (
	"context"
)

const (
	defaultPoolSize = 1000
	//defaultPoolLen==defaultWorkerNum
	defaultWorkerNum = 2
)

type placeHolder struct{}
type PlaceHolder placeHolder
type task func() error
type Result interface{}

type WorkerPool struct {
	poolList  []chan task
	workerNum int

	closeCh chan placeHolder
	errCh   chan error
	resCh   chan Result
}

func NewWorkerPool(workNum int) *WorkerPool {
	return &WorkerPool{
		workerNum: workNum,
		poolList:  make([]chan task, workNum),
		closeCh:   make(chan placeHolder),
		errCh:     make(chan error),
	}
}

func (wp *WorkerPool) Run(ctx context.Context) (err error) {
	// run all worker
	for i := 0; i < wp.workerNum; i++ {
		//Create thread corresponding channel
		wp.poolList[i] = make(chan task, defaultPoolSize)
		go wp.work(wp.poolList[i])
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

func (wp *WorkerPool) Execute(chanId int, t func() error) {
	// TODO: Matching scheduling algorithm (chanId int)
	//Corresponding data in the pipeline
	wp.poolList[chanId] <- t
}

func (wp *WorkerPool) work(pool chan task) {
	for {
		select {
		//Obtain data from the corresponding pipeline
		case t := <-pool:
			// TODO: consider how to wrap errors
			if err := t(); err != nil {
				wp.errCh <- err
			}

		case <-wp.closeCh:
			return
		}
	}
}
