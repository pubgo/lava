package env

import (
	"testing"

	"github.com/iancoleman/strcase"
	"github.com/pubgo/xerror"
)

func TestSnakeCase(t *testing.T) {
	defer xerror.RespTest(t)
	
	var snake = strcase.ToSnake
	xerror.AssertEqual(snake("a.b"), "a_b")
	xerror.AssertEqual(snake("a.b"), "a_b")
	xerror.AssertEqual(snake("a-b"), "a_b")
	xerror.AssertEqual(snake("aBcD"), "a_bc_d")
	xerror.AssertEqual(snake("aaBBccDD"), "aa_b_bcc_dd")
	xerror.AssertEqual(snake("aaBB/ccDD"), "aa_bb/cc_dd")
	t.Log(Pwd)
	t.Log(Home)
	t.Log(Hostname)
}
