package env

import (
	"os"
	"strconv"
	"strings"

	"github.com/pubgo/lava/pkg/errutil"
	"github.com/pubgo/lava/pkg/typex"
)

var trim = strings.TrimSpace

func Set(key, value string) error {
	return os.Setenv(Key(key), value)
}

func Get(names ...string) string {
	var val string
	GetWith(&val, names...)
	return trim(val)
}

func MustGet(names ...string) string {
	var val string
	GetWith(&val, names...)
	if val == "" {
		panic(&errutil.Err{
			Msg: "env not found",
			Detail: typex.M{
				"names": names,
				"all":   os.Environ(),
			}})
	}
	return trim(val)
}

func GetWith(val *string, names ...string) {
	for _, name := range names {
		env, ok := Lookup(name)
		env = trim(env)
		if ok && env != "" {
			*val = trim(env)
		}
	}
}

func GetBoolVal(val *bool, names ...string) {
	var dt = trim(Get(names...))
	if dt == "" {
		return
	}

	v, err := strconv.ParseBool(dt)
	if err != nil {
		return
	}

	*val = v
}

func GetIntVal(val *int, names ...string) {
	var dt = trim(Get(names...))
	if dt == "" {
		return
	}

	v, err := strconv.Atoi(dt)
	if err != nil {
		return
	}

	*val = v
}

func GetFloatVal(val *float64, names ...string) {
	var dt = trim(Get(names...))
	if dt == "" {
		return
	}

	v, err := strconv.ParseFloat(dt, 32)
	if err != nil {
		return
	}

	*val = v
}

func Lookup(key string) (string, bool) {
	return os.LookupEnv(Key(key))
}

func UnSetenv(key string) error {
	return os.Unsetenv(Key(key))
}

// Expand returns value of convert with environment variable.
// Return environment variable if value start with "${" and end with "}".
// Return default value if environment variable is empty or not exist.
//
// It accepts value formats "${env}" ,"${env||defaultValue}" , "defaultValue".
// Examples:
//	_ = Expand("${GOPATH}")
//	_ = Expand("${GOPATH||/usr/local/go}")
//	_ = Expand("hello")
func Expand(value string) string {
	return os.Expand(value, func(s string) string {
		vs := strings.SplitN(s, "||", 2)
		v := Get(trim(vs[0]))
		if len(vs) == 2 && v == "" {
			v = trim(vs[1])
		}
		return v
	})
}

func List() map[string]string {
	var data = make(map[string]string)
	for _, env := range os.Environ() {
		envList := strings.SplitN(env, "=", 2)
		if len(envList) != 2 {
			continue
		}

		key := trim(envList[0])
		val := trim(envList[1])
		if key == "" {
			continue
		}

		data[key] = trim(val)
	}
	return data
}

func Key(key string) string {
	return strings.ToUpper(trim(key))
}
