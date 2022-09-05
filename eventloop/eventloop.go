package eventloop

import (
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/console"
	"github.com/rarnu/goscript/require"
	"sync"
	"time"
)

type job struct {
	cancelled bool
	fn        func()
}

type Timer struct {
	job
	timer *time.Timer
}

type Interval struct {
	job
	ticker   *time.Ticker
	stopChan chan struct{}
}

type EventLoop struct {
	vm            *goscript.Runtime
	jobChan       chan func()
	jobCount      int32
	canRun        bool
	auxJobs       []func()
	auxJobsLock   sync.Mutex
	wakeup        chan struct{}
	stopCond      *sync.Cond
	running       bool
	enableConsole bool
	registry      *require.Registry
}

func NewEventLoop(opts ...Option) *EventLoop {
	vm := goscript.New()

	loop := &EventLoop{
		vm:            vm,
		jobChan:       make(chan func()),
		wakeup:        make(chan struct{}, 1),
		stopCond:      sync.NewCond(&sync.Mutex{}),
		enableConsole: true,
	}

	for _, opt := range opts {
		opt(loop)
	}
	if loop.registry == nil {
		loop.registry = new(require.Registry)
	}
	loop.registry.Enable(vm)
	if loop.enableConsole {
		console.Enable(vm)
	}
	_ = vm.Set("setTimeout", loop.setTimeout)
	_ = vm.Set("setInterval", loop.setInterval)
	_ = vm.Set("clearTimeout", loop.clearTimeout)
	_ = vm.Set("clearInterval", loop.clearInterval)

	return loop
}

type Option func(*EventLoop)

// EnableConsole 控制 console 模块是否被加载到事件循环所使用的 Runtime 中
// 默认情况下，事件循环在创建时加载了 console 模块，将 EnableConsole(false) 传递给 NewEventLoop 可以禁止这种行为
func EnableConsole(enableConsole bool) Option {
	return func(loop *EventLoop) {
		loop.enableConsole = enableConsole
	}
}

func WithRegistry(registry *require.Registry) Option {
	return func(loop *EventLoop) {
		loop.registry = registry
	}
}

func (loop *EventLoop) schedule(call goscript.FunctionCall, repeating bool) goscript.Value {
	if fn, ok := goscript.AssertFunction(call.Argument(0)); ok {
		delay := call.Argument(1).ToInteger()
		var args []goscript.Value
		if len(call.Arguments) > 2 {
			args = call.Arguments[2:]
		}
		f := func() { _, _ = fn(nil, args...) }
		loop.jobCount++
		if repeating {
			return loop.vm.ToValue(loop.addInterval(f, time.Duration(delay)*time.Millisecond))
		} else {
			return loop.vm.ToValue(loop.addTimeout(f, time.Duration(delay)*time.Millisecond))
		}
	}
	return nil
}

func (loop *EventLoop) setTimeout(call goscript.FunctionCall) goscript.Value {
	return loop.schedule(call, false)
}

func (loop *EventLoop) setInterval(call goscript.FunctionCall) goscript.Value {
	return loop.schedule(call, true)
}

// SetTimeout 在指定的超时周期触发后，在事件循环的上下文中运行指定的函数
// SetTimeout 返回一个 Timer，可以传递给 ClearTimeoutSetTimeout
// 传递给函数的 Runtime 实例和任何由它派生的值都不能在函数之外使用
// SetTimeout 在事件循环内部或外部调用都是安全的
func (loop *EventLoop) SetTimeout(fn func(*goscript.Runtime), timeout time.Duration) *Timer {
	t := loop.addTimeout(func() { fn(loop.vm) }, timeout)
	loop.addAuxJob(func() {
		loop.jobCount++
	})
	return t
}

// ClearTimeout 取消由 SetTimeout 返回的，还没有运行的定时器
// ClearTimeout 在事件循环内部或外部调用都是安全的
func (loop *EventLoop) ClearTimeout(t *Timer) {
	loop.addAuxJob(func() {
		loop.clearTimeout(t)
	})
}

// SetInterval 在每个指定的超时时间段后，在事件循环的上下文中重复运行指定的函数
// SetInterval 返回一个 Interval，可以传递给 ClearInterval
// 传递给该函数的 Runtime 实例以及由它派生的任何值都不能在该函数之外使用
// SetInterval 在事件循环内部或外部调用都是安全的
func (loop *EventLoop) SetInterval(fn func(*goscript.Runtime), timeout time.Duration) *Interval {
	i := loop.addInterval(func() { fn(loop.vm) }, timeout)
	loop.addAuxJob(func() {
		loop.jobCount++
	})
	return i
}

// ClearInterval 取消了 SetInterval 返回的 Interval
// ClearInterval 在事件循环内部或外部调用都是安全的
func (loop *EventLoop) ClearInterval(i *Interval) {
	loop.addAuxJob(func() {
		loop.clearInterval(i)
	})
}

func (loop *EventLoop) setRunning() {
	loop.stopCond.L.Lock()
	if loop.running {
		panic("Loop is already started")
	}
	loop.running = true
	loop.stopCond.L.Unlock()
}

// Run 调用指定的函数，启动事件循环并等待，直到没有更多的延迟任务需要运行，之后停止循环并返回
// 传递给该函数的 Runtime 实例以及由其派生的任何值都不能在该函数之外使用
// 当事件循环已经运行时，不要使用这个函数，请使用 RunOnLoop() 代替
// 如果事件循环已经开始，调用它将产生 panic
func (loop *EventLoop) Run(fn func(*goscript.Runtime)) {
	loop.setRunning()
	fn(loop.vm)
	loop.run(false)
}

// Start 在后台启动事件循环，循环将持续运行，直到调用 Stop()
// 如果事件循环已经开始，调用它将产生 panic
func (loop *EventLoop) Start() {
	loop.setRunning()
	go loop.run(true)
}

// Stop 停止用 Start() 启动的循环。在这个函数返回后，循环内将不再有任何工作被执行
// 在这之后可以再次调用 Start() 或 Run() 来恢复执行
// 注意，它不会取消活动超时，不允许同时运行 Start() 和 Stop()
// 在一个已经停止的循环上或在循环内部调用 Stop() 会挂起
func (loop *EventLoop) Stop() {
	loop.jobChan <- func() {
		loop.canRun = false
	}

	loop.stopCond.L.Lock()
	for loop.running {
		loop.stopCond.Wait()
	}
	loop.stopCond.L.Unlock()
}

// RunOnLoop 将在循环的上下文中尽快运行指定的函数
// 运行的顺序被保留（即函数将按照调用 RunOnLoop() 的相同顺序被调用）
// 传递给函数的 Runtime 实例以及从它派生出来的任何值都不能在函数之外使用，在循环内部或外部调用都是安全的
func (loop *EventLoop) RunOnLoop(fn func(*goscript.Runtime)) {
	loop.addAuxJob(func() { fn(loop.vm) })
}

func (loop *EventLoop) runAux() {
	loop.auxJobsLock.Lock()
	jobs := loop.auxJobs
	loop.auxJobs = nil
	loop.auxJobsLock.Unlock()
	for _, job := range jobs {
		job()
	}
}

func (loop *EventLoop) run(inBackground bool) {
	loop.canRun = true
	loop.runAux()

	for loop.canRun && (inBackground || loop.jobCount > 0) {
		select {
		case job := <-loop.jobChan:
			job()
			if loop.canRun {
				select {
				case <-loop.wakeup:
					loop.runAux()
				default:
				}
			}
		case <-loop.wakeup:
			loop.runAux()
		}
	}
	loop.stopCond.L.Lock()
	loop.running = false
	loop.stopCond.L.Unlock()
	loop.stopCond.Broadcast()
}

func (loop *EventLoop) addAuxJob(fn func()) {
	loop.auxJobsLock.Lock()
	loop.auxJobs = append(loop.auxJobs, fn)
	loop.auxJobsLock.Unlock()
	select {
	case loop.wakeup <- struct{}{}:
	default:
	}
}

func (loop *EventLoop) addTimeout(f func(), timeout time.Duration) *Timer {
	t := &Timer{
		job: job{fn: f},
	}
	t.timer = time.AfterFunc(timeout, func() {
		loop.jobChan <- func() {
			loop.doTimeout(t)
		}
	})

	return t
}

func (loop *EventLoop) addInterval(f func(), timeout time.Duration) *Interval {
	if timeout <= 0 {
		timeout = time.Millisecond
	}

	i := &Interval{
		job:      job{fn: f},
		ticker:   time.NewTicker(timeout),
		stopChan: make(chan struct{}),
	}

	go i.run(loop)
	return i
}

func (loop *EventLoop) doTimeout(t *Timer) {
	if !t.cancelled {
		t.fn()
		t.cancelled = true
		loop.jobCount--
	}
}

func (loop *EventLoop) doInterval(i *Interval) {
	if !i.cancelled {
		i.fn()
	}
}

func (loop *EventLoop) clearTimeout(t *Timer) {
	if t != nil && !t.cancelled {
		t.timer.Stop()
		t.cancelled = true
		loop.jobCount--
	}
}

func (loop *EventLoop) clearInterval(i *Interval) {
	if i != nil && !i.cancelled {
		i.cancelled = true
		close(i.stopChan)
		loop.jobCount--
	}
}

func (i *Interval) run(loop *EventLoop) {
L:
	for {
		select {
		case <-i.stopChan:
			i.ticker.Stop()
			break L
		case <-i.ticker.C:
			loop.jobChan <- func() {
				loop.doInterval(i)
			}
		}
	}
}
