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

func Set(key, value string) error {
	return xerror.Wrap(os.Setenv(upper(key), value))
}

func GetEnv(names ...string) string {
	var val string
	Get(&val, names...)
	return val
}

func Get(val *string, names ...string) {
	for _, name := range names {
		env, ok := os.LookupEnv(upper(name))
		env = trim(env)
		if ok && env != "" {
			*val = env
		}
	}
}

func Lookup(key string) (string, bool) { return os.LookupEnv(key) }

func Unsetenv(key string) error { return os.Unsetenv(key) }

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
		_vs := strings.Split(envRegexp.FindStringSubmatch(value)[1], "||")
		_v := os.Getenv(upper(_vs[0]))
		if len(_vs) == 2 && _v == "" {
			_v = trim(_vs[1])
		}
		return _v
	}

	return value
}
