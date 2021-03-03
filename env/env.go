package env

import (
	"github.com/pubgo/golug/gutils"

	"os"
	"regexp"
	"strings"
)

var trim = strings.TrimSpace
var envRegexp = regexp.MustCompile(`\${(.+)}`)

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
	GetVal(&val, names...)
	return val
}

func GetVal(val *string, names ...string) {
	for _, name := range names {
		env, ok := Lookup(name)
		env = trim(env)
		if ok && env != "" {
			*val = env
		}
	}
}

func GetBoolVal(val *bool, names ...string) {
	*val = gutils.ParseBool(trim(Get(names...)))
}

func Lookup(key string) (string, bool) {
	return os.LookupEnv(handleKey(key))
}

func Unsetenv(key string) error {
	return os.Unsetenv(handleKey(key))
}

// ExpandEnv returns value of convert with environment variable.
// Return environment variable if value start with "${" and end with "}".
// Return default value if environment variable is empty or not exist.
//
// It accept value formats "${env}" , "${env||}}" , "${env||defaultValue}" , "defaultvalue".
// Examples:
//	v1 := config.ExpandValueEnv("${GOPATH}")			// return the GOPATH environment variable.
//	v2 := config.ExpandValueEnv("${GOAsta||/usr/local/go}")	// return the default value "/usr/local/go/".
//	v3 := config.ExpandValueEnv("Astaxie")				// return the value "Astaxie".
func Expand(value string) string {
	value = trim(value)

	// 匹配环境变量格式
	if envRegexp.MatchString(value) {
		vs := strings.Split(envRegexp.FindStringSubmatch(value)[1], "||")
		v := Get(vs[0])
		if len(vs) == 2 && v == "" {
			v = trim(vs[1])
		}
		return v
	}

	return value
}

func List() map[string]string {
	var data = make(map[string]string)
	for _, env := range os.Environ() {
		envs := strings.SplitN(env, "=", 2)
		if len(envs) != 2 {
			continue
		}

		key := trim(envs[0])
		val := trim(envs[1])
		if key == "" {
			continue
		}

		if !strings.HasPrefix(key, Prefix) {
			continue
		}

		data[key] = trim(val)
	}
	return data
}
