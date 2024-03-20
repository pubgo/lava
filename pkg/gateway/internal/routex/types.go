package routex

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"net/http"
)

type serviceOptions struct {
	codecNames, compressorNames map[string]struct{}
	preferredCodec              string
	maxMsgBufferBytes           uint32
	maxGetURLBytes              uint32
}

type methodConfig struct {
	*serviceOptions
	descriptor                protoreflect.MethodDescriptor
	requestType, responseType protoreflect.MessageType
	methodPath                string
	handler                   http.Handler
	httpRule                  *routeTarget // First HTTP rule, if any.
}
