package croner

import (
	"context"
	"sync"
	"time"

	"github.com/adhocore/gronx"
)

/*
	Features:
		- Can start all tasks in none-blocking and blocking modes.
		- Assign Context to tasks to manage sub tasks lifecycle
		- Check whether service is stopped by channel in select mode
*/

type Service struct {
	mutex   sync.Mutex
	tasks   []*Task
	context context.Context
	cancel  context.CancelFunc
	C       chan struct{}
	Active  bool
}

// create a new Service
func NewService() *Service {
	return &Service{}
}

func (m *Service) resetContext() {
	if m.context != nil && m.context.Err() == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.context = ctx
	m.cancel = cancel
}

// stop activated service
func (m *Service) Clean() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.tasks = []*Task{}
}

// stop activated service
func (m *Service) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.Active {
		if m.cancel != nil {
			m.cancel()
		}
		close(m.C)
		m.Active = false
	}
}

// Start the Task without blocking-io
func (m *Service) StartTasksAwait() {
	m.resetContext()
	m.Active = true
	m.C = make(chan struct{})

	var wg sync.WaitGroup
	// m.mutex.Lock()
	wg.Add(len(m.tasks))
	for _, t := range m.tasks {
		f := func(t *Task) {
			t.start(m.context)
			wg.Done()
		}
		if t.aWait {
			f(t)
		} else {
			go f(t)
		}
	}
	// m.mutex.Unlock()
	wg.Wait()
}
func (m *Service) NewTask(d Duration, handlers ...Handlers) (*Task, error) {
	return m.newTask(d, false, handlers...)
}

// func (m *Service) NewTaskAwait(d Duration, handlers ...Handlers) (*Task, error) {
// 	return m.newTask(d, true, handlers...)
// }

// Create a new Task with the given duration and Handlers
func (m *Service) newTask(d Duration, aWait bool, handlers ...Handlers) (*Task, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var c = &Task{
		aWait: aWait,
	}

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

		c.calculateDuration = func() time.Duration {
			now := time.Now()
			ttt, _ := gronx.NextTick(*d.crontab, false)
			dur := ttt.Sub(now)
			return dur
		}
	} else if d.duration != nil {
		c.calculateDuration = func() time.Duration {
			return *d.duration
		}
	}
	m.tasks = append(m.tasks, c)
	return c, nil
}

// run f function in once, if f out the error then retry to run f again.
// func (m *Service) RetryAwait(f func(context.Context) error, retryDelayDuration time.Duration) error {
// 	m.mutex.Lock()
// 	defer m.mutex.Unlock()
// 	m.resetContext()

// 	timer := time.NewTimer(retryDelayDuration)
// 	for {
// 		timer.Reset(retryDelayDuration)
// 		select {
// 		case <-m.context.Done():
// 			return m.context.Err()
// 		case <-timer.C:
// 			if err := f(m.context); err != nil {
// 				continue
// 			}
// 			return nil
// 		}
// 	}
// }

// ========================= Options:

type Duration struct {
	duration *time.Duration
	crontab  *string
}

type RunType int

const (
	RunContinuousAsync RunType = 0
	RunOnceSync        RunType = 1
)

type RunnerType struct {
	runnerType RunType
}

func WithRunType(t RunType) RunnerType {
	return RunnerType{runnerType: t}
}

type RunnerHandler func(ctx context.Context) error

func WithCrontab(c string) Duration {
	return Duration{crontab: &c}
}
func WithDuration(d time.Duration) Duration {
	return Duration{duration: &d}
}

type Handlers struct {
	runner      RunnerHandler
	onStart     func()
	onTerminate func()
}

func Runner(f RunnerHandler) Handlers {
	return Handlers{runner: f}
}
func OnStart(f func()) Handlers {
	return Handlers{onStart: f}
}
func OnTerminate(f func()) Handlers {
	return Handlers{onTerminate: f}
}
