package types

import "github.com/pubgo/x/merge"

type CfgMap map[string]interface{}

func (t CfgMap) Decode(dst interface{}) error {
	return merge.MapStruct(dst, &t)
}


type M map[string]interface{}