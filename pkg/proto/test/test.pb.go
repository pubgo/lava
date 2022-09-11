// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: proto/test/test.proto

package testpbv1

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

type TestCode int32

const (
	TestCode_OK TestCode = 0
	// NotFound 找不到
	TestCode_NotFound TestCode = 100000
	// Unknown 未知
	TestCode_Unknown TestCode = 100001
)

// Enum value maps for TestCode.
var (
	TestCode_name = map[int32]string{
		0:      "OK",
		100000: "NotFound",
		100001: "Unknown",
	}
	TestCode_value = map[string]int32{
		"OK":       0,
		"NotFound": 100000,
		"Unknown":  100001,
	}
)

func (x TestCode) Enum() *TestCode {
	p := new(TestCode)
	*p = x
	return p
}

func (x TestCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TestCode) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_test_test_proto_enumTypes[0].Descriptor()
}

func (TestCode) Type() protoreflect.EnumType {
	return &file_proto_test_test_proto_enumTypes[0]
}

func (x TestCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TestCode.Descriptor instead.
func (TestCode) EnumDescriptor() ([]byte, []int) {
	return file_proto_test_test_proto_rawDescGZIP(), []int{0}
}

var File_proto_test_test_proto protoreflect.FileDescriptor

var file_proto_test_test_proto_rawDesc = []byte{
	0x0a, 0x15, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x74, 0x65, 0x73,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x6c, 0x61, 0x76, 0x61, 0x2e, 0x74, 0x65,
	0x73, 0x74, 0x2e, 0x76, 0x31, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2a, 0x31, 0x0a, 0x08, 0x54, 0x65, 0x73, 0x74, 0x43,
	0x6f, 0x64, 0x65, 0x12, 0x06, 0x0a, 0x02, 0x4f, 0x4b, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x08, 0x4e,
	0x6f, 0x74, 0x46, 0x6f, 0x75, 0x6e, 0x64, 0x10, 0xa0, 0x8d, 0x06, 0x12, 0x0d, 0x0a, 0x07, 0x55,
	0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x10, 0xa1, 0x8d, 0x06, 0x42, 0x2f, 0x5a, 0x2d, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x75, 0x62, 0x67, 0x6f, 0x2f, 0x6c,
	0x61, 0x76, 0x61, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65,
	0x73, 0x74, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_proto_test_test_proto_rawDescOnce sync.Once
	file_proto_test_test_proto_rawDescData = file_proto_test_test_proto_rawDesc
)

func file_proto_test_test_proto_rawDescGZIP() []byte {
	file_proto_test_test_proto_rawDescOnce.Do(func() {
		file_proto_test_test_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_test_test_proto_rawDescData)
	})
	return file_proto_test_test_proto_rawDescData
}

var file_proto_test_test_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_test_test_proto_goTypes = []interface{}{
	(TestCode)(0), // 0: lava.test.v1.TestCode
}
var file_proto_test_test_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_test_test_proto_init() }
func file_proto_test_test_proto_init() {
	if File_proto_test_test_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_test_test_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_test_test_proto_goTypes,
		DependencyIndexes: file_proto_test_test_proto_depIdxs,
		EnumInfos:         file_proto_test_test_proto_enumTypes,
	}.Build()
	File_proto_test_test_proto = out.File
	file_proto_test_test_proto_rawDesc = nil
	file_proto_test_test_proto_goTypes = nil
	file_proto_test_test_proto_depIdxs = nil
}