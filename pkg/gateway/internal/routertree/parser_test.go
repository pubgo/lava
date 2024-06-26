package routertree

import (
	"testing"

	"github.com/pubgo/funk/assert"
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

func TestName1(t *testing.T) {
	t.Log(string([]byte(nil)) == string(make([]byte, 0)))
	t.Log(string([]byte(nil)))
	t.Log(string(make([]byte, 0)))
}
