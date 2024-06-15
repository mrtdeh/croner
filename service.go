package croner

import (
	"context"
	"fmt"
	"sync"
	"time"
)

/*
	Features:
		- Can start all tasks in none-blocking and blocking modes.
		- Assign Context to tasks to manage sub tasks lifecycle
		- Check whether service is stopped by channel in select mode
*/

type Service struct {
	tasks  []*Task
	C      chan struct{}
	Active bool
}

// create a new Service
func NewService() *Service {
	return &Service{
		C: make(chan struct{}),
	}
}

// stop activated service
func (m *Service) Stop() {
	if m.Active {
		for _, t := range m.tasks {
			t.stop()
		}
		close(m.C)
		m.Active = false
	}
}

// Start the Task without blocking-io
func (m *Service) StartWait() {
	m.Active = true

	var wg sync.WaitGroup
	wg.Add(len(m.tasks))
	for _, t := range m.tasks {
		go func() {
			t.start()
			wg.Done()
		}()
	}
	wg.Wait()
}

// Create a new Task with the given duration and Handlers
func (m *Service) NewTask(d Duration, handlers ...Handlers) (*Task, error) {

	var c = &Task{}

	for _, h := range handlers {
		if h.runner != nil {
			c.runner = h.runner
		}
		if h.onStart != nil {
			c.onStart = h.onStart
		}
		if h.onTerminate != nil {
			c.onTerminate = h.onTerminate
		}
	}
	if c.runner == nil {
		panic("runner is nil")
	}

	if d.crontab != nil {
		i, err := parseCrontab(*d.crontab)
		if err != nil {
			return nil, fmt.Errorf("parse crontab failed : %v", err)
		}
		c.duration = time.Duration(i) * time.Minute
	} else if d.duration != nil {
		c.duration = *d.duration
	}
	m.tasks = append(m.tasks, c)
	return c, nil
}

// ========================= Options:

type Duration struct {
	duration *time.Duration
	crontab  *string
}

func WithCrontab(c string) Duration {
	return Duration{crontab: &c}
}
func WithDuration(d time.Duration) Duration {
	return Duration{duration: &d}
}

type Handlers struct {
	runner      func(context.Context)
	onStart     func()
	onTerminate func()
}

func Runner(f func(ctx context.Context)) Handlers {
	return Handlers{runner: f}
}
func OnStart(f func()) Handlers {
	return Handlers{onStart: f}
}
func OnTerminate(f func()) Handlers {
	return Handlers{onTerminate: f}
}
