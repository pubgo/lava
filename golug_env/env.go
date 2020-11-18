package golug_env

import (
	"os"
	"strings"

	"github.com/pubgo/xerror"
)

var prefix string

func Prefix(p string) {
	prefix = p
}

func withPrefix(key string) string {
	if prefix != "" {
		key = prefix + "_" + key
	}
	return key
}

func Set(key, value string) error {
	key = withPrefix(key)
	return xerror.Wrap(os.Setenv(strings.ToUpper(key), value))
}

func Get(val *string, names ...string) {
	for _, name := range names {
		name = withPrefix(name)
		env, ok := os.LookupEnv(strings.ToUpper(name))
		env = strings.TrimSpace(env)
		if ok && env != "" {
			*val = env
		}
	}
}

// Expand
// replaces ${var} or $var in the string according to the values
// of the current environment variables. References to undefined
// variables are replaced by the empty string.
func Expand(data string) string {
	return os.Expand(data, func(s string) string { return withPrefix(s) })
}

func Clear() {
	os.Clearenv()
}

func Lookup(key string) (string, bool) {
	key = withPrefix(key)
	return os.LookupEnv(key)
}

func Unsetenv(key string) error {
	key = withPrefix(key)
	return os.Unsetenv(key)
}
