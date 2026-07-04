package protojson

import (
	"github.com/pubgo/lava/v2/core/encoding"
	pkgprotojson "github.com/pubgo/lava/v2/pkg/encoding/protojson"
)

func init() {
	encoding.Register(Name, Default)
}

const Name = pkgprotojson.Name

var Default = pkgprotojson.Default

var UseNumber = pkgprotojson.UseNumber

type Codec = pkgprotojson.Codec
