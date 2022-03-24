package cluster

import (
	"encoding/json"
	"github.com/hashicorp/memberlist"
	pb "github.com/pubgo/lava/core/registry/gossip/proto"
	"github.com/pubgo/lava/core/registry/registry_type"
	event2 "github.com/pubgo/lava/event"
)

type delegate struct {
	queue *memberlist.TransmitLimitedQueue
	msg   chan []byte
}

func (d *delegate) NodeMeta(limit int) []byte { return []byte{} }

func (d *delegate) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	go func() { d.msg <- b }()
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return d.queue.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	if !join {
		return []byte{}
	}

	// 本节点的状态信息
	syncCh := make(chan *registry_type.Service, 1)
	services := map[string][]*registry_type.Service{}

	d.updates <- &update{
		Update: &pb.Update{
			Action: int32(event2.EventType_UPDATE),
		},
		sync: syncCh,
	}

	for srv := range syncCh {
		services[srv.Name] = append(services[srv.Name], srv)
	}

	b, _ := json.Marshal(services)
	return b
}

// MergeRemoteState 合并其他节点的状态
func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}

	if !join {
		return
	}

	// 别的节点同步过来的节点信息
	var services map[string][]*registry_type.Service
	if err := json.Unmarshal(buf, &services); err != nil {
		return
	}

	for _, service := range services {
		for _, srv := range service {
			d.updates <- &update{
				Update:  &pb.Update{Action: actionTypeCreate},
				Service: srv,
				sync:    nil,
			}
		}
	}
}

type NodeMetadata struct {
	GrpcPort int `json:"grpc_port"`
	HttpPort int `json:"http_port"`
}

func NewNodeMetadata() *NodeMetadata {
	return &NodeMetadata{
		GrpcPort: 0,
		HttpPort: 0,
	}
}

func NewNodeMetadataWithBytes(data []byte) (*NodeMetadata, error) {
	metadata := NewNodeMetadata()
	if err := json.Unmarshal(data, metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

func (m *NodeMetadata) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return data, nil
}
