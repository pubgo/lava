package supervisor

import (
	"context"
	"fmt"
)

func (m *Manager) OnClose(fn func()) {
	_ = m.Add(&onCloseService{fn: fn})
}

type onCloseService struct {
	fn func()
}

func (s *onCloseService) Name() string    { return "on-close-" + fmt.Sprintf("%p", s.fn) }
func (s *onCloseService) Error() error    { return nil }
func (s *onCloseService) String() string  { return s.Name() }
func (s *onCloseService) Metric() *Metric { return &Metric{Name: s.Name()} }
func (s *onCloseService) Serve(ctx context.Context) error {
	<-ctx.Done()
	s.fn()
	return NoRestartErr(nil)
}
