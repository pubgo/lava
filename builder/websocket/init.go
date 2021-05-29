package websocket

import (
	"sync"

	"github.com/pubgo/xlog"
)

const name = "websocket"

var log xlog.Xlog
var cfg Cfg
var wsM sync.Map

func init() {
	xlog.Watch(func(logs xlog.Xlog) {
		log = logs.Named(name)
	})
}

//defer conn.Close()
//conn.SetReadLimit(maxMessageSize)
//conn.SetReadDeadline(time.Now().Add(pongWait))
//conn.SetWriteDeadline(time.Now().Add(writeWait))
//conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
