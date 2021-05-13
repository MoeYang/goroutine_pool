package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/MoeYang/goroutine_pool"
)

func main() {

	defer pprof.StopCPUProfile()
	f, err := os.Create("cpuprofile")
	if err != nil {
		return
	}
	if err := pprof.StartCPUProfile(f); err != nil { //监控cpu
		log.Fatal("could not start CPU profile: ", err)
	}

	pool := goroutine_pool.NewPool(

		// set goroutine max Capacity.
		// if set -1, will create goroutine per request with no limit.
		goroutine_pool.WithCapacity(20000),

		// set whether get a worker need block, default false
		// if set true, pool.RunTask() would be blocked while no idle goroutine is in pool.
		goroutine_pool.WithBlock(true),
	)
	// close pool`s workers
	defer pool.Close()

	// run a task with no return value
	err = pool.RunTask(func() {
		time.Sleep(10 * time.Millisecond)
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	// run a task with a return value
	retCh, err := pool.RunTaskWithRet(func() interface{} {
		time.Sleep(10 * time.Millisecond)
		return "retString"
	})
	if err == nil {
		// read chan would be blocked until the task has been finished by any worker.
		// what`s more, the chan has been closed by worker after put return-Value in it.
		fmt.Println(<-retCh)
	} else {
		log.Fatal(err.Error())
	}

	// get pool`s close state
	isClosed := pool.IsClosed()
	fmt.Println(isClosed)

	// get how many workers are running user-task now.
	cnt := pool.RunningCount()
	fmt.Println(cnt)

}
