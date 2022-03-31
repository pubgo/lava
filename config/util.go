package config

import (
	"github.com/pubgo/lava/pkg/merge"
)

type CfgMap map[string]interface{}

func (t CfgMap) Decode(dst interface{}, opts ...func(cfg *DecoderConfig)) error {
	return merge.MapStruct(dst, t, opts...)
}
