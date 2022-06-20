package menuservice

import "github.com/pubgo/dix"

func init() {
	dix.Provider(New)
}
