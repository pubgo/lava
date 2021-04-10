package task_entry

import (
	"fmt"

	"github.com/pubgo/lug"
	"github.com/pubgo/lug/broker"
	"github.com/pubgo/lug/entry"
)

var name = "test-task"

func GetEntry() entry.Abc {
	ent := lug.NewTask(name)
	ent.Version("v0.0.1")
	ent.Description("entry task test")

	ent.Register("topic", func(msg *broker.Message) error {
		fmt.Println(*msg)
		return nil
	}, nil)
	return ent
}
