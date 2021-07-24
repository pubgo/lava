package rest_entry

import (
	_ "expvar"
	"net/http"

	"github.com/pubgo/lug"
	"github.com/pubgo/lug/entry"
)

var name = "test-http"

func GetEntry() entry.Entry {
	ent := lug.NewRest(name)
	ent.Description("entry http test")

	ent.BeforeStart(func() {
		go http.ListenAndServe(":8083", nil)
	})

	ent.Register(&Service{})
	return ent
}
