// Package concurrently is a dependency-aware helper for running concurrent jobs
package concurrently

import (
	"boshcompiler/internal/workstatus"
	"log"
	"strings"
	"time"
)

type callback func(string)

// Run will run the callback function `cb` (which must be thread safe) in parallel with the values specified in the
// `work` slice, using up to the specified number of workers. It is dependency-aware, so if a dependency map is specified,
// it will not run the callback function for a value until the callback functions for its dependents have finished.
func Run(workers int, work []string, deps map[string][]string, cb callback) {
	workQueue := make(chan string)
	status := workstatus.New(work)

	dependenciesFinished := func(e string) bool {
		for _, d := range deps[e] {
			if status.Get(d) != workstatus.Finished {
				return false
			}
		}
		return true
	}

	go func() {
		for status.NumInState(workstatus.Pending) != 0 {
			for _, e := range work {
				if status.Get(e) == workstatus.Pending && dependenciesFinished(e) {
					status.Set(e, workstatus.Queued)
					workQueue <- e
				}
			}
			time.Sleep(time.Second)
		}
		close(workQueue)
	}()

	for i := 0; i < workers; i++ {
		go func() {
			for e := range workQueue {
				status.Set(e, workstatus.Running)
				cb(e)
				status.Set(e, workstatus.Finished)
			}
		}()
	}

	reporter := time.NewTicker(time.Minute)
	go func() {
		for range reporter.C {
			r := status.InState(workstatus.Running)
			log.Printf("Still running (%d): %s\n", len(r), strings.Join(r, ", "))
		}
	}()

	for status.NumInState(workstatus.Finished) != len(work) {
		time.Sleep(time.Second)
	}
}
