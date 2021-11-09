# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [proto/hello/api.proto](#proto/hello/api.proto)
    - [TestApiOutput1](#hello.TestApiOutput1)
  
    - [TestApi](#hello.TestApi)
    - [TestApiV2](#hello.TestApiV2)
  
- [proto/hello/api1.proto](#proto/hello/api1.proto)
    - [TestApiData](#hello.TestApiData)
    - [TestApiOutput](#hello.TestApiOutput)
    - [TestReq](#hello.TestReq)
    - [TestReq.HeadersEntry](#hello.TestReq.HeadersEntry)
  
    - [PhoneType](#hello.PhoneType)
  
- [proto/hello/example.proto](#proto/hello/example.proto)
    - [ListUsersRequest](#hello.ListUsersRequest)
    - [UpdateUserRequest](#hello.UpdateUserRequest)
    - [User](#hello.User)
    - [UserRole](#hello.UserRole)
    - [UserRole.HeadersEntry](#hello.UserRole.HeadersEntry)
  
    - [Role](#hello.Role)
  
    - [UserService](#hello.UserService)
  
- [proto/hello/helloworld.proto](#proto/hello/helloworld.proto)
    - [HelloReply](#hello.HelloReply)
    - [HelloRequest](#hello.HelloRequest)
  
    - [Greeter](#hello.Greeter)
  
- [proto/hello/proto3.proto](#proto/hello/proto3.proto)
    - [Proto3Message](#hello.Proto3Message)
    - [Proto3Message.MapValue10Entry](#hello.Proto3Message.MapValue10Entry)
    - [Proto3Message.MapValue12Entry](#hello.Proto3Message.MapValue12Entry)
    - [Proto3Message.MapValue14Entry](#hello.Proto3Message.MapValue14Entry)
    - [Proto3Message.MapValue15Entry](#hello.Proto3Message.MapValue15Entry)
    - [Proto3Message.MapValue16Entry](#hello.Proto3Message.MapValue16Entry)
    - [Proto3Message.MapValue2Entry](#hello.Proto3Message.MapValue2Entry)
    - [Proto3Message.MapValue3Entry](#hello.Proto3Message.MapValue3Entry)
    - [Proto3Message.MapValue4Entry](#hello.Proto3Message.MapValue4Entry)
    - [Proto3Message.MapValue5Entry](#hello.Proto3Message.MapValue5Entry)
    - [Proto3Message.MapValue6Entry](#hello.Proto3Message.MapValue6Entry)
    - [Proto3Message.MapValue7Entry](#hello.Proto3Message.MapValue7Entry)
    - [Proto3Message.MapValue8Entry](#hello.Proto3Message.MapValue8Entry)
    - [Proto3Message.MapValue9Entry](#hello.Proto3Message.MapValue9Entry)
    - [Proto3Message.MapValueEntry](#hello.Proto3Message.MapValueEntry)
  
    - [EnumValue](#hello.EnumValue)
  
- [proto/hello/transport.proto](#proto/hello/transport.proto)
    - [Message](#hello.Message)
    - [Message.HeaderEntry](#hello.Message.HeaderEntry)
  
    - [Transport](#hello.Transport)
  
- [Scalar Value Types](#scalar-value-types)



<a name="proto/hello/api.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/hello/api.proto



<a name="hello.TestApiOutput1"></a>

### TestApiOutput1



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [google.protobuf.Value](#google.protobuf.Value) |  |  |
| abc | [string](#string) |  |  |





 

 

 


<a name="hello.TestApi"></a>

### TestApi
TestApi service

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Version | [TestReq](#hello.TestReq) | [TestApiOutput](#hello.TestApiOutput) | Version rpc |
| Version1 | [.google.protobuf.Value](#google.protobuf.Value) | [TestApiOutput1](#hello.TestApiOutput1) |  |
| VersionTest | [TestReq](#hello.TestReq) | [TestApiOutput](#hello.TestApiOutput) | VersionTest rpc |
| VersionTestCustom | [TestReq](#hello.TestReq) | [TestApiOutput](#hello.TestApiOutput) | VersionTest rpc custom |


<a name="hello.TestApiV2"></a>

### TestApiV2


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Version1 | [TestReq](#hello.TestReq) | [TestApiOutput](#hello.TestApiOutput) |  |
| VersionTest1 | [TestReq](#hello.TestReq) | [TestApiOutput](#hello.TestApiOutput) |  |

 



<a name="proto/hello/api1.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/hello/api1.proto



<a name="hello.TestApiData"></a>

### TestApiData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| version | [string](#string) |  |  |
| srvVersion | [string](#string) |  |  |






<a name="hello.TestApiOutput"></a>

### TestApiOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int32](#int32) |  |  |
| msg | [string](#string) |  |  |
| nowTime | [int64](#int64) |  |  |
| data | [TestApiData](#hello.TestApiData) |  |  |






<a name="hello.TestReq"></a>

### TestReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| input | [string](#string) |  |  |
| name | [string](#string) |  |  |
| lists | [google.protobuf.ListValue](#google.protobuf.ListValue) |  |  |
| headers | [TestReq.HeadersEntry](#hello.TestReq.HeadersEntry) | repeated |  |






<a name="hello.TestReq.HeadersEntry"></a>

### TestReq.HeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.ListValue](#google.protobuf.ListValue) |  |  |





 


<a name="hello.PhoneType"></a>

### PhoneType
枚举消息类型

| Name | Number | Description |
| ---- | ------ | ----------- |
| MOBILE | 0 | proto3版本中，首成员必须为0，成员不应有相同的值 |
| HOME | 1 |  |
| WORK | 2 |  |


 

 

 



<a name="proto/hello/example.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/hello/example.proto



<a name="hello.ListUsersRequest"></a>

### ListUsersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| created_since | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Only list users created after this timestamp |
| older_than | [google.protobuf.Duration](#google.protobuf.Duration) |  | Only list users older than this Duration |






<a name="hello.UpdateUserRequest"></a>

### UpdateUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#hello.User) |  | The user resource which replaces the resource on the server. |
| update_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  | The update mask applies to the resource. For the `FieldMask` definition, see https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#fieldmask |






<a name="hello.User"></a>

### User



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [uint32](#uint32) |  |  |
| role | [Role](#hello.Role) |  |  |
| create_date | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="hello.UserRole"></a>

### UserRole



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#hello.Role) |  |  |
| lists | [google.protobuf.ListValue](#google.protobuf.ListValue) |  |  |
| headers | [UserRole.HeadersEntry](#hello.UserRole.HeadersEntry) | repeated |  |






<a name="hello.UserRole.HeadersEntry"></a>

### UserRole.HeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.ListValue](#google.protobuf.ListValue) |  |  |





 


<a name="hello.Role"></a>

### Role


| Name | Number | Description |
| ---- | ------ | ----------- |
| GUEST | 0 |  |
| MEMBER | 1 |  |
| ADMIN | 2 |  |


 

 


<a name="hello.UserService"></a>

### UserService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| AddUser | [User](#hello.User) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| GetUser | [User](#hello.User) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| ListUsers | [ListUsersRequest](#hello.ListUsersRequest) | [User](#hello.User) stream |  |
| ListUsersByRole | [UserRole](#hello.UserRole) stream | [User](#hello.User) stream |  |
| UpdateUser | [UpdateUserRequest](#hello.UpdateUserRequest) | [User](#hello.User) |  |

 



<a name="proto/hello/helloworld.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/hello/helloworld.proto



<a name="hello.HelloReply"></a>

### HelloReply



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |






<a name="hello.HelloRequest"></a>

### HelloRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| strVal | [google.protobuf.StringValue](#google.protobuf.StringValue) |  |  |
| floatVal | [google.protobuf.FloatValue](#google.protobuf.FloatValue) |  |  |
| doubleVal | [google.protobuf.DoubleValue](#google.protobuf.DoubleValue) |  |  |
| boolVal | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  |  |
| bytesVal | [google.protobuf.BytesValue](#google.protobuf.BytesValue) |  |  |
| int32Val | [google.protobuf.Int32Value](#google.protobuf.Int32Value) |  |  |
| uint32Val | [google.protobuf.UInt32Value](#google.protobuf.UInt32Value) |  |  |
| int64Val | [google.protobuf.Int64Value](#google.protobuf.Int64Value) |  |  |
| uint64Val | [google.protobuf.UInt64Value](#google.protobuf.UInt64Value) |  |  |





 

 

 


<a name="hello.Greeter"></a>

### Greeter


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SayHello | [HelloRequest](#hello.HelloRequest) | [HelloReply](#hello.HelloReply) |  |

 



<a name="proto/hello/proto3.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/hello/proto3.proto



<a name="hello.Proto3Message"></a>

### Proto3Message



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nested | [Proto3Message](#hello.Proto3Message) |  | Next number: 46 |
| float_value | [float](#float) |  |  |
| double_value | [double](#double) |  |  |
| int64_value | [int64](#int64) |  |  |
| int32_value | [int32](#int32) |  |  |
| uint64_value | [uint64](#uint64) |  |  |
| uint32_value | [uint32](#uint32) |  |  |
| bool_value | [bool](#bool) |  |  |
| string_value | [string](#string) |  |  |
| bytes_value | [bytes](#bytes) |  |  |
| repeated_value | [string](#string) | repeated |  |
| repeated_message | [google.protobuf.UInt64Value](#google.protobuf.UInt64Value) | repeated |  |
| enum_value | [EnumValue](#hello.EnumValue) |  |  |
| repeated_enum | [EnumValue](#hello.EnumValue) | repeated |  |
| timestamp_value | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| duration_value | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |
| fieldmask_value | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |
| oneof_bool_value | [bool](#bool) |  |  |
| oneof_string_value | [string](#string) |  |  |
| wrapper_double_value | [google.protobuf.DoubleValue](#google.protobuf.DoubleValue) |  |  |
| wrapper_float_value | [google.protobuf.FloatValue](#google.protobuf.FloatValue) |  |  |
| wrapper_int64_value | [google.protobuf.Int64Value](#google.protobuf.Int64Value) |  |  |
| wrapper_int32_value | [google.protobuf.Int32Value](#google.protobuf.Int32Value) |  |  |
| wrapper_u_int64_value | [google.protobuf.UInt64Value](#google.protobuf.UInt64Value) |  |  |
| wrapper_u_int32_value | [google.protobuf.UInt32Value](#google.protobuf.UInt32Value) |  |  |
| wrapper_bool_value | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  |  |
| wrapper_string_value | [google.protobuf.StringValue](#google.protobuf.StringValue) |  |  |
| wrapper_bytes_value | [google.protobuf.BytesValue](#google.protobuf.BytesValue) |  |  |
| map_value | [Proto3Message.MapValueEntry](#hello.Proto3Message.MapValueEntry) | repeated |  |
| map_value2 | [Proto3Message.MapValue2Entry](#hello.Proto3Message.MapValue2Entry) | repeated |  |
| map_value3 | [Proto3Message.MapValue3Entry](#hello.Proto3Message.MapValue3Entry) | repeated |  |
| map_value4 | [Proto3Message.MapValue4Entry](#hello.Proto3Message.MapValue4Entry) | repeated |  |
| map_value5 | [Proto3Message.MapValue5Entry](#hello.Proto3Message.MapValue5Entry) | repeated |  |
| map_value6 | [Proto3Message.MapValue6Entry](#hello.Proto3Message.MapValue6Entry) | repeated |  |
| map_value7 | [Proto3Message.MapValue7Entry](#hello.Proto3Message.MapValue7Entry) | repeated |  |
| map_value8 | [Proto3Message.MapValue8Entry](#hello.Proto3Message.MapValue8Entry) | repeated |  |
| map_value9 | [Proto3Message.MapValue9Entry](#hello.Proto3Message.MapValue9Entry) | repeated |  |
| map_value10 | [Proto3Message.MapValue10Entry](#hello.Proto3Message.MapValue10Entry) | repeated |  |
| map_value12 | [Proto3Message.MapValue12Entry](#hello.Proto3Message.MapValue12Entry) | repeated |  |
| map_value14 | [Proto3Message.MapValue14Entry](#hello.Proto3Message.MapValue14Entry) | repeated |  |
| map_value15 | [Proto3Message.MapValue15Entry](#hello.Proto3Message.MapValue15Entry) | repeated |  |
| map_value16 | [Proto3Message.MapValue16Entry](#hello.Proto3Message.MapValue16Entry) | repeated |  |
| details | [google.protobuf.Any](#google.protobuf.Any) | repeated |  |






<a name="hello.Proto3Message.MapValue10Entry"></a>

### Proto3Message.MapValue10Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [float](#float) |  |  |






<a name="hello.Proto3Message.MapValue12Entry"></a>

### Proto3Message.MapValue12Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [double](#double) |  |  |






<a name="hello.Proto3Message.MapValue14Entry"></a>

### Proto3Message.MapValue14Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bool](#bool) |  |  |






<a name="hello.Proto3Message.MapValue15Entry"></a>

### Proto3Message.MapValue15Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [bool](#bool) |  |  |
| value | [string](#string) |  |  |






<a name="hello.Proto3Message.MapValue16Entry"></a>

### Proto3Message.MapValue16Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.UInt64Value](#google.protobuf.UInt64Value) |  |  |






<a name="hello.Proto3Message.MapValue2Entry"></a>

### Proto3Message.MapValue2Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [int32](#int32) |  |  |






<a name="hello.Proto3Message.MapValue3Entry"></a>

### Proto3Message.MapValue3Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [int32](#int32) |  |  |
| value | [string](#string) |  |  |






<a name="hello.Proto3Message.MapValue4Entry"></a>

### Proto3Message.MapValue4Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [int64](#int64) |  |  |






<a name="hello.Proto3Message.MapValue5Entry"></a>

### Proto3Message.MapValue5Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [int64](#int64) |  |  |
| value | [string](#string) |  |  |






<a name="hello.Proto3Message.MapValue6Entry"></a>

### Proto3Message.MapValue6Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [uint32](#uint32) |  |  |






<a name="hello.Proto3Message.MapValue7Entry"></a>

### Proto3Message.MapValue7Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [uint32](#uint32) |  |  |
| value | [string](#string) |  |  |






<a name="hello.Proto3Message.MapValue8Entry"></a>

### Proto3Message.MapValue8Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [uint64](#uint64) |  |  |






<a name="hello.Proto3Message.MapValue9Entry"></a>

### Proto3Message.MapValue9Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [uint64](#uint64) |  |  |
| value | [string](#string) |  |  |






<a name="hello.Proto3Message.MapValueEntry"></a>

### Proto3Message.MapValueEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 


<a name="hello.EnumValue"></a>

### EnumValue


| Name | Number | Description |
| ---- | ------ | ----------- |
| X | 0 |  |
| Y | 1 |  |
| Z | 2 |  |


 

 

 



<a name="proto/hello/transport.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/hello/transport.proto



<a name="hello.Message"></a>

### Message



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| header | [Message.HeaderEntry](#hello.Message.HeaderEntry) | repeated |  |
| body | [bytes](#bytes) |  |  |






<a name="hello.Message.HeaderEntry"></a>

### Message.HeaderEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 


<a name="hello.Transport"></a>

### Transport


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| TestStream | [Message](#hello.Message) stream | [Message](#hello.Message) stream |  |
| TestStream1 | [Message](#hello.Message) stream | [Message](#hello.Message) |  |
| TestStream2 | [Message](#hello.Message) | [Message](#hello.Message) stream |  |
| TestStream3 | [Message](#hello.Message) | [Message](#hello.Message) |  |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

