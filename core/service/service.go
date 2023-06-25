package service

import (
	"strconv"
	"strings"
)

type Service struct {
	Name      string            `json:"name,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Endpoints []string          `json:"endpoints,omitempty"`
	Nodes     []*Node           `json:"nodes,omitempty"`
}

type (
	Nodes []*Node
	Node  struct {
		Id       string            `json:"id,omitempty"`
		Version  string            `json:"version,omitempty"`
		Address  string            `json:"address,omitempty"`
		Port     int               `json:"port,omitempty"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}
)

func (t Node) GetPort() int {
	if t.Port != 0 {
		return t.Port
	}

	adders := strings.Split(t.Address, ":")
	if len(adders) > 1 {
		port, _ := strconv.Atoi(adders[len(adders)-1])
		return port
	}

	return 0
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

func addPort(addr string) int {
	adders := strings.Split(addr, ":")
	if len(adders) > 1 {
		port, _ := strconv.Atoi(adders[len(adders)-1])
		return port
	}

	return 0
}

// Instance is an instance of a service in a discovery system.
type Instance struct {
	// ID is the unique instance ID as registered.
	ID string `json:"id"`
	// Name is the service name as registered.
	Name string `json:"name"`
	// Version is the version of the compiled.
	Version string `json:"version"`
	// Metadata is the kv pair metadata associated with the service instance.
	Metadata map[string]string `json:"metadata"`
	// Endpoints are endpoint addresses of the service instance.
	// schema:
	//   http://127.0.0.1:8000?isSecure=false
	//   grpc://127.0.0.1:9000?isSecure=false
	Endpoints []string `json:"endpoints"`
}
