package routex

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

type MethodConfig struct {
	Descriptor protoreflect.MethodDescriptor
	MethodPath string
}
