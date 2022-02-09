package cluster

import (
	"github.com/hashicorp/memberlist"
)

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

func (*broadcast) UniqueBroadcast()                              {}
func (b *broadcast) Invalidates(other memberlist.Broadcast) bool { return false }
func (b *broadcast) Message() []byte {
	var msg = b.msg
	if l := len(msg); l > MaxPacketSize {
		logs.S().Infof("broadcast message size %d bigger then MaxPacketSize %d", l, MaxPacketSize)
	}
	return msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}
