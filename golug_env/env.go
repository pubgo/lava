package golug_env

import (
	"os"
	"regexp"
	"strings"

	"github.com/pubgo/xerror"
)

var trim = strings.TrimSpace
var upper = strings.ToUpper
var envRegexp = regexp.MustCompile(`\${(.+)}`)

func handleKey(key string) string {
	if !strings.HasPrefix(key, Prefix) {
		key = Prefix + "_" + key
	}
	return key
}

func Set(key, value string) error {
	key = upper(handleKey(key))
	return xerror.Wrap(os.Setenv(key, value))
}

func GetEnv(names ...string) string {
	var val string
	Get(&val, names...)
	return val
}

func Get(val *string, names ...string) {
	for _, name := range names {
		env, ok := Lookup(name)
		env = trim(env)
		if ok && env != "" {
			*val = env
		}
	}
}

func Lookup(key string) (string, bool) {
	key = upper(handleKey(key))
	return os.LookupEnv(key)
}

func Unsetenv(key string) error {
	key = upper(handleKey(key))
	return os.Unsetenv(key)
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
		v := GetEnv(vs[0])
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
