package test

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	fmt.Println(ErrTestUnknown.String())
	fmt.Println(ErrTestNotFound)
}
