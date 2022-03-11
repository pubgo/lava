package main

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
)

type subscriber struct {
	errHandler func(err error)
	stack      string
	in         chan<- interface{}
}

var ErrClosed = errors.New("closed")

type pubSub struct {
	shutdown     atomic.Bool
	wg           *sync.WaitGroup
	logger       *zap.Logger
	topic        string
	counter      atomic.Int32
	subscribers  map[uintptr]*subscriber
	queue        chan interface{}
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func (t *pubSub) Sub(fn func(data <-chan interface{}), errHandlers ...func(err error)) {
	xerror.Assert(fn == nil, "[fn] should not be nil")

	var funcStack = stack.Func(fn)
	var errHandler = func(err error) {
		t.logger.Error("subscriber handler failed", logutil.ErrField(err, zap.String(logkey.Stack, funcStack))...)
	}

	if len(errHandlers) > 0 {
		errHandler = func(err error) {
			defer xerror.RespExit(funcStack)
			errHandlers[0](err)
		}
	}

	// 关闭检查
	if t.shutdown.Load() {
		errHandler(ErrClosed)
		return
	}

	var pipe = make(chan interface{})
	var pointer = reflect.ValueOf(pipe).Pointer()

	t.queue <- func() {
		// 订阅者初始化
		t.subscribers[pointer] = &subscriber{stack: funcStack, errHandler: errHandler, in: pipe}
	}

	t.wg.Add(1)
	go func() {
		defer func() {
			t.queue <- func() {
				// 订阅者关闭
				if t.subscribers[pointer] != nil {
					close(t.subscribers[pointer].in)
					t.subscribers[pointer] = nil
				}
			}

			var gErr error
			switch err := recover().(type) {
			case nil:
			case error:
				if err.(error) != ErrClosed {
					gErr = err.(error)
				}
			default:
				gErr = fmt.Errorf("%#v", err)
			}

			// 订阅者错误处理
			xerror.Exit(xerror.Try(func() { errHandler(gErr) }), funcStack)

			t.wg.Done()
		}()

		// 订阅者数据消费
		fn(pipe)
	}()
}

func (t *pubSub) loop() {
	defer t.wg.Done()

	for val := range t.queue {
		switch val.(type) {
		case func():
			val.(func())()
		default:
			for k, v := range t.subscribers {
				if v == nil {
					delete(t.subscribers, k)
					continue
				}

				select {
				case v.in <- val:
				// 超时
				case <-time.Tick(time.Millisecond * 10):
					v.errHandler(nil)
				}
			}
		}
	}
}

func (t *pubSub) Close() error {
	t.queue <- func() {
		// 关闭所有订阅者
		for k, v := range t.subscribers {
			delete(t.subscribers, k)

			if v != nil {
				close(v.in)
			}
		}
	}
	close(t.queue)
	t.wg.Wait()
	return nil
}

func (t *pubSub) Pub(val interface{}) error {
	if t.counter.Load() == 0 {
		zap.S().Warnf("topic(%s) has no subscriber", t.topic)
		return nil
	}

	// buf大小
	// 长度判断
	select {
	case t.queue <- val:
	case <-time.Tick(time.Second):
		return fmt.Errorf("write timeout")
	}

	return nil
}

func newPipeline(topic string) *pubSub {
	var wg sync.WaitGroup
	var p = &pubSub{
		readTimeout:  time.Millisecond * 10,
		writeTimeout: time.Millisecond * 10,
		wg:           &wg,
		topic:        topic,
		queue:        make(chan interface{}, 100),
		logger:       zap.L().Named("pubSub"),
		subscribers:  make(map[uintptr]*subscriber),
	}

	wg.Add(1)
	go p.loop()
	return p
}

func main() {
	var p = newPipeline("test")
	p.Sub(
		func(data <-chan interface{}) {
			for range data {

			}
		}, func(err error) {

		},
	)

	p.Pub(nil)
}
