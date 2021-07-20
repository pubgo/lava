package env

import (
	"github.com/pubgo/xerror"

	"testing"
)

func TestSnakeCase(t *testing.T) {
	xerror.RespTest(t)

	xerror.Assert(snakeCase("a.b") != "a.b", "snakeCase error")
	xerror.Assert(snakeCase("a-b") != "a-b", "snakeCase error")
	xerror.Assert(snakeCase("aBcD") != "a_bc_d", "snakeCase error")
	xerror.Assert(snakeCase("aaBBccDD") != "aa_b_bcc_d_d", "snakeCase error")
	xerror.Assert(snakeCase("aaBB/ccDD") != "aa_b_b/cc_d_d", "snakeCase error")
}
