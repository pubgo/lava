// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/user/user.proto

package gid

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/protobuf/types/known/emptypb"
	_ "google.golang.org/genproto/googleapis/rpc/status"
	_ "google.golang.org/protobuf/types/known/fieldmaskpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	_ "github.com/gogo/protobuf/gogoproto"
	_ "google.golang.org/protobuf/types/descriptorpb"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	_ "google.golang.org/protobuf/types/known/wrapperspb"
	_ "google.golang.org/protobuf/types/known/durationpb"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/protobuf/types/known/structpb"
	github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *Tag) Validate() error {
	return nil
}
func (this *GenerateResponse) Validate() error {
	return nil
}
func (this *GenerateRequest) Validate() error {
	return nil
}
func (this *TypesRequest) Validate() error {
	return nil
}
func (this *TypesResponse) Validate() error {
	return nil
}
func (this *ErrorResponse) Validate() error {
	if this.Error != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Error); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Error", err)
		}
	}
	return nil
}
func (this *ErrorObject) Validate() error {
	return nil
}
func (this *ABitOfEverything) Validate() error {
	if this.SingleNested != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.SingleNested); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("SingleNested", err)
		}
	}
	for _, item := range this.Nested {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Nested", err)
			}
		}
	}
	if oneOfNester, ok := this.GetOneofValue().(*ABitOfEverything_OneofEmpty); ok {
		if oneOfNester.OneofEmpty != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.OneofEmpty); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("OneofEmpty", err)
			}
		}
	}
	// Validation of proto3 map<> fields is unsupported.
	// Validation of proto3 map<> fields is unsupported.
	// Validation of proto3 map<> fields is unsupported.
	if this.TimestampValue != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.TimestampValue); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("TimestampValue", err)
		}
	}
	for _, item := range this.RepeatedNestedAnnotation {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("RepeatedNestedAnnotation", err)
			}
		}
	}
	if this.NestedAnnotation != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.NestedAnnotation); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("NestedAnnotation", err)
		}
	}
	return nil
}
func (this *ABitOfEverything_Nested) Validate() error {
	return nil
}
func (this *ABitOfEverythingRepeated) Validate() error {
	return nil
}
func (this *CheckStatusResponse) Validate() error {
	if this.Status != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Status); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Status", err)
		}
	}
	return nil
}
func (this *Body) Validate() error {
	return nil
}
func (this *MessageWithBody) Validate() error {
	if this.Data != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Data); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Data", err)
		}
	}
	return nil
}
func (this *UpdateV2Request) Validate() error {
	if this.Abe != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Abe); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Abe", err)
		}
	}
	if this.UpdateMask != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.UpdateMask); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("UpdateMask", err)
		}
	}
	return nil
}
func (this *Book) Validate() error {
	if this.CreateTime != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.CreateTime); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("CreateTime", err)
		}
	}
	return nil
}
func (this *CreateBookRequest) Validate() error {
	if this.Book != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Book); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Book", err)
		}
	}
	return nil
}
func (this *UpdateBookRequest) Validate() error {
	if this.Book != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Book); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Book", err)
		}
	}
	if this.UpdateMask != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.UpdateMask); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("UpdateMask", err)
		}
	}
	return nil
}
