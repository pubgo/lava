package models

import "fmt"

const (
	ObjectTypeResource      = "res"
	ObjectTypeResourceGroup = "group"
	ObjectTypeAction        = "act"
)

type Object interface {
	ObjectType() string
	ObjectID() string
	String() string
}

func ObjRepr(o Object) string {
	return fmt.Sprintf("%s/%s", o.ObjectType(), o.ObjectID())
}
