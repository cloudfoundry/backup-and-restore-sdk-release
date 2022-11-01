// Package workstatus is a concurrent map where the keys are strings
// and the values are WorkState types.
//
// We did consider using the Go sync.Map type, but because it does not (yet) support Generics,
// using it requires a lot of type casting which bulks up the code and makes it harder to reason about.
package workstatus

import (
	"sort"
	"sync"
)

type WorkState string

const (
	Pending  WorkState = "pending"
	Queued   WorkState = "queued"
	Running  WorkState = "running"
	Finished WorkState = "finished"
)

// New creates a new object, initialising it with the provided slice of strings which are all set to `Pending` state
func New(initial []string) *WorkStatus {
	data := make(map[string]WorkState)
	for _, e := range initial {
		data[e] = Pending
	}

	return &WorkStatus{data: data}
}

type WorkStatus struct {
	mutex sync.RWMutex
	data  map[string]WorkState
}

// Set adds or updates the provided key to the provided value
func (w *WorkStatus) Set(n string, s WorkState) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.data[n] = s
}

// Get returns the value of the provided key. If the key has not been set, an empty string is returned
func (w *WorkStatus) Get(n string) WorkState {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.data[n]
}

// InState returns a slice of keys whose values match the specified state
func (w *WorkStatus) InState(st WorkState) (result []string) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	for k, v := range w.data {
		if v == st {
			result = append(result, k)
		}
	}
	sort.Strings(result)
	return
}

// NumInState returns a count of keys whose values match the specified state
func (w *WorkStatus) NumInState(st WorkState) (result int) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	for _, v := range w.data {
		if v == st {
			result++
		}
	}
	return
}
