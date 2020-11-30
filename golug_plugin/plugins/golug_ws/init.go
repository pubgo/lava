package golug_ws

import (
	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"sync"
)

const name = "ws"

var log xlog.XLog
var cfg Cfg
var wsM sync.Map

func init() {
	xerror.Exit(golug_log.Watch(func(logs xlog.XLog) {
		log = logs.Named(name)
	}))
}

//defer conn.Close()
//conn.SetReadLimit(maxMessageSize)
//conn.SetReadDeadline(time.Now().Add(pongWait))
//conn.SetWriteDeadline(time.Now().Add(writeWait))
//conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
