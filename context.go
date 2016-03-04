package context

import (
	"sync"
	"time"

	gocontext "golang.org/x/net/context"
)

type TimerContext interface {
	gocontext.Context
	Reinit(d time.Duration)
}

type timerContext struct {
	lock           sync.RWMutex
	timer          *time.Timer
	deadline       time.Time
	conversionChan chan struct{}
	err            error
}

func New(d time.Duration) TimerContext {
	ctx := &timerContext{
		timer:          time.NewTimer(d),
		deadline:       time.Now().Add(d),
		conversionChan: make(chan struct{}),
	}
	go ctx.converter()
	return ctx
}

func (ctx *timerContext) converter() {
	ctx.lock.RLock()
	c := ctx.timer.C
	ctx.lock.RUnlock()
	<-c
	ctx.lock.Lock()
	ctx.err = gocontext.DeadlineExceeded
	close(ctx.conversionChan)
	ctx.lock.Unlock()
}

func (ctx *timerContext) Reinit(d time.Duration) {
	ctx.lock.Lock()
	if !ctx.timer.Reset(d) {
		ctx.conversionChan = make(chan struct{})
		go ctx.converter()
	}
	ctx.deadline = time.Now().Add(d)
	ctx.lock.Unlock()
}

func (ctx *timerContext) Deadline() (deadline time.Time, ok bool) {
	ctx.lock.RLock()
	d := ctx.deadline
	ctx.lock.RUnlock()
	return d, true
}

func (ctx *timerContext) Done() <-chan struct{} {
	ctx.lock.RLock()
	c := ctx.conversionChan
	ctx.lock.RUnlock()
	return c
}

func (ctx *timerContext) Err() error {
	ctx.lock.RLock()
	err := ctx.err
	ctx.lock.RUnlock()
	return err
}

func (ctx *timerContext) Value(key interface{}) interface{} {
	return nil
}
