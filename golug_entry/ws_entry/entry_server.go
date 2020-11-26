package ws_entry

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/fasthttp/websocket"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
	"github.com/valyala/fasthttp"
)

func wsHandle(ctx context.Context, rsp interface{}) (err error) {
	defer xerror.RespErr(&err)

	//defer conn.Close()
	//conn.SetReadLimit(maxMessageSize)
	//conn.SetReadDeadline(time.Now().Add(pongWait))
	//conn.SetWriteDeadline(time.Now().Add(writeWait))
	//conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	mth := ctx.Value(mth{}).(reflect.Value)
	mthInType := mth.Type().In(1)
	mthOutType := mth.Type().In(2)
	c := rsp.(*websocket.Conn)
	cancel := xprocess.GoLoop(func(_ context.Context) (err error) {
		defer xerror.RespErr(&err)

		mt, msg, err := c.ReadMessage()
		xerror.Panic(err)

		mthIn := reflect.New(mthInType.Elem())
		ret := reflect.ValueOf(json.Unmarshal).Call([]reflect.Value{reflect.ValueOf(msg), mthIn})
		if !ret[0].IsNil() {
			return xerror.Wrap(ret[0].Interface().(error))
		}

		mthOut := reflect.New(mthOutType.Elem())
		ret = mth.Call([]reflect.Value{reflect.ValueOf(ctx), mthIn, mthOut})
		if !ret[0].IsNil() {
			return xerror.Wrap(ret[0].Interface().(error))
		}

		dt, _err := json.Marshal(mthOut.Interface())
		if err != nil {
			return xerror.Wrap(_err)
		}

		return xerror.Wrap(c.WriteMessage(mt, dt))
	})

	c.SetCloseHandler(func(code int, text string) error {
		xlog.Debugf("%d, %s", code, text)
		return xerror.Wrap(cancel())
	})

	return nil
}

func init() {
	view.Request().Header.VisitAll(func(key, value []byte) {
		headers[strings.ToLower(string(key))] = string(value)
	})

	var upgrade = websocket.FastHTTPUpgrader{
		HandshakeTimeout:  0,
		Subprotocols:      nil,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin:       func(ctx *fasthttp.RequestCtx) bool { return true },
	}

	if err != nil {
		if err == websocket.ErrBadHandshake {
			log.Errorf("%#v", err)
		}
		return
	}
	return xerror.Wrap(upgrade.Upgrade(view.Context(), func(conn *websocket.Conn) { xerror.Panic(handle(ctx, request, conn)) }))
}
