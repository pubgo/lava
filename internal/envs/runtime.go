package envs

import (
	"fmt"

	"github.com/pubgo/xerror"
)

//  项目名字, 由project和domain组成
var name = ""

// SetName 设置项目名字
func SetName(project, domain string) {
	xerror.Assert(project == "" || domain == "", "[project,domain] should not be null")
	name = fmt.Sprintf("%s.%s", project, domain)
}

func Name() string { return name }
