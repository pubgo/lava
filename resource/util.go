package resource

import (
	"fmt"

	"github.com/pubgo/lava/consts"
)

const _resIdKey = "_id"

func GetResId(m map[string]interface{}) string {
	if m == nil {
		return consts.KeyDefault
	}

	var val, ok = m[_resIdKey]
	if !ok || val == nil {
		return consts.KeyDefault
	}

	return fmt.Sprintf("%v", val)
}
