package env

import (
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/xerror"
)

// Decode decode env to v
func Decode(v interface{}) error {
	var envMap = make(map[string]string)
	for _, env := range os.Environ() {
		if envList := strings.SplitN(env, "=", 2); len(envList) == 2 && trim(envList[0]) != "" {
			envMap[envList[0]] = envList[1]
		}
	}

	var cfg = &mapstructure.DecoderConfig{
		TagName:          "env",
		Metadata:         nil,
		Result:           v,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}

	decoder, err := mapstructure.NewDecoder(cfg)
	xerror.Panic(err)

	return decoder.Decode(envMap)
}
