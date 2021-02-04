package task_entry

import (
	"fmt"

	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_broker"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
)

var name = "test-task"

func GetEntry() golug_entry.Entry {
	ent := golug.NewTask(name)
	ent.Version("v0.0.1")
	ent.Description("entry task test")

	xerror.Panic(ent.Register("topic", func(msg *golug_broker.Message) error {
		fmt.Println(*msg)
		return nil
	}))
	return ent
}
