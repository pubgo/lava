package env

import (
	"os"
	"strconv"
	"strings"
)

var trim = strings.TrimSpace

func handleKey(key string) string {
	if !strings.HasPrefix(key, Prefix) {
		key = Prefix + "_" + key
	}
	return strings.ToUpper(key)
}

func Set(key, value string) error {
	return os.Setenv(handleKey(key), value)
}

func Get(names ...string) string {
	var val string
	GetWith(&val, names...)
	return val
}

func GetWith(val *string, names ...string) {
	for _, name := range names {
		env, ok := Lookup(name)
		env = trim(env)
		if ok && env != "" {
			*val = env
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
	return os.LookupEnv(handleKey(key))
}

func Unsetenv(key string) error {
	return os.Unsetenv(handleKey(key))
}

// Expand returns value of convert with environment variable.
// Return environment variable if value start with "${" and end with "}".
// Return default value if environment variable is empty or not exist.
//
// It accept value formats "${env}" , "${env|}}" , "${env||defaultValue}" , "defaultvalue".
// Examples:
//	v1 := config.Expand("${GOPATH}")			// return the GOPATH environment variable.
//	v2 := config.Expand("${GOPATH||/usr/local/go}")	// return the default value "/usr/local/go/".
//	v3 := config.Expand("Astaxie")				// return the value "Astaxie".
func Expand(value string) string {
	return os.Expand(value, func(s string) string {
		vs := strings.Split(s, "||")
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
