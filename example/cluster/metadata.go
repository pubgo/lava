package cluster

import (
	"encoding/json"
)

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
