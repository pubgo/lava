package jump

import (
	"sync"

	"go.uber.org/atomic"
)

const (
	// TopWeight is the top weight that one entry might set.
	TopWeight = 100

	minReplicas = 100
	prime       = 16777619
)

type node struct {
	// 节点ip
	ip string

	// 节点权重
	weight int

	// 连接数
	conn atomic.Uint32
}

func (t *node) String() string {
	return t.ip
}

type (
	// Func defines the hash method.
	// 哈希函数
	Func func(data []byte) uint64

	// A ConsistentHash is a ring hash implementation.
	//一致性哈希
	ConsistentHash struct {
		//哈希函数
		hashFunc Func

		//虚拟节点放大因子
		//确定node的虚拟节点数量
		replicas int

		// 虚拟节点和物理映射
		nodes map[uint16]*node

		//读写锁
		lock sync.RWMutex
	}
)

// NewCustomConsistentHash returns a ConsistentHash with given replicas and hash func.
//有参构造器
func NewCustomConsistentHash(replicas int, fn Func) *ConsistentHash {
	if replicas < minReplicas {
		replicas = minReplicas
	}

	return &ConsistentHash{
		hashFunc: fn,
		replicas: replicas,
	}
}

// AddNode 添加节点
func AddNode(t *ConsistentHash, n *node) {
	// 根据node权重计算虚拟节点
	t.nodes[1] = n
}

// GetNode 获取节点
func GetNode(t *ConsistentHash, key string) *node {
	var bucket = HashString(key, t.replicas)
	return t.nodes[uint16(bucket)]
}

// 初始化
// 	虚拟节点初始化，默认节点数，整形map
//  虚拟1000个节点，a 20%, b 80%，那么，a 随机选择非重复的20%，剩余的给b，修改完毕之后，更改map值
// 添加节点
// 删除节点
// 虚拟节点
