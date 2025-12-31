package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pubgo/funk/v2/log"
	"github.com/samber/lo"
	"github.com/valyala/fasthttp"
)

var port = flag.Int("port", 8080, "http port")

func main() {
	flag.Parse()

	wd := lo.Must1(os.Getwd())
	if len(os.Args) > 0 {
		wd = flag.Arg(1)
	}

	log.Info().Msgf("file dir: %s", wd)
	log.Info().Msgf("http://localhost:%v", lo.FromPtr(port))

	fs := &fasthttp.FS{
		Root:               wd,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: true,
		Compress:           false,
		AcceptByteRange:    true,
		// PathRewrite:     fasthttp.NewVHostPathRewriter(0),
	}

	s := &fasthttp.Server{
		Handler: fs.NewRequestHandler(),
		Logger:  log.NewStd(log.GetLogger("fileserver")),
	}

	fmt.Println(s.ListenAndServe(fmt.Sprintf(":%v", lo.FromPtr(port))))
}
