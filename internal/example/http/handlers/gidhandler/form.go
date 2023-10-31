package gidhandler

type GenerateRequest struct {
	// type of id e.g uuid, shortid, snowflake (64 bit), bigflake (128 bit)
	Type string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
}

type GenerateResponse struct {
	// the unique id generated
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// the type of id generated
	Type string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
}

// List the types of IDs available. No query params needed.
type TypesRequest struct{}

// TypesResponse 返回值类型
type TypesResponse struct {
	Types []string `json:"types,omitempty" doc:"类型" required:"true" example:"[\"a\",\"b\"]" readOnly:"true"`
}
