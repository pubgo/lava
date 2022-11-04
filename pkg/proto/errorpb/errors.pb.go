// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: proto/errorpb/errors.proto

package errorpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Code int32

const (
	Code_OK                 Code = 0
	Code_Canceled           Code = 1
	Code_Unknown            Code = 2
	Code_InvalidArgument    Code = 3
	Code_DeadlineExceeded   Code = 4
	Code_NotFound           Code = 5
	Code_AlreadyExists      Code = 6
	Code_PermissionDenied   Code = 7
	Code_ResourceExhausted  Code = 8
	Code_FailedPrecondition Code = 9
	Code_Aborted            Code = 10
	Code_OutOfRange         Code = 11
	Code_Unimplemented      Code = 12
	Code_Internal           Code = 13
	Code_Unavailable        Code = 14
	Code_DataLoss           Code = 15
	Code_Unauthenticated    Code = 16
)

// Enum value maps for Code.
var (
	Code_name = map[int32]string{
		0:  "OK",
		1:  "Canceled",
		2:  "Unknown",
		3:  "InvalidArgument",
		4:  "DeadlineExceeded",
		5:  "NotFound",
		6:  "AlreadyExists",
		7:  "PermissionDenied",
		8:  "ResourceExhausted",
		9:  "FailedPrecondition",
		10: "Aborted",
		11: "OutOfRange",
		12: "Unimplemented",
		13: "Internal",
		14: "Unavailable",
		15: "DataLoss",
		16: "Unauthenticated",
	}
	Code_value = map[string]int32{
		"OK":                 0,
		"Canceled":           1,
		"Unknown":            2,
		"InvalidArgument":    3,
		"DeadlineExceeded":   4,
		"NotFound":           5,
		"AlreadyExists":      6,
		"PermissionDenied":   7,
		"ResourceExhausted":  8,
		"FailedPrecondition": 9,
		"Aborted":            10,
		"OutOfRange":         11,
		"Unimplemented":      12,
		"Internal":           13,
		"Unavailable":        14,
		"DataLoss":           15,
		"Unauthenticated":    16,
	}
)

func (x Code) Enum() *Code {
	p := new(Code)
	*p = x
	return p
}

func (x Code) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Code) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_errorpb_errors_proto_enumTypes[0].Descriptor()
}

func (Code) Type() protoreflect.EnumType {
	return &file_proto_errorpb_errors_proto_enumTypes[0]
}

func (x Code) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Code.Descriptor instead.
func (Code) EnumDescriptor() ([]byte, []int) {
	return file_proto_errorpb_errors_proto_rawDescGZIP(), []int{0}
}

type Error struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code      uint32            `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	BizCode   string            `protobuf:"bytes,2,opt,name=biz_code,proto3" json:"biz_code,omitempty"`
	Reason    string            `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
	Operation string            `protobuf:"bytes,4,opt,name=operation,proto3" json:"operation,omitempty"`
	Project   string            `protobuf:"bytes,5,opt,name=project,proto3" json:"project,omitempty"`
	Version   string            `protobuf:"bytes,6,opt,name=version,proto3" json:"version,omitempty"`
	ErrMsg    string            `protobuf:"bytes,7,opt,name=err_msg,proto3" json:"err_msg,omitempty"`
	ErrDetail string            `protobuf:"bytes,8,opt,name=err_detail,proto3" json:"err_detail,omitempty"`
	Tags      map[string]string `protobuf:"bytes,9,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Error) Reset() {
	*x = Error{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_errorpb_errors_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Error) ProtoMessage() {}

func (x *Error) ProtoReflect() protoreflect.Message {
	mi := &file_proto_errorpb_errors_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Error.ProtoReflect.Descriptor instead.
func (*Error) Descriptor() ([]byte, []int) {
	return file_proto_errorpb_errors_proto_rawDescGZIP(), []int{0}
}

func (x *Error) GetCode() uint32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *Error) GetBizCode() string {
	if x != nil {
		return x.BizCode
	}
	return ""
}

func (x *Error) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *Error) GetOperation() string {
	if x != nil {
		return x.Operation
	}
	return ""
}

func (x *Error) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

func (x *Error) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *Error) GetErrMsg() string {
	if x != nil {
		return x.ErrMsg
	}
	return ""
}

func (x *Error) GetErrDetail() string {
	if x != nil {
		return x.ErrDetail
	}
	return ""
}

func (x *Error) GetTags() map[string]string {
	if x != nil {
		return x.Tags
	}
	return nil
}

var File_proto_errorpb_errors_proto protoreflect.FileDescriptor

var file_proto_errorpb_errors_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x70, 0x62, 0x2f,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x6c, 0x61,
	0x76, 0x61, 0x2e, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc6, 0x02, 0x0a, 0x05,
	0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x62, 0x69, 0x7a,
	0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x62, 0x69, 0x7a,
	0x5f, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x1c, 0x0a,
	0x09, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x18, 0x0a, 0x07, 0x65, 0x72, 0x72, 0x5f, 0x6d, 0x73, 0x67, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x65, 0x72, 0x72, 0x5f, 0x6d, 0x73, 0x67, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x72, 0x72,
	0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65,
	0x72, 0x72, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x12, 0x30, 0x0a, 0x04, 0x74, 0x61, 0x67,
	0x73, 0x18, 0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6c, 0x61, 0x76, 0x61, 0x2e, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x73, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x54, 0x61, 0x67, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x1a, 0x37, 0x0a, 0x09, 0x54,
	0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x2a, 0xac, 0x02, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x06, 0x0a,
	0x02, 0x4f, 0x4b, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x65,
	0x64, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x10, 0x02,
	0x12, 0x13, 0x0a, 0x0f, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x41, 0x72, 0x67, 0x75, 0x6d,
	0x65, 0x6e, 0x74, 0x10, 0x03, 0x12, 0x14, 0x0a, 0x10, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e,
	0x65, 0x45, 0x78, 0x63, 0x65, 0x65, 0x64, 0x65, 0x64, 0x10, 0x04, 0x12, 0x0c, 0x0a, 0x08, 0x4e,
	0x6f, 0x74, 0x46, 0x6f, 0x75, 0x6e, 0x64, 0x10, 0x05, 0x12, 0x11, 0x0a, 0x0d, 0x41, 0x6c, 0x72,
	0x65, 0x61, 0x64, 0x79, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x10, 0x06, 0x12, 0x14, 0x0a, 0x10,
	0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x44, 0x65, 0x6e, 0x69, 0x65, 0x64,
	0x10, 0x07, 0x12, 0x15, 0x0a, 0x11, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x45, 0x78,
	0x68, 0x61, 0x75, 0x73, 0x74, 0x65, 0x64, 0x10, 0x08, 0x12, 0x16, 0x0a, 0x12, 0x46, 0x61, 0x69,
	0x6c, 0x65, 0x64, 0x50, 0x72, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x10,
	0x09, 0x12, 0x0b, 0x0a, 0x07, 0x41, 0x62, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x10, 0x0a, 0x12, 0x0e,
	0x0a, 0x0a, 0x4f, 0x75, 0x74, 0x4f, 0x66, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x10, 0x0b, 0x12, 0x11,
	0x0a, 0x0d, 0x55, 0x6e, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x65, 0x64, 0x10,
	0x0c, 0x12, 0x0c, 0x0a, 0x08, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x10, 0x0d, 0x12,
	0x0f, 0x0a, 0x0b, 0x55, 0x6e, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x10, 0x0e,
	0x12, 0x0c, 0x0a, 0x08, 0x44, 0x61, 0x74, 0x61, 0x4c, 0x6f, 0x73, 0x73, 0x10, 0x0f, 0x12, 0x13,
	0x0a, 0x0f, 0x55, 0x6e, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65,
	0x64, 0x10, 0x10, 0x42, 0x3f, 0x0a, 0x0f, 0x63, 0x6f, 0x6d, 0x2e, 0x6c, 0x61, 0x76, 0x61, 0x2e,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x50, 0x01, 0x5a, 0x27, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x75, 0x62, 0x67, 0x6f, 0x2f, 0x6c, 0x61, 0x76, 0x61, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x70,
	0x62, 0xf8, 0x01, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_errorpb_errors_proto_rawDescOnce sync.Once
	file_proto_errorpb_errors_proto_rawDescData = file_proto_errorpb_errors_proto_rawDesc
)

func file_proto_errorpb_errors_proto_rawDescGZIP() []byte {
	file_proto_errorpb_errors_proto_rawDescOnce.Do(func() {
		file_proto_errorpb_errors_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_errorpb_errors_proto_rawDescData)
	})
	return file_proto_errorpb_errors_proto_rawDescData
}

var file_proto_errorpb_errors_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_errorpb_errors_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_errorpb_errors_proto_goTypes = []interface{}{
	(Code)(0),     // 0: lava.errors.Code
	(*Error)(nil), // 1: lava.errors.Error
	nil,           // 2: lava.errors.Error.TagsEntry
}
var file_proto_errorpb_errors_proto_depIdxs = []int32{
	2, // 0: lava.errors.Error.tags:type_name -> lava.errors.Error.TagsEntry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proto_errorpb_errors_proto_init() }
func file_proto_errorpb_errors_proto_init() {
	if File_proto_errorpb_errors_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_errorpb_errors_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Error); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_errorpb_errors_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_errorpb_errors_proto_goTypes,
		DependencyIndexes: file_proto_errorpb_errors_proto_depIdxs,
		EnumInfos:         file_proto_errorpb_errors_proto_enumTypes,
		MessageInfos:      file_proto_errorpb_errors_proto_msgTypes,
	}.Build()
	File_proto_errorpb_errors_proto = out.File
	file_proto_errorpb_errors_proto_rawDesc = nil
	file_proto_errorpb_errors_proto_goTypes = nil
	file_proto_errorpb_errors_proto_depIdxs = nil
}
