package ws_entry

import (
	"github.com/pubgo/golug/golug_entry"
)

type wsEntry struct {
	golug_entry.Entry
}

func newEntry(name string) *wsEntry {
	ent := &wsEntry{}
	return ent
}

func New(name string) *wsEntry {
	return newEntry(name)
}
