// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/proto/login/merge.proto

// 账户合并相关

package login

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/gogo/protobuf/gogoproto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *TelephoneRequest) Validate() error {
	return nil
}
func (this *WeChatRequest) Validate() error {
	return nil
}
func (this *WeChatUnMergeRequest) Validate() error {
	return nil
}
func (this *Reply) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}