package logs

import (
	"fmt"

	"github.com/nxadm/tail"
)

func tailFile(filename string, config tail.Config, done chan bool) {
	defer func() { done <- true }()
	t, err := tail.TailFile(filename, config)
	if err != nil {
		fmt.Println(err)
		return
	}
	for line := range t.Lines {
		fmt.Println(line.Err,line.Time,line.Num,line.SeekInfo)
	}
	err = t.Wait()
	if err != nil {
		fmt.Println(err)
	}
}
