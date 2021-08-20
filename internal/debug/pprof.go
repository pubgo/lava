package debug

import (
	"github.com/felixge/fgprof"
	"github.com/go-chi/chi/v5"

	"bytes"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"net/url"
	rpp "runtime/pprof"
	"sort"
	"strings"
)

func init() {
	On(func(app *chi.Mux) {
		app.HandleFunc("/debug/fgprof", fgprof.Handler().ServeHTTP)
		app.HandleFunc("/debug/pprof/", pprofHandle)
		app.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		app.HandleFunc("/debug/pprof/profile", pprof.Profile)
		app.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		app.HandleFunc("/debug/pprof/trace", pprof.Trace)
		app.HandleFunc("/debug/pprof/allocs", pprof.Handler("allocs").ServeHTTP)
		app.HandleFunc("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
		app.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
		app.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
		app.HandleFunc("/debug/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
		app.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	})
}

func pprofHandle(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/debug/pprof/") {
		name := strings.TrimPrefix(r.URL.Path, "/debug/pprof/")
		if name != "" {
			pprof.Handler(name).ServeHTTP(w, r)
			return
		}
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var profiles []profileEntry
	for _, p := range rpp.Profiles() {
		profiles = append(profiles, profileEntry{
			Name:  p.Name(),
			Href:  p.Name(),
			Desc:  profileDescriptions[p.Name()],
			Count: p.Count(),
		})
	}

	// Adding other profiles exposed from within this package
	for _, p := range []string{"cmdline", "profile", "trace"} {
		profiles = append(profiles, profileEntry{
			Name: p,
			Href: p,
			Desc: profileDescriptions[p],
		})
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	if err := indexTmplExecute(w, profiles); err != nil {
		log.Print(err)
	}
}

func indexTmplExecute(w io.Writer, profiles []profileEntry) error {
	var b bytes.Buffer
	b.WriteString(`<html>
<head>
<title>/debug/pprof/</title>
<style>
.profile-name{
	display:inline-block;
	width:6rem;
}
</style>
</head>
<body>
/debug/pprof/<br>
<br>
Types of profiles available:
<table>
<thead><td>Count</td><td>Profile</td></thead>
`)

	for _, profile := range profiles {
		link := &url.URL{Path: profile.Href, RawQuery: "debug=1"}
		fmt.Fprintf(&b, "<tr><td>%d</td><td><a href='%s'>%s</a></td></tr>\n", profile.Count, link, html.EscapeString(profile.Name))
	}

	b.WriteString(`</table>
<a href="goroutine?debug=2">full goroutine stack dump</a>
<br/>
<p>
Profile Descriptions:
<ul>
`)
	for _, profile := range profiles {
		fmt.Fprintf(&b, "<li><div class=profile-name>%s: </div> %s</li>\n", html.EscapeString(profile.Name), html.EscapeString(profile.Desc))
	}
	b.WriteString(`</ul>
</p>
</body>
</html>`)

	_, err := w.Write(b.Bytes())
	return err
}

type profileEntry struct {
	Name  string
	Href  string
	Desc  string
	Count int
}

var profileDescriptions = map[string]string{
	"allocs":       "A sampling of all past memory allocations",
	"block":        "Stack traces that led to blocking on synchronization primitives",
	"cmdline":      "The command line invocation of the current program",
	"goroutine":    "Stack traces of all current goroutines",
	"heap":         "A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.",
	"mutex":        "Stack traces of holders of contended mutexes",
	"profile":      "CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.",
	"threadcreate": "Stack traces that led to the creation of new OS threads",
	"trace":        "A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.",
}
