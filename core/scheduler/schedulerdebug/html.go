package schedulerdebug

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/samber/lo"
	. "maragu.dev/gomponents" //nolint
	_ "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html" //nolint

	"github.com/pubgo/lava/v2/core/scheduler"
)

type NodeFn func() Node

func (f NodeFn) Render(w io.Writer) error {
	return f().Render(w)
}

func ListSchedulers(schedulers []*scheduler.Job) Node {
	return Div(
		Class("overflow-x-auto"),
		Table(
			Class("table"),
			THead(
				Tr(
					Th(Text("Name")),
					Th(Text("Status")),
					Th(Text("PreRun")),
					Th(Text("ExecTime")),
					Th(Text("Error")),
					Th(Text("Result")),
					Th(Text("RunCount")),
					Th(Text("Spec")),
				),
			),
			TBody(
				Map(schedulers, func(s *scheduler.Job) Node {
					return Tr(
						Th(Text(s.Spec.Name)),
						Th(Text(string(s.Status))),
						Th(Textf("%v", s.PreExecTime)),
						Th(Textf("%v", s.ExecTime)),
						Th(NodeFn(func() Node {
							if s.Error != nil {
								return Text(s.Error.Error())
							} else {
								return Text("null")
							}
						})),
						Th(Text(string(s.Result))),
						Th(Textf("%v", s.Runs)),
						Th(NodeFn(func() Node {
							modeId := hex.EncodeToString([]byte(s.Spec.Name))
							return Group{
								Script(Rawf(`function model%s(params) {document.getElementById('%s').showModal()}`, modeId, modeId)),
								Button(
									Class("btn"),
									Attr("onclick", fmt.Sprintf("model%s()", modeId)),
									Text("open modal"),
								),
								Dialog(
									ID(modeId),
									Class("modal"),
									Div(
										Class("modal-box"),
										H3(
											Class("text-lg font-bold"),
											Text("Hello!"),
										),
										P(
											Class("py-4"),
											Textf("%s", lo.Must(json.MarshalIndent(s.Spec, "  ", "  "))),
										),
										Div(
											Class("modal-action"),
											Form(
												Method("dialog"),
												Button(
													Class("btn"),
													Text("Close"),
												),
											),
										),
									),
								),
							}
						})),
					)
				}),
			),
		),
	)
}

const timeOnly = "15:04:05"

func Page(now time.Time, schedulers []*scheduler.Job) Node {
	return Doctype(
		HTML(
			Head(
				Meta(Charset("utf-8")),
				Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
				// Meta(Attr("http-equiv", "refresh"), Attr("content", "5")),
				TitleEl(Text("scheduler")),
				Group{
					Link(Href("https://cdn.jsdelivr.net/npm/daisyui@5"), Rel("stylesheet"), Type("text/css")),
					Script(Src("https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4")),
					Script(Src("https://cdn.tailwindcss.com?plugins=forms,typography")),
					Script(Src("https://unpkg.com/htmx.org")),
				},
			),
			Body(Group{
				Div(
					H1(Text(`Scheduler Table`)),
					partial(now),
				),

				ListSchedulers(schedulers),
			}),
		),
	)
}

func partial(now time.Time) Node {
	return P(ID("partial"), Textf(`Time was last updated at %v.`, now.Format(timeOnly)))
}
