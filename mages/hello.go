package mages

import "github.com/magefile/mage/mg"

// https://github.com/carolynvs/magex

type Ns mg.Namespace

// Init test namespace
func (Ns) Init() {
}

// Hello import测试
func Hello(a string, b string) {
	println("hello", a, b)
}
