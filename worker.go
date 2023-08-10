package cruiseprocessor

import (
	"context"
)

const (
	defaultPoolSize  = 1000
	defaultWorkerNum = 2
	//defaultPoolLen==defaultWorkerNum
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
		//创建线程对应通道
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
	// TODO: 匹配调度算法 ==> chanId int
	//对应管道内的数据
	wp.poolList[chanId] <- t
}

func (wp *WorkerPool) work(pool chan task) {
	for {
		select {
		//取得对应管道内的数据
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
