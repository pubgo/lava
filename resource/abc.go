package resource

type Resource interface {
	// Close 关闭资源
	Close() error
	// UpdateResObj 更新资源, 资源对象不变
	UpdateResObj(val interface{})
	// Kind 资源种类
	Kind() string
}
