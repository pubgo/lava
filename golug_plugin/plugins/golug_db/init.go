package golug_db

import (
	"sync"
)

var name = "db"
var cfg Cfg
var clientM sync.Map
