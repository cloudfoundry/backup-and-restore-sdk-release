package concurrently_test

import (
	"boshcompiler/internal/concurrently"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sync"
	"sync/atomic"
	"time"
)

var _ = Describe("Run", func() {
	It("runs all the work", func() {
		var (
			lock sync.Mutex
			done []string
		)

		work := []string{"alpha", "beta", "gamma", "delta"}

		concurrently.Run(10, work, nil, func(id string) {
			lock.Lock()
			defer lock.Unlock()
			done = append(done, id)
		})

		Expect(done).To(ConsistOf("alpha", "beta", "gamma", "delta"))
	})

	It("does not run work until the dependencies are finished", func() {
		const (
			dependent  = "dependent"
			dependency = "dependency"
		)

		var done atomic.Bool

		work := []string{dependent, dependency}
		deps := map[string][]string{
			dependent:  {dependency},
			dependency: {},
		}
		concurrently.Run(5, work, deps, func(id string) {
			defer GinkgoRecover()

			switch id {
			case dependent:
				if !done.Load() {
					Fail("dependent running before dependency")
				}
			case dependency:
				time.Sleep(100 * time.Millisecond)
				done.Store(true)
			}
		})
	})
})
