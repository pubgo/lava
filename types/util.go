package types

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/xerror"
)

type CfgMap map[string]interface{}

func (t CfgMap) Decode(dst interface{}, opts ...func(cfg *mapstructure.DecoderConfig)) (err error) {
	defer xerror.RespErr(&err)
	merge.MapStruct(dst, t, opts...)
	return
}

type M = map[string]interface{}
type L = []interface{}
