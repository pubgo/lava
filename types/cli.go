package types

import (
	"github.com/urfave/cli/v2"
	"strings"
)

type Flags=[]cli.Flag
type Command = cli.Command

func EnvOf(str ...string) []string {
	for i := range str {
		str[i] = strings.ToUpper(str[i])
	}
	return str
}
