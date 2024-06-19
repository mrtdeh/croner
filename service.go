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

func parseOptions(opts []Options) (opt Options) {
	for _, o := range opts {
		if o.crontab != nil {
			opt.crontab = o.crontab
		}
		if o.duration != nil {
			opt.duration = o.duration
		}
		if o.runner != nil {
			opt.runner = o.runner
		}

		if o.runnerType > 0 {
			opt.runnerType = o.runnerType
		}
	}
	if opt.runner == nil {
		panic("must provide a runner")
	}
	return opt
}

// Create a new Task with the given duration and Handlers
func (m *Service) NewTask(opts ...Options) (*Task, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	opt := parseOptions(opts)

	var c = &Task{}
	if opt.runnerType == RunOnceSync {
		c.aWait = true
	}
	c.runner = opt.runner

	if opt.crontab != nil {

		c.calculateDuration = func() time.Duration {
			now := time.Now()
			ttt, _ := gronx.NextTick(*opt.crontab, false)
			dur := ttt.Sub(now)
			return dur
		}
	} else if opt.duration != nil {
		c.calculateDuration = func() time.Duration {
			return *opt.duration
		}
	}
	m.tasks = append(m.tasks, c)
	return c, nil
}

// ========================= Options:

type RunType int

const (
	RunContinuousAsync RunType = 0
	RunOnceSync        RunType = 1
)

type Options struct {
	duration   *time.Duration
	crontab    *string
	runnerType RunType
	runner     RunnerHandler
}

func WithRunType(t RunType) Options {
	return Options{runnerType: t}
}

type RunnerHandler func(ctx context.Context) error

func WithCrontab(c string) Options {
	return Options{crontab: &c}
}
func WithDuration(d time.Duration) Options {
	return Options{duration: &d}
}

func WithRunner(f RunnerHandler) Options {
	return Options{runner: f}
}
