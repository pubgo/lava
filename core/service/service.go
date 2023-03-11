package service

import (
	"strconv"
	"strings"
	"time"
)

type Service struct {
	Name      string            `json:"name,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Endpoints []*Endpoint       `json:"endpoints,omitempty"`
	Nodes     []*Node           `json:"nodes,omitempty"`
}

type Nodes []*Node
type Node struct {
	Id       string            `json:"id,omitempty"`
	Version  string            `json:"version,omitempty"`
	Address  string            `json:"address,omitempty"`
	Port     int               `json:"port,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

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

// ServiceInstance todo
type ServiceInstance struct {
	Region       string      `json:"region,omitempty"`
	InstanceName string      `json:"instanceName,omitempty"`
	ServiceName  string      `json:"serviceName,omitempty"`
	Type         ServiceType `json:"serviceType,omitempty"`
	Address      string      `json:"address,omitempty"`
	Version      string      `json:"version,omitempty"`
	GitBranch    string      `json:"gitBranch,omitempty"`
	GitCommit    string      `json:"gitCommit,omitempty"`
	BuildEnv     string      `json:"buildEnv,omitempty"`
	BuildAt      string      `json:"buildAt,omitempty"`
	Online       int64       `json:"online,omitempty"` // 毫秒时间戳

	Meta map[string]interface{} `json:"meta,omitempty"`

	Prefix   string        `json:"-"`
	Interval time.Duration `json:"-"`
	TTL      int64         `json:"-"`
}

// ServiceType 服务类型
type ServiceType string

const (
	// API 提供API访问的服务
	API = ServiceType("api")
	// Worker 后台作业服务
	Worker = ServiceType("worker")
)
