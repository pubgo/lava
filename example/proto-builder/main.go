package main

import (
	"fmt"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/desc/protoprint"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/pubgo/xerror"
)

func main() {
	md, err := desc.LoadMessageDescriptorForMessage((*empty.Empty)(nil))
	xerror.Exit(err)

	file := builder.NewFile("foo/bar.proto").SetPackageName("foo.bar")
	en := builder.NewEnum("Options").
		AddValue(builder.NewEnumValue("OPTION_1").SetComments(builder.Comments{LeadingComment: " OPTION_1"})).
		AddValue(builder.NewEnumValue("OPTION_2")).
		AddValue(builder.NewEnumValue("OPTION_3"))
	file.AddEnum(en)

	msg := builder.NewMessage("FooRequest").
		AddField(builder.NewField("id", builder.FieldTypeInt64())).
		AddField(builder.NewField("name", builder.FieldTypeString())).
		AddField(builder.NewField("options", builder.FieldTypeEnum(en)).
			SetRepeated())
	file.AddMessage(msg)

	sb := builder.NewService("FooService").
		AddMethod(builder.NewMethod("DoSomething", builder.RpcTypeMessage(msg, false), builder.RpcTypeMessage(msg, false))).
		AddMethod(builder.NewMethod("ReturnThings", builder.RpcTypeImportedMessage(md, false), builder.RpcTypeMessage(msg, true)))
	file.AddService(sb)

	fd, err := file.Build()
	xerror.Exit(err)
	fmt.Println(fd.String())
	fmt.Println(fd.AsProto().String())
	var p protoprint.Printer
	fmt.Println(xerror.ExitErr(p.PrintProtoToString(fd)))

	files := map[string]string{"test.proto": `
syntax = "proto3";
import "google/protobuf/descriptor.proto";
message Test {}
message Foo {
  repeated Bar bar = 1;
  message Bar {
    Baz baz = 1;
    string name = 2;
  }
  enum Baz {
	ZERO = 0;
	FROB = 1;
	NITZ = 2;
  }
}
extend google.protobuf.MethodOptions {
  Foo foo = 54321;
}
service TestService {
  rpc Get (Test) returns (Test) {
    option (foo).bar = { baz:FROB name:"abc" };
    option (foo).bar = { baz:NITZ name:"xyz" };
  }
}
`}

	pa := &protoparse.Parser{Accessor: protoparse.FileContentsFromMap(files)}
	fds, err := pa.ParseFiles("test.proto")
	//proto.Equal()
	_ = fds
}
