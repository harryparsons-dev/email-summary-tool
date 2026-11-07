package servicetests

import (
	"email-summary-tool/services/background"
	"sync/atomic"
	"testing"
)

func TestTasksWaitBlocksUntilTasksFinish(t *testing.T) {
	tasks := background.NewTasks()
	var finished atomic.Int32

	for range 5 {
		tasks.Add()
		go func() {
			defer tasks.Done()
			finished.Add(1)
		}()
	}
	tasks.Wait()

	if got := finished.Load(); got != 5 {
		t.Fatalf("finished tasks = %d, want 5", got)
	}
}
