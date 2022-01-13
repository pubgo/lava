package resource

import (
	"io"
)

// Resource 资源对象接口
// 	Resource 是真实对象的wrapper, 可以通过Unwrap获取真实的内部对象
type Resource interface {
	// Kind 资源类型
	Kind() string
	// Unwrap 获取真实的资源对象
	//	Unwrap 得到内部真实对象
	Unwrap() io.Closer
	// UpdateObj 更新资源
	UpdateObj(val Resource)
}
