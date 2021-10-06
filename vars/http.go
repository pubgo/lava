package vars

import (
	"expvar"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pubgo/lug/mux"
	"github.com/pubgo/xerror"

	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	h "github.com/maragudk/gomponents/html"
)

func init() {
	var index = func(keys []string) g.Node {
		var nodes []g.Node
		nodes = append(nodes, h.H1(g.Text("/debug/expvar")))
		nodes = append(nodes, h.A(g.Text("/debug"), g.Attr("href", "/debug")), h.Br())
		for i := range keys {
			nodes = append(nodes, h.A(g.Text(keys[i]), g.Attr("href", keys[i])), h.Br())
		}
		return c.HTML5(c.HTML5Props{
			Title:    "/debug/expvar",
			Language: "en",
			Body:     nodes,
		})
	}

	mux.Get("/debug/expvar/{name}", func(w http.ResponseWriter, request *http.Request) {
		var name = chi.URLParam(request, "name")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintln(w, expvar.Get(name).String())
	})
	mux.Get("/debug/expvar", func(w http.ResponseWriter, request *http.Request) {
		var keys []string
		expvar.Do(func(kv expvar.KeyValue) {
			keys = append(keys, fmt.Sprintf("/debug/expvar/%s", kv.Key))
		})
		xerror.Panic(index(keys).Render(w))
	})
}