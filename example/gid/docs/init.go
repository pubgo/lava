package docs

import (
	"github.com/pubgo/lug/plugins/swagger"
)

func init() {
	swagger.Init(AssetNames, MustAsset)
}
