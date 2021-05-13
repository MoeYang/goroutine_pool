package goroutine_pool

import "sync/atomic"

const (
	defaultCap = 1024
	noLimitCap = -1 // if cap equals -1, will create goWorker per request with no limit.

	// pool state
	PoolOpen = iota
	PoolClose
)

type Pool interface {
	// RunTask submit a user-task to a goroutine,
	// and would return an error when creating a goWorker failed
	RunTask(func()) error
	// RunTaskWithRet submit a user-task which has a return value,
	// and would return a closed-chan that func`s result Value would be pushed in,
	// or a nil-chan if some error happend while creating a goWorker.
	RunTaskWithRet(func() interface{}) (<-chan interface{}, error)
	// Close Stop the pool and close all goroutines.
	Close()
	// IsClosed return whether the pool has been closed.
	IsClosed() bool
	// RunningCount return workers count which are running.
	RunningCount() int32
}

// goPool is a implement of interface Pool
type goPool struct {
	// workerPool can get a goWorker to run task
	workers workerStore
	// state is whether the pool is open
	state int32

	cap     int32
	isBlock bool
}

// NewPool create a goroutine goPool
func NewPool(opts ...Option) Pool {
	p := &goPool{state: PoolOpen}
	for _, opt := range opts {
		opt(p)
	}
	p.workers = newWorkerStore(p.cap, p.isBlock, p)

	return p
}

func (p *goPool) RunTask(f func()) error {
	// if pool was closed, return err
	if p.IsClosed() {
		return ErrorPoolClosed
	}
	worker := p.workers.getWorker()
	// has no goWorker in pool, return err
	if worker == nil {
		return ErrorPoolArriveCapacity
	}
	// run user-task
	worker.submitTask(f)
	return nil
}

func (p *goPool) RunTaskWithRet(f func() interface{}) (<-chan interface{}, error) {
	// if pool was closed, return err
	if p.IsClosed() {
		return nil, ErrorPoolClosed
	}
	worker := p.workers.getWorker()
	// has no goWorker in pool, return err
	if worker == nil {
		return nil, ErrorPoolArriveCapacity
	}
	// run user-task
	var taskWithRet struct {
		f     func()
		retCh chan interface{}
	}
	taskWithRet.retCh = make(chan interface{}, 1)
	taskWithRet.f = func() {
		taskWithRet.retCh <- f()
	}
	worker.submitTask(taskWithRet.f)
	return taskWithRet.retCh, nil
}

func (p *goPool) Close() {
	// cas to promised only close once
	if atomic.CompareAndSwapInt32(&p.state, PoolOpen, PoolClose) {
		p.workers.close()
	}
}

// IsClosed return whether the pool has been closed.
func (p *goPool) IsClosed() bool {
	return atomic.LoadInt32(&p.state) == PoolClose
}

func (p *goPool) RunningCount() int32 {
	return p.workers.getRunningCount()
}

type Option func(*goPool)

// WithShardCount set max capacity
func WithCapacity(cap int32) Option {
	return func(p *goPool) {
		if cap <= 0 && cap != noLimitCap {
			cap = defaultCap
		}
		p.cap = cap
	}
}

// WithBlock set whether get a worker need block
func WithBlock(isBlock bool) Option {
	return func(p *goPool) {
		p.isBlock = isBlock
	}
}
