package memberlist

import (
	"fmt"

	"github.com/hashicorp/memberlist"
)

type delegate struct{}

func (d delegate) NodeMeta(limit int) []byte {
	panic("implement me")
}

func (d delegate) NotifyMsg(bytes []byte) {
	panic("implement me")
}

func (d delegate) GetBroadcasts(overhead, limit int) [][]byte {
	panic("implement me")
}

func (d delegate) LocalState(join bool) []byte {
	panic("implement me")
}

func (d delegate) MergeRemoteState(buf []byte, join bool) {
	panic("implement me")
}

type eventDelegate struct{}

func (ed *eventDelegate) NotifyJoin(node *memberlist.Node) {
	fmt.Println("A node has joined: " + node.String())
}

func (ed *eventDelegate) NotifyLeave(node *memberlist.Node) {
	fmt.Println("A node has left: " + node.String())
}

func (ed *eventDelegate) NotifyUpdate(node *memberlist.Node) {
	fmt.Println("A node was updated: " + node.String())
}

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

func (b2 broadcast) Invalidates(b memberlist.Broadcast) bool {
	panic("implement me")
}

func (b2 broadcast) Message() []byte {
	panic("implement me")
}

func (b2 broadcast) Finished() {
	panic("implement me")
}
