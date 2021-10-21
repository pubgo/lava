package task_entry

import (
	"fmt"
	"github.com/pubgo/lava/plugins/broker"

	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/task"
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
