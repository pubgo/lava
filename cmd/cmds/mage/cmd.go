package mage

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/mage"
	"github.com/manifoldco/promptui"
	"github.com/manifoldco/promptui/list"
	"github.com/urfave/cli/v2"
)

var Cmd = &cli.Command{
	Name:  "mage",
	Usage: "mage install",
	Action: func(ctx *cli.Context) error {
		cmds := exec.Command("mage", "-l")
		cmds.Stderr = os.Stderr
		out, err := cmds.Output()
		if err != nil {
			return err
		}

		scan := bufio.NewScanner(bytes.NewBuffer(out))

		var targets []string
		for scan.Scan() {
			line := scan.Text()
			if strings.HasPrefix(line, "Targets:") {
				continue
			}
			line = strings.TrimSpace(line)
			targets = append(targets, line)
		}

		templates := &promptui.SelectTemplates{
			Label:    "{{.}}",
			Active:   promptui.IconSelect + " {{.}}",
			Inactive: "  {{.|faint}}",
			Selected: promptui.IconGood + " {{.}}",
		}

		size := maxSize
		if len(targets) < size {
			size = len(targets)
		}

		prompt := promptui.Select{
			Label:             "Select a mage target:",
			Items:             targets,
			Templates:         templates,
			HideHelp:          true,
			Size:              size,
			Searcher:          searcher(targets),
			StartInSearchMode: true,
		}

		_, result, err := prompt.Run()

		if err != nil {
			return err
		}

		result = strings.Split(result, " ")[0]

		fmt.Printf("mage %s\n", result)

		os.Args = append(os.Args, result)
		os.Exit(mage.Main())
		return nil
	},
}

const (
	maxSize = 10
)

func searcher(targets []string) list.Searcher {
	return func(input string, index int) bool {
		if strings.Contains(strings.ToLower(targets[index]), input) {
			return true
		}
		return false
	}
}
