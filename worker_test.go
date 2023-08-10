package cruiseprocessor

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func mockTasks(num int) []func() error {
	fns := []func() error{}
	for i := 0; i < num; i++ {
		fn := func() error {
			time.Sleep(time.Second)
			fmt.Printf("now: %+v\n", time.Now().String())
			return nil
		}
		fns = append(fns, fn)
	}
	return fns
}

func TestNormal(t *testing.T) {
	fns := mockTasks(10)
	for _, fn := range fns {
		fn()
	}
}

func TestWorkerPool(t *testing.T) {
	wp := NewWorkerPool(5)
	//Run
	go func() {
		wp.Run(context.TODO())
	}()
	//Waiting for the creation of thread pool
	time.Sleep(2 * time.Second)
	fns := mockTasks(10)
	for i, fn := range fns {
		wp.Execute(i%wp.workerNum, fn)
	}
	time.Sleep(time.Second)
}

func BenchmarkWorkerPool_Execute(b *testing.B) {
	wp := NewWorkerPool(5)
	go func() {
		wp.Run(context.TODO())
	}()
	//Waiting for the creation of thread pool
	time.Sleep(time.Second)
	fns := mockTasks(1000)
	for i, fn := range fns {
		wp.Execute(i%wp.workerNum, fn)
	}
	//time.Sleep(time.Second)
}
