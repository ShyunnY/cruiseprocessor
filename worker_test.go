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
	go func() {
		wp.Run(context.TODO())
	}()
	fns := mockTasks(10)
	for _, fn := range fns {
		wp.Execute(fn)
	}
	time.Sleep(time.Minute)
}
