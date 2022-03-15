package cluster

import (
	"time"
	
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
)

type Node struct {
	// 节点名字
	Name string
	// 节点权重
	Weight uint16
	// 节点标签
	Tags map[string]string
	// 节点状态
	NodeState memberlist.NodeStateType
	// 节点创建时间
	CreateAt time.Time
	// 节点更新时间
	UpdateAt time.Time
	// 节点状态更新时间
	StatusLTime serf.LamportTime // lamport clock time of last received message
	// 节点退出时间
	LeaveTime time.Time // wall clock time of leave
}
