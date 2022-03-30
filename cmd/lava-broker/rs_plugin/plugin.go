package rs_plugin

import (
	"bufio"
	"bytes"
	"context"
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/cmd/lava-broker/rsocket/sockets"
	"github.com/pubgo/lava/cmd/lava-broker/rsocket/util"
	"github.com/pubgo/lava/core/logging/logutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/service/service_type"
	"github.com/rsocket/rsocket-go/core"
	"github.com/rsocket/rsocket-go/core/transport"
	"go.uber.org/zap"
	"io"
)

type Server struct {
}

func Enable(srv service_type.Service) {
	// todo config
	// todo server
	var cfg = util.NewServer()
	var ctx, cancel = context.WithCancel(srv.Ctx())

	var ln = srv.RegisterMatcher(40, func(reader io.Reader) bool {
		br := bufio.NewReader(&io.LimitedReader{R: reader, N: 4096})
		l, part, err := br.ReadLine()
		if err != nil || part {
			logutil.LogOrErr(zap.L(), "ReadLine", func() error { return err })
			return false
		}

		// 用于rsocket匹配
		var frame = transport.NewLengthBasedFrameDecoder(bytes.NewBuffer(l))
		data, err := frame.Read()
		if err != nil {
			logutil.LogOrErr(zap.L(), "frame.Read", func() error { return err })
			return false
		}

		var header = core.ParseFrameHeader(data)
		if header.Type().String() == "UNKNOWN" {
			return false
		}

		if logutil.Enabled(zap.DebugLevel) {
			zap.L().Debug(header.String())
		}

		return true
	})

	srv.BeforeStarts(func() {
		syncx.GoSafe(func() {
			cfg.Build(func(opts *sockets.Handler) {
				for _, desc := range srv.ServiceDesc() {
					opts.RegisterService(desc.ServiceDesc, desc.Handler)
				}
			})
		})
	})

	srv.BeforeStops(cancel)
}

func Main() {
	var srv = lava.NewService("lava-broker", "desc")
	Enable(srv)

	lava.Run(srv)
}
