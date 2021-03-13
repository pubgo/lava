package registry

import (
	"strconv"
	"strings"
)

type Service struct {
	Name      string            `json:"name,omitempty"`
	Version   string            `json:"version,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Endpoints []*Endpoint       `json:"endpoints,omitempty"`
	Nodes     []*Node           `json:"nodes,omitempty"`
}

func NodeOf(nodes ...*Node) []*Node {
	return nodes
}

type Node struct {
	Id       string            `json:"id,omitempty"`
	Address  string            `json:"address,omitempty"`
	Port     int               `json:"port,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (t Node) GetPort() int {
	if t.Port != 0 {
		return t.Port
	}

	parts := strings.Split(t.Address, ":")
	if len(parts) > 1 {
		t.Port, _ = strconv.Atoi(parts[len(parts)-1])
	}

	return t.Port
}

type Endpoint struct {
	Name     string            `json:"name,omitempty"`
	Request  *Value            `json:"request,omitempty"`
	Response *Value            `json:"response,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Value struct {
	Name   string   `json:"name,omitempty"`
	Type   string   `json:"type,omitempty"`
	Values []*Value `json:"values,omitempty"`
}
