package logs

import (
	"encoding/json"
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/nxadm/tail"
	"github.com/pubgo/xerror"
	"github.com/segmentio/ksuid"
)

func tailFile(index bleve.Index, filename string, config tail.Config, done chan bool) {
	defer xerror.RespExit()

	defer func() { done <- true }()
	t, err := tail.TailFile(filename, config)
	xerror.Panic(err)
	for line := range t.Lines {
		fmt.Println(line.Err, line.Time, line.Num, line.SeekInfo)
		var dd map[string]interface{}
		xerror.Panic(json.Unmarshal([]byte(line.Text), &dd))
		xerror.Panic(index.Index(ksuid.New().String(), dd))
	}
	xerror.Panic(t.Wait())
}
