package gossip

import "github.com/hashicorp/memberlist"

type event struct {
	action int32
	node   string
}

type eventDelegate struct {
	events chan *event
}

func (ed *eventDelegate) NotifyJoin(n *memberlist.Node) {
	ed.events <- &event{action: nodeActionJoin, node: n.Address()}
}
func (ed *eventDelegate) NotifyLeave(n *memberlist.Node) {
	ed.events <- &event{action: nodeActionLeave, node: n.Address()}
}
func (ed *eventDelegate) NotifyUpdate(n *memberlist.Node) {
	ed.events <- &event{action: nodeActionUpdate, node: n.Address()}
}
