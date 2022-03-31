package resource

import (
	"reflect"

	"github.com/pubgo/lava/resource/resource_type"
)

type IResource = resource_type.Resource

var resourceType = reflect.TypeOf((*IResource)(nil)).Elem()
