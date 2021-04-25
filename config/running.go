package config

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/xlog"
)

const (
	Dev runningMode = iota + 1
	Test
	Stag
	Prod
	Release
)

// runningMode 项目运行模式
type runningMode uint8

func (t runningMode) String() string {
	switch t {
	case 1:
		return "dev"
	case 2:
		return "test"
	case 3:
		return "stag"
	case 4:
		return "prod"
	case 5:
		return "release"
	default:
		xlog.Errorf("running mode(%d) not match", t)
		return consts.Unknown
	}
}

func IsDev() bool {
	return Mode == Dev.String()
}

func IsTest() bool {
	return Mode == Test.String()
}

func IsStag() bool {
	return Mode == Stag.String()
}

func IsProd() bool {
	return Mode == Prod.String()
}

func IsRelease() bool {
	return Mode == Release.String()
}
