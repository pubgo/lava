package golug_env

import (
	"encoding/base32"
	"os"
	"regexp"
	"strings"

	"github.com/pubgo/xerror"
)

// RunEnvMode 项目运行模式
type RunEnvMode struct {
	Dev     string
	Test    string
	Stag    string
	Prod    string
	Release string
}

var trim = strings.TrimSpace
var upper = strings.ToUpper

func withPrefix(key string) string {
	if Domain != "" {
		key = Domain + "_" + key
	}
	return key
}

func Set(key, value string) error {
	key = withPrefix(key)
	return xerror.Wrap(os.Setenv(upper(key), value))
}

func GetEnv(names ...string) string {
	var val string
	Get(&val, names...)
	return val
}

func Get(val *string, names ...string) {
	for _, name := range names {
		name = withPrefix(name)
		env, ok := os.LookupEnv(upper(name))
		env = trim(env)
		if ok && env != "" {
			*val = env
		}
	}
}

func GetSysEnv(names ...string) string {
	var val string
	GetSys(&val, names...)
	return val
}

func GetSys(val *string, names ...string) {
	for _, name := range names {
		env, ok := os.LookupEnv(upper(name))
		env = trim(env)
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

var _envRegexp = regexp.MustCompile(`\${(.+)}`)
var _safeEnvRegexp = regexp.MustCompile(`!{(.+)}`)

// ExpandEnv returns value of convert with environment variable.
// Return environment variable if value start with "${" and end with "}".
// Return default value if environment variable is empty or not exist.
//
// It accept value formats "${env}" , "${env||}}" , "${env||defaultValue}" , "defaultvalue".
// Examples:
//	v1 := config.ExpandValueEnv("${GOPATH}")			// return the GOPATH environment variable.
//	v2 := config.ExpandValueEnv("${GOAsta||/usr/local/go}")	// return the default value "/usr/local/go/".
//	v3 := config.ExpandValueEnv("Astaxie")				// return the value "Astaxie".
func ExpandEnv(value string) string {
	value = trim(value)

	// 匹配环境变量格式
	if _envRegexp.MatchString(value) {
		_vs := strings.Split(_envRegexp.FindStringSubmatch(value)[1], "||")
		_v := os.Getenv(upper(_vs[0]))
		if len(_vs) == 2 && _v == "" {
			_v = trim(_vs[1])
		}
		return _v
	}

	// 匹配加密数据格式
	if _safeEnvRegexp.MatchString(value) {
		_v := _safeEnvRegexp.FindStringSubmatch(value)[1]
		return string(myXorDecrypt(_v, []byte(DefaultSecret)))
	}

	return value
}

// myXorEncrypt encrypt
func myXorEncrypt(text, key []byte) string {
	var _lk = len(key)
	for i := 0; i < len(text); i++ {
		text[i] ^= key[i*i*i%_lk]
	}
	return base32.StdEncoding.EncodeToString(text)
}

//myXorDecrypt decrypt
func myXorDecrypt(text string, key []byte) []byte {
	var _lk = len(key)
	_text, err := base32.StdEncoding.DecodeString(text)
	xerror.Panic(err)

	for i := 0; i < len(_text); i++ {
		_text[i] ^= key[i*i*i%_lk]
	}
	return _text
}

// IsTrue true
func IsTrue(data string) bool {
	switch upper(data) {
	case "TRUE", "T", "1", "OK", "GOOD", "REAL", "ACTIVE", "ENABLED":
		return true
	default:
		return false
	}
}

func IsDev() bool {
	return Mode == RunMode.Dev
}

func IsTest() bool {
	return Mode == RunMode.Test
}

func IsStag() bool {
	return Mode == RunMode.Stag
}

func IsProd() bool {
	return Mode == RunMode.Prod
}

func IsRelease() bool {
	return Mode == RunMode.Release
}
