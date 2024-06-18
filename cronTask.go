package croner

import (
	"context"
	"time"
)

type Task struct {
	calculateDuration func() time.Duration
	onStart           func()
	runner            RunnerHandler
	onTerminate       func()
	active            bool
	aWait             bool
}

// Stop the running Task
// func (t *Task) stop() {
// 	if t.active {
// 		t.cancel()
// 	}
// }

// Start the Task with blocking-io
func (t *Task) start(ctx context.Context) {
	if t.active {
		panic("you want run a runned task")
	}

	t.active = true
	defer func() {
		t.active = false
	}()

	if t.onStart != nil {
		t.onStart()
	}

	if t.onTerminate != nil {
		defer t.onTerminate()
	}

	timer := time.NewTimer(t.calculateDuration())
	defer timer.Stop()

	for {
		timer.Reset(t.calculateDuration())
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			err := t.runner(ctx)
			if err == nil && t.aWait {
				return
			}

		}

	}
}
