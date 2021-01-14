package golug_registry

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pubgo/xerror"
)

var registries sync.Map
var Default Registry

func Register(name string, r Registry) error {
	if name == "" || r == nil {
		return fmt.Errorf("[name] or [r] is nil")
	}

	_, ok := registries.LoadOrStore(name, r)
	if ok {
		return fmt.Errorf("registry %s is exists", name)
	}
	return nil
}

func Init(rawUrl string) {
	url1, err := url.Parse(rawUrl)
	xerror.PanicF(err, "url %s parse error", rawUrl)

	scheme := url1.Scheme
	xerror.Assert(Get(scheme) == nil, "registry [%s] not exists", scheme)

	params := url1.Query()
	var opts []Option
	if val := params.Get("secure"); val != "" {
		b, err := strconv.ParseBool(val)
		xerror.PanicF(err, "secure %s ParseBool error", val)
		opts = append(opts, Secure(b))
	}

	if val := params.Get("timeout"); val != "" {
		dur, err := time.ParseDuration(val)
		xerror.PanicF(err, "timeout %s ParseDuration error", val)
		opts = append(opts, Timeout(dur))
	}

	if val := params.Get("ttl"); val != "" {
		dur, err := time.ParseDuration(val)
		xerror.PanicF(err, "ttl %s ParseDuration error", val)
		opts = append(opts, TTL(dur))
	}

	addrs := strings.Split(url1.Host, ",")
	xerror.Assert(len(addrs) == 0, "%s host should be nil", rawUrl)
	opts = append(opts, Addrs(addrs...))

	Default = Get(scheme)
	xerror.PanicF(Default.Init(opts...), "[%s] init error", scheme)
}

func Get(name string) Registry {
	val, ok := registries.Load(name)
	if !ok {
		return nil
	}

	return val.(Registry)
}

func List() map[string]Registry {
	var data = make(map[string]Registry)
	registries.Range(func(key, value interface{}) bool {
		data[key.(string)] = value.(Registry)
		return true
	})
	return data
}
