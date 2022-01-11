package config

import "testing"

func TestName(t *testing.T) {
	Init()

	t.Log(CfgPath)
	t.Log(Home)
}
