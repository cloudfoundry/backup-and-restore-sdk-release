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

func (w *WorkStatus) Set(n string, s WorkState) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.data[n] = s
}

func (w *WorkStatus) Get(n string) WorkState {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.data[n]
}

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
