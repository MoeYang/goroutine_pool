package goroutine_pool

// worker running in a loop to run user define`s tasks.
type worker interface {
	// submitTask submit a user-task to goWorker
	submitTask(task)
	// stop the goWorker`s loop, and it will not put back to workerStore
	stop()
}

// goWorker is a implement of interface worker.
type goWorker struct {
	// submit user-task by chan
	taskCh chan task
	// workerStore which the goWorker belongs to
	store workerStore
}

// newWorker create worker which running in a goroutine.
func newWorker(store workerStore) worker {
	w := &goWorker{
		taskCh: make(chan task),
		store:  store,
	}
	w.run()
	return w
}

// submitTask submit a user-task to goWorker
func (w *goWorker) submitTask(t task) {
	w.taskCh <- t
}

// stop the goWorker`s loop, and it will not put back to workerStore
func (w *goWorker) stop() {
	close(w.taskCh)
}

// run a loop to deal user-task, it will be called only once when create a goWorker
func (w *goWorker) run() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				// if panic, put a newWorker in pool
				w.store.putWorker(newWorker(w.store))
				// TODO how to deal the panic info?
			}
		}()
		for t := range w.taskCh {
			// call user task
			t()
			// user-task done, put goWorker back to workerStore to reuse it.
			w.store.putWorker(w)
		}
	}()
}

// task is what workers would call in goroutine
type task func()
