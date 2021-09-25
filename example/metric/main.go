package main

import (
	"fmt"
	"github.com/pubgo/lug/logger"
)

func main() {
	logger.ErrLog(fmt.Errorf("ok"))
}
