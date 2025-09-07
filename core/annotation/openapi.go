package annotation

import "github.com/pubgo/lava/v2/lava"

var _ lava.Annotation = (*Openapi)(nil)

type Openapi struct {
	ServiceName string `json:"service_name"`
}

func (s Openapi) Name() string {
	return "Openapi"
}
