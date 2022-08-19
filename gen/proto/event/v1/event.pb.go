// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: proto/event/v1/event.proto

package eventpbv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type EventType int32

const (
	EventType_UNKNOWN EventType = 0
	EventType_CREATE  EventType = 1
	EventType_UPDATE  EventType = 2
	EventType_DELETE  EventType = 3
)

// Enum value maps for EventType.
var (
	EventType_name = map[int32]string{
		0: "UNKNOWN",
		1: "CREATE",
		2: "UPDATE",
		3: "DELETE",
	}
	EventType_value = map[string]int32{
		"UNKNOWN": 0,
		"CREATE":  1,
		"UPDATE":  2,
		"DELETE":  3,
	}
)

func (x EventType) Enum() *EventType {
	p := new(EventType)
	*p = x
	return p
}

func (x EventType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EventType) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_event_v1_event_proto_enumTypes[0].Descriptor()
}

func (EventType) Type() protoreflect.EnumType {
	return &file_proto_event_v1_event_proto_enumTypes[0]
}

func (x EventType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use EventType.Descriptor instead.
func (EventType) EnumDescriptor() ([]byte, []int) {
	return file_proto_event_v1_event_proto_rawDescGZIP(), []int{0}
}

var File_proto_event_v1_event_proto protoreflect.FileDescriptor

var file_proto_event_v1_event_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2f, 0x76, 0x31,
	0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x6c, 0x61,
	0x76, 0x61, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2a, 0x3c, 0x0a, 0x09, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e,
	0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x52, 0x45, 0x41, 0x54, 0x45, 0x10,
	0x01, 0x12, 0x0a, 0x0a, 0x06, 0x55, 0x50, 0x44, 0x41, 0x54, 0x45, 0x10, 0x02, 0x12, 0x0a, 0x0a,
	0x06, 0x44, 0x45, 0x4c, 0x45, 0x54, 0x45, 0x10, 0x03, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x75, 0x62, 0x67, 0x6f, 0x2f, 0x6c, 0x61,
	0x76, 0x61, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_event_v1_event_proto_rawDescOnce sync.Once
	file_proto_event_v1_event_proto_rawDescData = file_proto_event_v1_event_proto_rawDesc
)

func file_proto_event_v1_event_proto_rawDescGZIP() []byte {
	file_proto_event_v1_event_proto_rawDescOnce.Do(func() {
		file_proto_event_v1_event_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_event_v1_event_proto_rawDescData)
	})
	return file_proto_event_v1_event_proto_rawDescData
}

var file_proto_event_v1_event_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_event_v1_event_proto_goTypes = []interface{}{
	(EventType)(0), // 0: lava.event.v1.EventType
}
var file_proto_event_v1_event_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_event_v1_event_proto_init() }
func file_proto_event_v1_event_proto_init() {
	if File_proto_event_v1_event_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_event_v1_event_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_event_v1_event_proto_goTypes,
		DependencyIndexes: file_proto_event_v1_event_proto_depIdxs,
		EnumInfos:         file_proto_event_v1_event_proto_enumTypes,
	}.Build()
	File_proto_event_v1_event_proto = out.File
	file_proto_event_v1_event_proto_rawDesc = nil
	file_proto_event_v1_event_proto_goTypes = nil
	file_proto_event_v1_event_proto_depIdxs = nil
}
