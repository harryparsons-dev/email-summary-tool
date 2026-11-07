// Package background tracks work that handlers run in background goroutines.
package background

import "sync"

// Tasks counts running background tasks so callers can wait for them, for
// example tests that check the result of a background job.
type Tasks struct {
	wg sync.WaitGroup
}

func NewTasks() *Tasks {
	return &Tasks{}
}

// Add records that a background task is starting. Call it before the
// goroutine starts, and call Done when the task finishes.
func (t *Tasks) Add() {
	t.wg.Add(1)
}

// Done records that a background task has finished.
func (t *Tasks) Done() {
	t.wg.Done()
}

// Wait blocks until every background task has finished.
func (t *Tasks) Wait() {
	t.wg.Wait()
}
