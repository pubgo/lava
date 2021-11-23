package types

import (
	"strings"

	"github.com/urfave/cli/v2"
)

type Flags = []cli.Flag
type Command = cli.Command
type Commands = []cli.Command

func EnvOf(str ...string) []string {
	for i := range str {
		str[i] = strings.ToUpper(str[i])
	}
	return str
}
