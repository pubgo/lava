package task_entry

import (
	"fmt"
	broker2 "github.com/pubgo/lug/abc/broker"

	"github.com/pubgo/lug"
	"github.com/pubgo/lug/entry"
)

var name = "test-task"

func GetEntry() entry.Entry {
	ent := lug.NewTask(name)
	ent.Description("entry task test")

	ent.Register("topic", func(msg *broker2.Message) error {
		fmt.Println(*msg)
		return nil
	}, nil)
	return ent
}
