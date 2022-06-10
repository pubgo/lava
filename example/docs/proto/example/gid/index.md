# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [example/gid/a_bit_of_everything.proto](#example_gid_a_bit_of_everything-proto)
    - [LoginReply](#gid-LoginReply)
    - [LoginRequest](#gid-LoginRequest)
    - [LogoutReply](#gid-LogoutReply)
    - [LogoutRequest](#gid-LogoutRequest)
  
    - [LoginService](#gid-LoginService)
  
- [example/gid/echo_service.proto](#example_gid_echo_service-proto)
    - [DynamicMessage](#gid-DynamicMessage)
    - [DynamicMessageUpdate](#gid-DynamicMessageUpdate)
    - [Embedded](#gid-Embedded)
    - [SimpleMessage](#gid-SimpleMessage)
  
    - [EchoService](#gid-EchoService)
  
- [example/gid/id.proto](#example_gid_id-proto)
    - [GenerateRequest](#gid-GenerateRequest)
    - [GenerateResponse](#gid-GenerateResponse)
    - [Tag](#gid-Tag)
    - [TypesRequest](#gid-TypesRequest)
    - [TypesResponse](#gid-TypesResponse)
  
    - [Id](#gid-Id)
  
- [Scalar Value Types](#scalar-value-types)



<a name="example_gid_a_bit_of_everything-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## example/gid/a_bit_of_everything.proto



<a name="gid-LoginReply"></a>

### LoginReply



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| access | [bool](#bool) |  | Whether you have access or not |






<a name="gid-LoginRequest"></a>

### LoginRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| username | [string](#string) |  | The entered username |
| password | [string](#string) |  | The entered password |






<a name="gid-LogoutReply"></a>

### LogoutReply



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  | Message that tells you whether your logout was succesful or not |






<a name="gid-LogoutRequest"></a>

### LogoutRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timeoflogout | [string](#string) |  | The time the logout was registered |
| test | [int32](#int32) |  | This is the title

This is the &#34;Description&#34; of field test you can use as many newlines as you want

it will still format the same in the table |
| stringarray | [string](#string) | repeated | This is an array

It displays that using [] infront of the type |





 

 

 


<a name="gid-LoginService"></a>

### LoginService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Login | [LoginRequest](#gid-LoginRequest) | [LoginReply](#gid-LoginReply) | Login

{{.MethodDescriptorProto.Name}} is a call with the method(s) {{$first := true}}{{range .Bindings}}{{if $first}}{{$first = false}}{{else}}, {{end}}{{.HTTPMethod}}{{end}} within the &#34;{{.Service.Name}}&#34; service. It takes in &#34;{{.RequestType.Name}}&#34; and returns a &#34;{{.ResponseType.Name}}&#34;.

## {{.RequestType.Name}} | Field ID | Name | Type | Description | | ----------- | --------- | --------------------------------------------------------- | ---------------------------- | {{range .RequestType.Fields}} | {{.Number}} | {{.Name}} | {{if eq .Label.String &#34;LABEL_REPEATED&#34;}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## {{.ResponseType.Name}} | Field ID | Name | Type | Description | | ----------- | --------- | ---------------------------------------------------------- | ---------------------------- | {{range .ResponseType.Fields}} | {{.Number}} | {{.Name}} | {{if eq .Label.String &#34;LABEL_REPEATED&#34;}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}} |
| Logout | [LogoutRequest](#gid-LogoutRequest) | [LogoutReply](#gid-LogoutReply) | Logout

{{.MethodDescriptorProto.Name}} is a call with the method(s) {{$first := true}}{{range .Bindings}}{{if $first}}{{$first = false}}{{else}}, {{end}}{{.HTTPMethod}}{{end}} within the &#34;{{.Service.Name}}&#34; service. It takes in &#34;{{.RequestType.Name}}&#34; and returns a &#34;{{.ResponseType.Name}}&#34;.

## {{.RequestType.Name}} | Field ID | Name | Type | Description | | ----------- | --------- | --------------------------------------------------------- | ---------------------------- | {{range .RequestType.Fields}} | {{.Number}} | {{.Name}} | {{if eq .Label.String &#34;LABEL_REPEATED&#34;}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## {{.ResponseType.Name}} | Field ID | Name | Type | Description | | ----------- | --------- | ---------------------------------------------------------- | ---------------------------- | {{range .ResponseType.Fields}} | {{.Number}} | {{.Name}} | {{if eq .Label.String &#34;LABEL_REPEATED&#34;}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}} |

 



<a name="example_gid_echo_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## example/gid/echo_service.proto



<a name="gid-DynamicMessage"></a>

### DynamicMessage
DynamicMessage represents a message which can have its structure
built dynamically using Struct and Values.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| struct_field | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |
| value_field | [google.protobuf.Value](#google-protobuf-Value) |  |  |






<a name="gid-DynamicMessageUpdate"></a>

### DynamicMessageUpdate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| body | [DynamicMessage](#gid-DynamicMessage) |  | google.protobuf.FieldMask update_mask = 2; |






<a name="gid-Embedded"></a>

### Embedded
Embedded represents a message embedded in SimpleMessage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| progress | [int64](#int64) |  |  |
| note | [string](#string) |  |  |






<a name="gid-SimpleMessage"></a>

### SimpleMessage
SimpleMessage represents a simple message sent to the Echo service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | Id represents the message identifier. |
| num | [int64](#int64) |  |  |
| line_num | [int64](#int64) |  |  |
| lang | [string](#string) |  |  |
| status | [Embedded](#gid-Embedded) |  |  |
| en | [int64](#int64) |  |  |
| no | [Embedded](#gid-Embedded) |  |  |





 

 

 


<a name="gid-EchoService"></a>

### EchoService
Echo service responds to incoming echo requests.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Echo | [SimpleMessage](#gid-SimpleMessage) | [SimpleMessage](#gid-SimpleMessage) | Echo method receives a simple message and returns it.

The message posted as the id parameter will also be returned. |
| EchoBody | [SimpleMessage](#gid-SimpleMessage) | [SimpleMessage](#gid-SimpleMessage) | EchoBody method receives a simple message and returns it. |
| EchoDelete | [SimpleMessage](#gid-SimpleMessage) | [SimpleMessage](#gid-SimpleMessage) | EchoDelete method receives a simple message and returns it. |
| EchoPatch | [DynamicMessageUpdate](#gid-DynamicMessageUpdate) | [DynamicMessageUpdate](#gid-DynamicMessageUpdate) | EchoPatch method receives a NonStandardUpdateRequest and returns it. |
| EchoUnauthorized | [SimpleMessage](#gid-SimpleMessage) | [SimpleMessage](#gid-SimpleMessage) | EchoUnauthorized method receives a simple message and returns it. It must always return a google.rpc.Code of `UNAUTHENTICATED` and a HTTP Status code of 401. |

 



<a name="example_gid_id-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## example/gid/id.proto



<a name="gid-GenerateRequest"></a>

### GenerateRequest
Generate a unique ID. Defaults to uuid.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) |  | type of id e.g uuid, shortid, snowflake (64 bit), bigflake (128 bit) |






<a name="gid-GenerateResponse"></a>

### GenerateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | the unique id generated |
| type | [string](#string) |  | the type of id generated |






<a name="gid-Tag"></a>

### Tag



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="gid-TypesRequest"></a>

### TypesRequest
List the types of IDs available. No query params needed.






<a name="gid-TypesResponse"></a>

### TypesResponse
TypesResponse 返回值类型


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| types | [string](#string) | repeated |  |





 

 

 


<a name="gid-Id"></a>

### Id
Id 生成随机ID服务

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Generate | [GenerateRequest](#gid-GenerateRequest) | [GenerateResponse](#gid-GenerateResponse) | Generate 生成ID |
| Types | [TypesRequest](#gid-TypesRequest) | [TypesResponse](#gid-TypesResponse) | Types id类型 |

 



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

