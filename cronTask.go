package croner

import (
	"context"
	"time"
)

type Task struct {
	duration    time.Duration
	onStart     func()
	runner      func(context.Context)
	onTerminate func()
	cancel      context.CancelFunc
	active      bool
}

// Stop the running Task
func (t *Task) stop() {
	if t.active {
		t.cancel()
	}
}

// Start the Task with blocking-io
func (t *Task) start() {
	if t.active {
		panic("you want run a runned task")
	}

	t.active = true
	defer func() {
		t.active = false
	}()

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	if t.onStart != nil {
		t.onStart()
	}

	if t.onTerminate != nil {
		defer t.onTerminate()
	}

	timer := time.NewTimer(t.duration)
	defer timer.Stop()

	for {
		timer.Reset(t.duration)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			t.runner(ctx)
		}

	}
}
