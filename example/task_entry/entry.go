package task_entry

import (
	"fmt"

	"github.com/pubgo/golug"
	"github.com/pubgo/golug/broker"
	"github.com/pubgo/golug/entry"
)

var name = "test-task"

func GetEntry() entry.Entry {
	ent := golug.NewTask(name)
	ent.Version("v0.0.1")
	ent.Description("entry task test")

	ent.Register("topic", func(msg *broker.Message) error {
		fmt.Println(*msg)
		return nil
	}, nil)
	return ent
}
