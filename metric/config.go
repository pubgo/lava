package metric

import (
	"github.com/pubgo/golug/consts"

	"net/http"
)

var (
	Name            = "metric"
	cfg             = consts.Default
	DefaultServeMux = &http.ServeMux{}
)
