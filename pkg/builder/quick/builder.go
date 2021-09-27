package quick

import (
	"github.com/lucas-clemente/quic-go"
)

type Builder struct {
	srv quic.Listener
}

func (t *Builder) Get() quic.Listener {
	if t.srv == nil {
		panic("please init chi")
	}

	return t.srv
}

func (t *Builder) Build(cfg Cfg) error {
	//t.srv, err = quic.ListenAddr(addr, tlsConf, t.ToCfg())
	return nil
}

func New() Builder {
	return Builder{}
}
