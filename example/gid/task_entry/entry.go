package task_entry

import (
	"fmt"
	"github.com/pubgo/lava/entry/task"

	"github.com/pubgo/lava/abc/broker"
	"github.com/pubgo/lava/entry"
)

var name = "test-task"

func GetEntry() entry.Entry {
	ent := task.New(name)
	ent.Description("entry task test")

	ent.Register("topic", func(msg *broker.Message) error {
		fmt.Println(*msg)
		return nil
	}, nil)
	return ent
}
