package main

import (
	"os"

	"github.com/magefile/mage/mage"
)

func main() {
	os.Args = []string{os.Args[0], "-v", "install"}
	os.Exit(mage.Main())
}
