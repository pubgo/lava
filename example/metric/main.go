package main

import (
	"fmt"
	"github.com/pubgo/lug/logutil"
)

func main() {
	logutil.ErrLog(fmt.Errorf("ok"))
}
