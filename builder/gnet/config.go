package gnet

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pool/goroutine"
)

var _ = goroutine.Default()

type Cfg struct {
	srv interface{}
}

func (t *Cfg) Build() error {
	return nil
}

func (t *Cfg) Get() interface{} {
	if t.srv == nil {
		panic("please init gnet")
	}

	return t.srv
}

func init() {
	encoderConfig := gnet.EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               4,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}
	decoderConfig := gnet.DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    0,
		InitialBytesToStrip: 4,
	}

	codec := gnet.NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)
	log.Fatal(gnet.Serve(
		&gnet.EventServer{},
		fmt.Sprintf("tcp://:%d", 8089),
		gnet.WithMulticore(true),
		gnet.WithTCPKeepAlive(time.Minute*5), // todo 需要确定是否对长连接有影响
		gnet.WithCodec(codec),
	))
}
