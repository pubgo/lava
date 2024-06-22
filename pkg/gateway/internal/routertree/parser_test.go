package routertree

import (
	"github.com/pubgo/funk/assert"
	"testing"
)

func TestName(t *testing.T) {
	t.Log(parseToRoute(assert.Exit1(parse("/hello"))))
	t.Log(parseToRoute(assert.Exit1(parse("/hello/world"))))
	t.Log(parseToRoute(assert.Exit1(parse("/hello-world"))))
	t.Log(parseToRoute(assert.Exit1(parse("/hello_world"))))
	t.Log(parseToRoute(assert.Exit1(parse("/hello.world"))))
	t.Log(parseToRoute(assert.Exit1(parse("/user.echo"))))
	t.Log(parseToRoute(assert.Exit1(parse("/user.echo/{abc.abc}/hello"))))
}
