package taskcmd

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/dix"
	"github.com/pubgo/lava/servers/tasks"
	"github.com/urfave/cli/v3"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name: "test-task",
		Action: func(ctx context.Context, command *cli.Command) error {
			dix.Inject(di, tasks.New(new(service))).Run()
			return nil
		},
	}
}

type service struct {
	cancel context.CancelFunc
}

func (s *service) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Println("test task")
			time.Sleep(time.Second)
		}
	}
}

func (s *service) Stop() {
	s.cancel()
}

func (s *service) Run() {
	// TODO implement me
	panic("implement me")
}
