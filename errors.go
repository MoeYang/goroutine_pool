package goroutine_pool

import "errors"

var (
	// ErrorPoolArriveCapacity would be return if submit a task to a pool who does not has more capacity to run a new goroutine.
	ErrorPoolArriveCapacity = errors.New("error cause of the pool is running to many goroutine")
	// ErrorPoolClosed would be return if submit a task to a closed pool.
	ErrorPoolClosed = errors.New("error cause of the pool is closed")
)
