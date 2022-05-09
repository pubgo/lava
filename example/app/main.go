package main

import (
	"github.com/pubgo/dix"
)

var _ error = (*Data)(nil)

type Data struct {
	hello string
}

func (d Data) Error() string {
	return d.hello
}

func init() {
	// data 参数可以是ptr，interface，struct，map和func
	dix.Provider(&Data{hello: "hello"})
	dix.ProviderNs("ns", &Data{hello: "hello"}) // 设置namespace, 默认为default

	dix.Provider(error(&Data{hello: "hello"}))
	dix.Provider(struct {
		D *Data `dix:""`
	}{D: &Data{hello: "hello"}})
	dix.Provider(map[string]interface{}{"hello": &Data{hello: "hello"}})

	// 如果provider为func，那么func的参数可以为ptr，interface，struct
	dix.Provider(func(data *Data) {})
	dix.Provider(func(data Data) {})
	dix.Provider(func(data error) {})

	// 依赖注入,data要注入的对象, 必须为ptr类型，ns为namespace, 默认为default
	type A struct {
		D *Data `dix:""` // 属性必须是可导出的，同时tag必须为dix
	}

	var a A        // 注入对象到结构体
	dix.Inject(&a) // 注入Data对象到a的D中

	var d *Data    // 注入对象到指针
	dix.Inject(&d) // 注入Data对象到d中

	var e error    // 注入对象到interface
	dix.Inject(&e) // 注入Data对象到e中

	// 查看依赖关系图
	dix.Graph()
}
