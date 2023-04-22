package config

import "reflect"

var defaultCfg = make(map[reflect.Type]any)

func RegDefaultConfig(conf any) {
	defaultCfg[reflect.TypeOf(conf)] = conf
}

func ListDefaultConfig() map[reflect.Type]any {
	return defaultCfg
}
