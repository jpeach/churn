package workload

import (
	"sync"
	"time"
)

// Task is a single unit of a workload.
type Task interface {
	Perform() error
	Finalize() error
	Interval() time.Duration
}

func runTask(stopChan <-chan struct{}, t Task) error {
	for {
		select {
		case <-stopChan:
			return nil
		case <-time.After(t.Interval()):
			if err := t.Perform(); err != nil {
				return err
			}
		}
	}

	t.Finalize()

	return nil
}

// Run runs a collection of Tasks until the stopChan becomes ready.
func Run(stopChan <-chan struct{}, tasks []Task) error {

	w := sync.WaitGroup{}

	for _, t := range tasks {
		t := t
		w.Add(1)

		go func() {
			runTask(stopChan, t)
			w.Done()
		}()
	}

	w.Wait()
	return nil
}
