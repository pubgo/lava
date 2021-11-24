package debug

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/templates"
	"github.com/go-echarts/statsview/statics"
	"github.com/go-echarts/statsview/viewer"

	"github.com/pubgo/lava/mux"
)

// ViewManager ...
type ViewManager struct {
	Smgr   *viewer.StatsMgr
	Ctx    context.Context
	Cancel context.CancelFunc
	Views  []viewer.Viewer
}

// Register registers views to the ViewManager
func (vm *ViewManager) Register(views ...viewer.Viewer) {
	vm.Views = append(vm.Views, views...)
}

func InitView() {
	viewer.SetConfiguration(viewer.WithTheme(viewer.ThemeWesteros), viewer.WithTemplate(`
$(function () { setInterval({{ .ViewID }}_sync, {{ .Interval }}); });
function {{ .ViewID }}_sync() {
    $.ajax({
        type: "GET",
        url: "/debug/statsview/view/{{ .Route }}",
        dataType: "json",
        success: function (result) {
            let opt = goecharts_{{ .ViewID }}.getOption();

            let x = opt.xAxis[0].data;
            x.push(result.time);
            if (x.length > {{ .MaxPoints }}) {
                x = x.slice(1);
            }
            opt.xAxis[0].data = x;

            for (let i = 0; i < result.values.length; i++) {
                let y = opt.series[i].data;
                y.push({ value: result.values[i] });
                if (y.length > {{ .MaxPoints }}) {
                    y = y.slice(1);
                }
                opt.series[i].data = y;

                goecharts_{{ .ViewID }}.setOption(opt);
            }
        }
    });
}`))
	_ = New()

	templates.PageTpl = `
{{- define "page" }}
<!DOCTYPE html>
<html>
    {{- template "header" . }}
<body>
<p>&nbsp;&nbsp;🚀 <a href="https://github.com/go-echarts/statsview"><b>StatsView</b></a> <em>is a real-time Golang runtime stats visualization profiler</em></p>
<style> .box { justify-content:center; display:flex; flex-wrap:wrap } </style>
<div class="box"> {{- range .Charts }} {{ template "base" . }} {{- end }} </div>
</body>
</html>
{{ end }}
`
}

// New creates a new ViewManager instance
func New() *ViewManager {
	page := components.NewPage()
	page.PageTitle = "statsview"
	page.AssetsHost = "/debug/statsview/statics/"
	page.Assets.JSAssets.Add("jquery.min.js")

	mgr := &ViewManager{}
	mgr.Ctx, mgr.Cancel = context.WithCancel(context.Background())
	mgr.Register(
		viewer.NewGoroutinesViewer(),
		viewer.NewHeapViewer(),
		viewer.NewStackViewer(),
		viewer.NewGCNumViewer(),
		viewer.NewGCSizeViewer(),
		viewer.NewGCCPUFractionViewer(),
	)

	smgr := viewer.NewStatsMgr(mgr.Ctx)
	for _, v := range mgr.Views {
		v.SetStatsMgr(smgr)
	}

	mux.Route("/debug/statsview", func(r chi.Router) {
		r.Get("/", func(writer http.ResponseWriter, request *http.Request) { page.Render(writer) })

		for _, v := range mgr.Views {
			page.AddCharts(v.View())
			r.Get("/view/"+v.Name(), v.Serve)
		}

		staticsPrev := "/statics/"
		r.Get(staticsPrev+"echarts.min.js", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(statics.EchartJS))
		})

		r.Get(staticsPrev+"jquery.min.js", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(statics.JqueryJS))
		})

		r.Get(staticsPrev+"themes/westeros.js", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(statics.WesterosJS))
		})

		r.Get(staticsPrev+"themes/macarons.js", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(statics.MacaronsJS))
		})

	})

	return mgr
}
