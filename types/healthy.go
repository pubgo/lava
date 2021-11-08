package types

import "net/http"

type Healthy func(req *http.Request) error
