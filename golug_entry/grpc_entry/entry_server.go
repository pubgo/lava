package grpc_entry

import (
	"encoding/json"
	"fmt"
	"github.com/pubgo/golug/golug_data"
	"net/http"
	"reflect"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type entryServerWrapper struct {
	*grpc.Server
	handlers []func()
}

func (t *entryServerWrapper) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	defer xerror.RespExit()

	t.Server.RegisterService(sd, handler)

	spn := fmt.Sprintf("%s.%s", sd.Metadata, sd.ServiceName)
	val, ok := golug_data.Get(spn)
	if !ok {
		return
	}

	var data map[string]interface{}
	xerror.Panic(json.Unmarshal([]byte(val.(string)), &data))

	vh := reflect.ValueOf(handler)
	for mthName, v := range data {
		vh.MethodByName(mthName)
	}

	http.HandleFunc("", func(writer http.ResponseWriter, request *http.Request) {
		grpc.DialContext()
	})

	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		grpclog.Fatalf("grpc: Server.Register found the handler of type %v that does not satisfy %v", st, ht)
	}

}
