package debug

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/templates"
	_ "github.com/go-echarts/statsview/expvar"
	"github.com/go-echarts/statsview/statics"
	"github.com/go-echarts/statsview/viewer"
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

var page1 *components.Page
var smgr1 *viewer.StatsMgr

func AddView(view viewer.Viewer) {
	view.SetStatsMgr(smgr1)
	page1.AddCharts(view.View())
	HandleFunc("/debug/statsview/view/"+view.Name(), view.Serve)
}

func InitView() {
	viewer.SetConfiguration(viewer.WithTheme(viewer.ThemeWesteros), viewer.WithAddr("localhost:8081"))
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
	page1 = page
	page.PageTitle = "statsview"
	page.AssetsHost = fmt.Sprintf("http://%s/debug/statsview/statics/", viewer.LinkAddr())
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
	smgr1 = smgr
	for _, v := range mgr.Views {
		v.SetStatsMgr(smgr)
	}

	for _, v := range mgr.Views {
		page.AddCharts(v.View())
		HandleFunc("/debug/statsview/view/"+v.Name(), v.Serve)
	}

	http.HandleFunc("/debug/statsview", func(w http.ResponseWriter, _ *http.Request) {
		page.Render(w)
	})

	staticsPrev := "/debug/statsview/statics/"
	HandleFunc(staticsPrev+"echarts.min.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(statics.EchartJS))
	})

	HandleFunc(staticsPrev+"jquery.min.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(statics.JqueryJS))
	})

	HandleFunc(staticsPrev+"themes/westeros.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(statics.WesterosJS))
	})

	HandleFunc(staticsPrev+"themes/macarons.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(statics.MacaronsJS))
	})

	return mgr
}
