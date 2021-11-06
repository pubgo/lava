package hello

import (
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/pubgo/lava/pkg/clix"
	"github.com/pubgo/lava/pkg/shutil"
)

var Cmd = clix.Command(func(cmd *cobra.Command, flags *pflag.FlagSet) {
	cmd.Use = "hello"
	cmd.Short = "hello"
	cmd.Run = func(cmd *cobra.Command, args []string) {
		// Components routing:
		app.Route("/", &hello{})
		app.Route("/hello", &hello{})
		app.RunWhenOnBrowser()

		// HTTP routing:
		http.Handle("/", &app.Handler{
			Name:        "Hello",
			Description: "An Hello World! example",
		})

		shutil.GoModGraph()
		http.ListenAndServe(":8088", nil)
	}
})
