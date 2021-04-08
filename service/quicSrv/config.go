package quicSrv

import (
	"net"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/logging"
	"github.com/lucas-clemente/quic-go/quictrace"
)

type Cfg struct {
	Versions                              []quic.VersionNumber
	ConnectionIDLength                    int
	HandshakeTimeout                      time.Duration
	MaxIdleTimeout                        time.Duration
	MaxReceiveStreamFlowControlWindow     uint64
	MaxReceiveConnectionFlowControlWindow uint64
	MaxIncomingStreams                    int64
	MaxIncomingUniStreams                 int64
	StatelessResetKey                     []byte
	KeepAlive                             bool
	TokenStore                            quic.TokenStore
	QuicTracer                            quictrace.Tracer
	Tracer                                logging.Tracer
	AcceptToken                           func(clientAddr net.Addr, token *quic.Token) bool
}

func (t Cfg) Build() (quic.Listener, error) {
	return nil, nil
}

func (t Cfg) ToQuicCfg() quic.Config {
	return quic.Config{}
}
