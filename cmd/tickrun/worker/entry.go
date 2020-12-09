package worker

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_entry"
)

var name = "tickrun_worker"

func GetEntry() golug_entry.Entry {
	ent := golug.NewCtlEntry(name)
	ent.Version("v0.0.1")
	ent.Description("worker handle")
	return ent
}
