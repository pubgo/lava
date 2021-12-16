package retry

import (
	"time"

	"github.com/pubgo/xerror"
)

type Retry func() Backoff

func (d Retry) Do(f func(i int) error) (err error) {
	var wrap = func(i int) (err error) {
		defer xerror.RespErr(&err)
		return f(i)
	}

	var b = d()
	for i := 0; ; i++ {
		if err = wrap(i); err == nil {
			return
		}

		dur, stop := b.Next()
		if stop {
			return
		}

		time.Sleep(dur)
	}
}

func (d Retry) DoVal(f func(i int) (interface{}, error)) (val interface{}, err error) {
	var wrap = func(i int) (val interface{}, err error) {
		defer xerror.RespErr(&err)
		return f(i)
	}

	var b = d()
	for i := 0; ; i++ {
		if val, err = wrap(i); err == nil {
			return
		}

		dur, stop := b.Next()
		if stop {
			return
		}

		time.Sleep(dur)
	}
}

func New(bs ...Backoff) Retry {
	var b = WithMaxRetries(3, NewConstant(DefaultConstant))
	if len(bs) > 0 {
		b = bs[0]
	}

	return func() Backoff { return b }
}

func Default() Retry {
	var b = WithMaxRetries(3, NewConstant(time.Millisecond*10))
	return func() Backoff { return b }
}
