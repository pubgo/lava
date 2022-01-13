package endpoint

import (
	"errors"

	"github.com/pubgo/lava/version"
)

type Endpoint struct {
	//endpoint unique id
	Id string

	//endpoint name
	Name string

	//endpoint version
	Version string

	// schema/name/version/id
	Absolute string

	//service server ip address
	Address string

	// name.version
	// eg. api.hello/v.1.0.0
	Scope string
}

// NewEndpoint new a endpoint with schema,id,name,version,address
func NewEndpoint(id, name, address string) (*Endpoint, error) {
	if name == "" || address == "" || id == "" {
		return nil, errors.New("not complete")
	}
	e := new(Endpoint)
	e.Id = id
	e.Name = name
	e.Version = version.Version
	e.Address = address
	e.Scope = e.Name + "/" + e.Version
	e.Absolute = version.Domain + "/" + e.Scope + "/" + e.Address
	return e, nil
}
