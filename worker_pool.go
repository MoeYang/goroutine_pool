package goroutine_pool

import (
	"sync"
	"sync/atomic"
)

// workerStore is a interface which you can implement it with different policys to store/reuse workers.
type workerStore interface {
	// getWorker return a idle goWorker running in a single goroutine if store has, or return nil if store reachs limit capacity.
	getWorker() worker
	// putWorker put back the goWorker so you can reuse it after.
	putWorker(worker)
	// close all workers
	close()
	// getRunningCount return workers count which are running.
	getRunningCount() int32

	incrRunningCount()

	decrRunningCount()
}

// newWorkerStore init a workerStore
func newWorkerStore(cap int32, isBlock bool, p Pool) workerStore {
	store := capacityStore{
		cap:     cap,
		pool:    p,
		isBlock: isBlock,
	}
	if isBlock {
		store.waitCh = make(chan struct{})
	}
	if cap != noLimitCap {
		store.workers = make([]worker, cap)
		for i := range store.workers {
			store.workers[i] = newWorker(&store)
		}
	}
	return &store
}

// capacityStore can create no more than cap workers
type capacityStore struct {
	cap          int32         // the capacity of the store
	runningCount int32         // the count of workers which borrowed out from store
	isBlock      bool          // if get a worker need block, default false
	workers      []worker      // idle workers
	pool         Pool          // which pool this store belongs to
	waitCh       chan struct{} // func getWorker() would blocked on this chan if there are no idle workers in pool.

	lock sync.Mutex
}

func (store *capacityStore) incrRunningCount() {
	atomic.AddInt32(&store.runningCount, 1)
}

func (store *capacityStore) decrRunningCount() {
	if store.isBlock {
		select {
		case store.waitCh <- struct{}{}:
		default:
		}
	}
	atomic.AddInt32(&store.runningCount, -1)
}

// getWorker return a idle goWorker if store has, or return nil if store reachs limit capacity.
func (store *capacityStore) getWorker() worker {
	if store.cap == noLimitCap {
		store.incrRunningCount()
		return newWorker(store)
	}
tryGet:
	store.lock.Lock()
	var w worker
	// if not closed and running workers count less than the store`s cap, can return a goWorker.
	if !store.pool.IsClosed() {
		if store.getRunningCount() < store.cap {
			// borrowed out a goWorker, incr the count of running workers.
			store.incrRunningCount()
			// has idle goWorker, return one
			lIdx := len(store.workers) - 1
			w = store.workers[lIdx]
			store.workers = store.workers[:lIdx]
		} else if store.isBlock {
			// block, wating for a worker finished work and been returned to pool
			store.lock.Unlock()
			<-store.waitCh
			goto tryGet
		}
	}
	store.lock.Unlock()
	return w
}

// putWorker return the goWorker to idle-goWorker-array
func (store *capacityStore) putWorker(w worker) {
	if store.cap == noLimitCap {
		store.decrRunningCount()
		return
	}
	// need first lock, so putWorker() can mutex with func close()
	store.lock.Lock()
	// if closed, no need to put worker back, stop it straightly.
	if store.pool.IsClosed() {
		w.stop()
		store.lock.Unlock()
		return
	}
	// put the goWorker back to idle-goWorker-array, and decrease the count of running workers.
	if int32(len(store.workers)) < store.cap {
		store.workers = append(store.workers, w)
	}
	store.decrRunningCount()
	store.lock.Unlock()
}

func (store *capacityStore) close() {
	store.lock.Lock()
	for i, w := range store.workers {
		w.stop()
		store.workers[i] = nil
	}
	store.workers = store.workers[:0]
	atomic.StoreInt32(&store.runningCount, 0)
	store.lock.Unlock()
}

func (store *capacityStore) getRunningCount() int32 {
	return atomic.LoadInt32(&store.runningCount)
}
