package config_type

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/merge"
)

type CfgMap map[string]interface{}

func (t CfgMap) Decode(dst interface{}, opts ...func(cfg *mapstructure.DecoderConfig)) (err error) {
	defer xerror.RespErr(&err)
	merge.MapStruct(dst, t, opts...)
	return
}
