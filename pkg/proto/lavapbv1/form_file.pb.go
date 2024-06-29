// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v5.27.0
// source: proto/lava/form_file.proto

package lavapbv1

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

type FormFile struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Filename    string   `protobuf:"bytes,2,opt,name=filename,proto3" json:"filename,omitempty"`
	ContentType []string `protobuf:"bytes,3,rep,name=content_type,json=contentType,proto3" json:"content_type,omitempty"`
	Site        int64    `protobuf:"varint,4,opt,name=site,proto3" json:"site,omitempty"`
}

func (x *FormFile) Reset() {
	*x = FormFile{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_lava_form_file_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FormFile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FormFile) ProtoMessage() {}

func (x *FormFile) ProtoReflect() protoreflect.Message {
	mi := &file_proto_lava_form_file_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FormFile.ProtoReflect.Descriptor instead.
func (*FormFile) Descriptor() ([]byte, []int) {
	return file_proto_lava_form_file_proto_rawDescGZIP(), []int{0}
}

func (x *FormFile) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FormFile) GetFilename() string {
	if x != nil {
		return x.Filename
	}
	return ""
}

func (x *FormFile) GetContentType() []string {
	if x != nil {
		return x.ContentType
	}
	return nil
}

func (x *FormFile) GetSite() int64 {
	if x != nil {
		return x.Site
	}
	return 0
}

var File_proto_lava_form_file_proto protoreflect.FileDescriptor

var file_proto_lava_form_file_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x61, 0x76, 0x61, 0x2f, 0x66, 0x6f, 0x72,
	0x6d, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x6c, 0x61,
	0x76, 0x61, 0x2e, 0x76, 0x31, 0x22, 0x71, 0x0a, 0x08, 0x46, 0x6f, 0x72, 0x6d, 0x46, 0x69, 0x6c,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x04, 0x73, 0x69, 0x74, 0x65, 0x42, 0x33, 0x5a, 0x31, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x75, 0x62, 0x67, 0x6f, 0x2f, 0x6c, 0x61, 0x76,
	0x61, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x61, 0x76, 0x61,
	0x70, 0x62, 0x76, 0x31, 0x3b, 0x6c, 0x61, 0x76, 0x61, 0x70, 0x62, 0x76, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_lava_form_file_proto_rawDescOnce sync.Once
	file_proto_lava_form_file_proto_rawDescData = file_proto_lava_form_file_proto_rawDesc
)

func file_proto_lava_form_file_proto_rawDescGZIP() []byte {
	file_proto_lava_form_file_proto_rawDescOnce.Do(func() {
		file_proto_lava_form_file_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_lava_form_file_proto_rawDescData)
	})
	return file_proto_lava_form_file_proto_rawDescData
}

var file_proto_lava_form_file_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_proto_lava_form_file_proto_goTypes = []interface{}{
	(*FormFile)(nil), // 0: lava.v1.FormFile
}
var file_proto_lava_form_file_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_lava_form_file_proto_init() }
func file_proto_lava_form_file_proto_init() {
	if File_proto_lava_form_file_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_lava_form_file_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FormFile); i {
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
			RawDescriptor: file_proto_lava_form_file_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_lava_form_file_proto_goTypes,
		DependencyIndexes: file_proto_lava_form_file_proto_depIdxs,
		MessageInfos:      file_proto_lava_form_file_proto_msgTypes,
	}.Build()
	File_proto_lava_form_file_proto = out.File
	file_proto_lava_form_file_proto_rawDesc = nil
	file_proto_lava_form_file_proto_goTypes = nil
	file_proto_lava_form_file_proto_depIdxs = nil
}
