package main

import (
	"fmt"
	"github.com/pubgo/lava/logger"
)

func main() {
	logger.ErrLog(fmt.Errorf("ok"))
}
