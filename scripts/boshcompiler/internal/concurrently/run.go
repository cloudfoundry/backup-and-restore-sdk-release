// Package concurrently is a dependency-aware helper for running concurrent jobs
package concurrently

import (
	"boshcompiler/internal/workstatus"
	"log"
	"strings"
	"time"
)

type callback func(string)

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
