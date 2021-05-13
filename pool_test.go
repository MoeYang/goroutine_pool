package goroutine_pool

import (
	"sync"
	"testing"
	"time"
)

const (
	runTimes = 1000000
)

func f() {
	time.Sleep(10 * time.Millisecond)
}

func BenchmarkPool(b *testing.B) {
	pool := NewPool(WithCapacity(30000), WithBlock(true))
	defer pool.Close()
	var wg sync.WaitGroup
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runTimes)
		for j := 0; j < runTimes; j++ {
			_ = pool.RunTask(func() {
				f()
				wg.Done()
			})
		}
		wg.Wait()
	}
	b.StopTimer()
}

func BenchmarkGoroutine(b *testing.B) {
	var wg sync.WaitGroup
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runTimes)
		for j := 0; j < runTimes; j++ {
			go func() {
				f()
				wg.Done()
			}()
		}
		wg.Wait()
	}
	b.StopTimer()
}

func BenchmarkPoolNoWait(b *testing.B) {
	pool := NewPool(WithCapacity(30000), WithBlock(true))
	defer pool.Close()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < runTimes; j++ {
			_ = pool.RunTask(f)
		}
	}
	b.StopTimer()
}

func BenchmarkGoroutineNoWait(b *testing.B) {
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < runTimes; j++ {
			go f()
		}
	}
	b.StopTimer()
}
