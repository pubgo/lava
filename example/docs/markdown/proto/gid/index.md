# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [proto/gid/a_bit_of_everything.proto](#proto/gid/a_bit_of_everything.proto)
    - [LoginReply](#gid.LoginReply)
    - [LoginRequest](#gid.LoginRequest)
    - [LogoutReply](#gid.LogoutReply)
    - [LogoutRequest](#gid.LogoutRequest)
  
    - [LoginService](#gid.LoginService)
  
- [proto/gid/echo_service.proto](#proto/gid/echo_service.proto)
    - [DynamicMessage](#gid.DynamicMessage)
    - [DynamicMessageUpdate](#gid.DynamicMessageUpdate)
    - [Embedded](#gid.Embedded)
    - [SimpleMessage](#gid.SimpleMessage)
  
    - [EchoService](#gid.EchoService)
  
- [proto/gid/id.proto](#proto/gid/id.proto)
    - [ABitOfEverything](#gid.ABitOfEverything)
    - [ABitOfEverything.MapValueEntry](#gid.ABitOfEverything.MapValueEntry)
    - [ABitOfEverything.MappedNestedValueEntry](#gid.ABitOfEverything.MappedNestedValueEntry)
    - [ABitOfEverything.MappedStringValueEntry](#gid.ABitOfEverything.MappedStringValueEntry)
    - [ABitOfEverything.Nested](#gid.ABitOfEverything.Nested)
    - [ABitOfEverythingRepeated](#gid.ABitOfEverythingRepeated)
    - [Body](#gid.Body)
    - [Book](#gid.Book)
    - [CheckStatusResponse](#gid.CheckStatusResponse)
    - [CreateBookRequest](#gid.CreateBookRequest)
    - [ErrorObject](#gid.ErrorObject)
    - [ErrorResponse](#gid.ErrorResponse)
    - [GenerateRequest](#gid.GenerateRequest)
    - [GenerateResponse](#gid.GenerateResponse)
    - [MessageWithBody](#gid.MessageWithBody)
    - [Tag](#gid.Tag)
    - [TypesRequest](#gid.TypesRequest)
    - [TypesResponse](#gid.TypesResponse)
    - [UpdateBookRequest](#gid.UpdateBookRequest)
    - [UpdateV2Request](#gid.UpdateV2Request)
  
    - [ABitOfEverything.Nested.DeepEnum](#gid.ABitOfEverything.Nested.DeepEnum)
    - [NumericEnum](#gid.NumericEnum)
  
    - [ABitOfEverythingService](#gid.ABitOfEverythingService)
    - [AnotherServiceWithNoBindings](#gid.AnotherServiceWithNoBindings)
    - [Id](#gid.Id)
    - [camelCaseServiceName](#gid.camelCaseServiceName)
  
- [Scalar Value Types](#scalar-value-types)



<a name="proto/gid/a_bit_of_everything.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/gid/a_bit_of_everything.proto



<a name="gid.LoginReply"></a>

### LoginReply



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| access | [bool](#bool) |  | Whether you have access or not |






<a name="gid.LoginRequest"></a>

### LoginRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| username | [string](#string) |  | The entered username |
| password | [string](#string) |  | The entered password |






<a name="gid.LogoutReply"></a>

### LogoutReply



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  | Message that tells you whether your logout was succesful or not |






<a name="gid.LogoutRequest"></a>

### LogoutRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timeoflogout | [string](#string) |  | The time the logout was registered |
| test | [int32](#int32) |  | This is the title

This is the &#34;Description&#34; of field test you can use as many newlines as you want

it will still format the same in the table |
| stringarray | [string](#string) | repeated | This is an array

It displays that using [] infront of the type |





 

 

 


<a name="gid.LoginService"></a>

### LoginService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Login | [LoginRequest](#gid.LoginRequest) | [LoginReply](#gid.LoginReply) | Login

{{.MethodDescriptorProto.Name}} is a call with the method(s) {{$first := true}}{{range .Bindings}}{{if $first}}{{$first = false}}{{else}}, {{end}}{{.HTTPMethod}}{{end}} within the &#34;{{.Service.Name}}&#34; service. It takes in &#34;{{.RequestType.Name}}&#34; and returns a &#34;{{.ResponseType.Name}}&#34;.

## {{.RequestType.Name}} | Field ID | Name | Type | Description | | ----------- | --------- | --------------------------------------------------------- | ---------------------------- | {{range .RequestType.Fields}} | {{.Number}} | {{.Name}} | {{if eq .Label.String &#34;LABEL_REPEATED&#34;}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## {{.ResponseType.Name}} | Field ID | Name | Type | Description | | ----------- | --------- | ---------------------------------------------------------- | ---------------------------- | {{range .ResponseType.Fields}} | {{.Number}} | {{.Name}} | {{if eq .Label.String &#34;LABEL_REPEATED&#34;}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}} |
| Logout | [LogoutRequest](#gid.LogoutRequest) | [LogoutReply](#gid.LogoutReply) | Logout

{{.MethodDescriptorProto.Name}} is a call with the method(s) {{$first := true}}{{range .Bindings}}{{if $first}}{{$first = false}}{{else}}, {{end}}{{.HTTPMethod}}{{end}} within the &#34;{{.Service.Name}}&#34; service. It takes in &#34;{{.RequestType.Name}}&#34; and returns a &#34;{{.ResponseType.Name}}&#34;.

## {{.RequestType.Name}} | Field ID | Name | Type | Description | | ----------- | --------- | --------------------------------------------------------- | ---------------------------- | {{range .RequestType.Fields}} | {{.Number}} | {{.Name}} | {{if eq .Label.String &#34;LABEL_REPEATED&#34;}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## {{.ResponseType.Name}} | Field ID | Name | Type | Description | | ----------- | --------- | ---------------------------------------------------------- | ---------------------------- | {{range .ResponseType.Fields}} | {{.Number}} | {{.Name}} | {{if eq .Label.String &#34;LABEL_REPEATED&#34;}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}} |

 



<a name="proto/gid/echo_service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/gid/echo_service.proto



<a name="gid.DynamicMessage"></a>

### DynamicMessage
DynamicMessage represents a message which can have its structure
built dynamically using Struct and Values.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| struct_field | [google.protobuf.Struct](#google.protobuf.Struct) |  |  |
| value_field | [google.protobuf.Value](#google.protobuf.Value) |  |  |






<a name="gid.DynamicMessageUpdate"></a>

### DynamicMessageUpdate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| body | [DynamicMessage](#gid.DynamicMessage) |  | google.protobuf.FieldMask update_mask = 2; |






<a name="gid.Embedded"></a>

### Embedded
Embedded represents a message embedded in SimpleMessage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| progress | [int64](#int64) |  |  |
| note | [string](#string) |  |  |






<a name="gid.SimpleMessage"></a>

### SimpleMessage
SimpleMessage represents a simple message sent to the Echo service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | Id represents the message identifier. |
| num | [int64](#int64) |  |  |
| line_num | [int64](#int64) |  |  |
| lang | [string](#string) |  |  |
| status | [Embedded](#gid.Embedded) |  |  |
| en | [int64](#int64) |  |  |
| no | [Embedded](#gid.Embedded) |  |  |





 

 

 


<a name="gid.EchoService"></a>

### EchoService
Echo service responds to incoming echo requests.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Echo | [SimpleMessage](#gid.SimpleMessage) | [SimpleMessage](#gid.SimpleMessage) | Echo method receives a simple message and returns it.

The message posted as the id parameter will also be returned. |
| EchoBody | [SimpleMessage](#gid.SimpleMessage) | [SimpleMessage](#gid.SimpleMessage) | EchoBody method receives a simple message and returns it. |
| EchoDelete | [SimpleMessage](#gid.SimpleMessage) | [SimpleMessage](#gid.SimpleMessage) | EchoDelete method receives a simple message and returns it. |
| EchoPatch | [DynamicMessageUpdate](#gid.DynamicMessageUpdate) | [DynamicMessageUpdate](#gid.DynamicMessageUpdate) | EchoPatch method receives a NonStandardUpdateRequest and returns it. |
| EchoUnauthorized | [SimpleMessage](#gid.SimpleMessage) | [SimpleMessage](#gid.SimpleMessage) | EchoUnauthorized method receives a simple message and returns it. It must always return a google.rpc.Code of `UNAUTHENTICATED` and a HTTP Status code of 401. |

 



<a name="proto/gid/id.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/gid/id.proto



<a name="gid.ABitOfEverything"></a>

### ABitOfEverything
Intentionally complicated message type to cover many features of Protobuf.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| single_nested | [ABitOfEverything.Nested](#gid.ABitOfEverything.Nested) |  |  |
| uuid | [string](#string) |  |  |
| nested | [ABitOfEverything.Nested](#gid.ABitOfEverything.Nested) | repeated |  |
| float_value | [float](#float) |  |  |
| double_value | [double](#double) |  |  |
| int64_value | [int64](#int64) |  |  |
| uint64_value | [uint64](#uint64) |  |  |
| int32_value | [int32](#int32) |  |  |
| fixed64_value | [fixed64](#fixed64) |  |  |
| fixed32_value | [fixed32](#fixed32) |  |  |
| bool_value | [bool](#bool) |  |  |
| string_value | [string](#string) |  |  |
| bytes_value | [bytes](#bytes) |  |  |
| uint32_value | [uint32](#uint32) |  |  |
| enum_value | [NumericEnum](#gid.NumericEnum) |  |  |
| sfixed32_value | [sfixed32](#sfixed32) |  |  |
| sfixed64_value | [sfixed64](#sfixed64) |  |  |
| sint32_value | [sint32](#sint32) |  |  |
| sint64_value | [sint64](#sint64) |  |  |
| repeated_string_value | [string](#string) | repeated |  |
| oneof_empty | [google.protobuf.Empty](#google.protobuf.Empty) |  |  |
| oneof_string | [string](#string) |  |  |
| map_value | [ABitOfEverything.MapValueEntry](#gid.ABitOfEverything.MapValueEntry) | repeated |  |
| mapped_string_value | [ABitOfEverything.MappedStringValueEntry](#gid.ABitOfEverything.MappedStringValueEntry) | repeated |  |
| mapped_nested_value | [ABitOfEverything.MappedNestedValueEntry](#gid.ABitOfEverything.MappedNestedValueEntry) | repeated |  |
| nonConventionalNameValue | [string](#string) |  |  |
| timestamp_value | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| repeated_enum_value | [NumericEnum](#gid.NumericEnum) | repeated | repeated enum value. it is comma-separated in query |
| repeated_enum_annotation | [NumericEnum](#gid.NumericEnum) | repeated | repeated numeric enum comment (This comment is overridden by the field annotation) |
| enum_value_annotation | [NumericEnum](#gid.NumericEnum) |  | numeric enum comment (This comment is overridden by the field annotation) |
| repeated_string_annotation | [string](#string) | repeated | repeated string comment (This comment is overridden by the field annotation) |
| repeated_nested_annotation | [ABitOfEverything.Nested](#gid.ABitOfEverything.Nested) | repeated | repeated nested object comment (This comment is overridden by the field annotation) |
| nested_annotation | [ABitOfEverything.Nested](#gid.ABitOfEverything.Nested) |  | nested object comments (This comment is overridden by the field annotation) |
| int64_override_type | [int64](#int64) |  |  |
| required_string_via_field_behavior_annotation | [string](#string) |  | mark a field as required in Open API definition |
| output_only_string_via_field_behavior_annotation | [string](#string) |  | mark a field as readonly in Open API definition |






<a name="gid.ABitOfEverything.MapValueEntry"></a>

### ABitOfEverything.MapValueEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [NumericEnum](#gid.NumericEnum) |  |  |






<a name="gid.ABitOfEverything.MappedNestedValueEntry"></a>

### ABitOfEverything.MappedNestedValueEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [ABitOfEverything.Nested](#gid.ABitOfEverything.Nested) |  |  |






<a name="gid.ABitOfEverything.MappedStringValueEntry"></a>

### ABitOfEverything.MappedStringValueEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="gid.ABitOfEverything.Nested"></a>

### ABitOfEverything.Nested
Nested is nested type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is nested field. |
| amount | [uint32](#uint32) |  |  |
| ok | [ABitOfEverything.Nested.DeepEnum](#gid.ABitOfEverything.Nested.DeepEnum) |  | DeepEnum comment. |






<a name="gid.ABitOfEverythingRepeated"></a>

### ABitOfEverythingRepeated
ABitOfEverythingRepeated is used to validate repeated path parameter functionality


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path_repeated_float_value | [float](#float) | repeated | repeated values. they are comma-separated in path |
| path_repeated_double_value | [double](#double) | repeated |  |
| path_repeated_int64_value | [int64](#int64) | repeated |  |
| path_repeated_uint64_value | [uint64](#uint64) | repeated |  |
| path_repeated_int32_value | [int32](#int32) | repeated |  |
| path_repeated_fixed64_value | [fixed64](#fixed64) | repeated |  |
| path_repeated_fixed32_value | [fixed32](#fixed32) | repeated |  |
| path_repeated_bool_value | [bool](#bool) | repeated |  |
| path_repeated_string_value | [string](#string) | repeated |  |
| path_repeated_bytes_value | [bytes](#bytes) | repeated |  |
| path_repeated_uint32_value | [uint32](#uint32) | repeated |  |
| path_repeated_enum_value | [NumericEnum](#gid.NumericEnum) | repeated |  |
| path_repeated_sfixed32_value | [sfixed32](#sfixed32) | repeated |  |
| path_repeated_sfixed64_value | [sfixed64](#sfixed64) | repeated |  |
| path_repeated_sint32_value | [sint32](#sint32) | repeated |  |
| path_repeated_sint64_value | [sint64](#sint64) | repeated |  |






<a name="gid.Body"></a>

### Body



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name="gid.Book"></a>

### Book
An example resource type from AIP-123 used to test the behavior described in
the CreateBookRequest message.

See: https://google.aip.dev/123


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | The resource name of the book.

Format: `publishers/{publisher}/books/{book}`

Example: `publishers/1257894000000000000/books/my-book` |
| id | [string](#string) |  | Output only. The book&#39;s ID. |
| create_time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Output only. Creation time of the book. |






<a name="gid.CheckStatusResponse"></a>

### CheckStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [google.rpc.Status](#google.rpc.Status) |  |  |






<a name="gid.CreateBookRequest"></a>

### CreateBookRequest
A standard Create message from AIP-133 with a user-specified ID.
The user-specified ID (the `book_id` field in this example) must become a
query parameter in the OpenAPI spec.

See: https://google.aip.dev/133#user-specified-ids


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| parent | [string](#string) |  | The publisher in which to create the book.

Format: `publishers/{publisher}`

Example: `publishers/1257894000000000000` |
| book | [Book](#gid.Book) |  | The book to create. |
| book_id | [string](#string) |  | The ID to use for the book.

This must start with an alphanumeric character. |






<a name="gid.ErrorObject"></a>

### ErrorObject



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int32](#int32) |  |  |
| message | [string](#string) |  |  |






<a name="gid.ErrorResponse"></a>

### ErrorResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| correlationId | [string](#string) |  |  |
| error | [ErrorObject](#gid.ErrorObject) |  |  |






<a name="gid.GenerateRequest"></a>

### GenerateRequest
Generate a unique ID. Defaults to uuid.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) |  | type of id e.g uuid, shortid, snowflake (64 bit), bigflake (128 bit) |






<a name="gid.GenerateResponse"></a>

### GenerateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | the unique id generated |
| type | [string](#string) |  | the type of id generated |






<a name="gid.MessageWithBody"></a>

### MessageWithBody



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| data | [Body](#gid.Body) |  |  |






<a name="gid.Tag"></a>

### Tag



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="gid.TypesRequest"></a>

### TypesRequest
List the types of IDs available. No query params needed.






<a name="gid.TypesResponse"></a>

### TypesResponse
TypesResponse 返回值类型


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| types | [string](#string) | repeated |  |






<a name="gid.UpdateBookRequest"></a>

### UpdateBookRequest
A standard Update message from AIP-134

See: https://google.aip.dev/134#request-message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| book | [Book](#gid.Book) |  | The book to update.

The book&#39;s `name` field is used to identify the book to be updated. Format: publishers/{publisher}/books/{book} |
| update_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  | The list of fields to be updated. |
| allow_missing | [bool](#bool) |  | If set to true, and the book is not found, a new book will be created. In this situation, `update_mask` is ignored. |






<a name="gid.UpdateV2Request"></a>

### UpdateV2Request
UpdateV2Request request for update includes the message and the update mask


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| abe | [ABitOfEverything](#gid.ABitOfEverything) |  |  |
| update_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  | The paths to update. |





 


<a name="gid.ABitOfEverything.Nested.DeepEnum"></a>

### ABitOfEverything.Nested.DeepEnum
DeepEnum is one or zero.

| Name | Number | Description |
| ---- | ------ | ----------- |
| FALSE | 0 | FALSE is false. |
| TRUE | 1 | TRUE is true. |



<a name="gid.NumericEnum"></a>

### NumericEnum
NumericEnum is one or zero.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ZERO | 0 | ZERO means 0 |
| ONE | 1 | ONE means 1 |


 

 


<a name="gid.ABitOfEverythingService"></a>

### ABitOfEverythingService
ABitOfEverything service is used to validate that APIs with complicated
proto messages and URL templates are still processed correctly.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Create | [ABitOfEverything](#gid.ABitOfEverything) | [ABitOfEverything](#gid.ABitOfEverything) | Create a new ABitOfEverything

This API creates a new ABitOfEverything |
| CreateBody | [ABitOfEverything](#gid.ABitOfEverything) | [ABitOfEverything](#gid.ABitOfEverything) |  |
| CreateBook | [CreateBookRequest](#gid.CreateBookRequest) | [Book](#gid.Book) | Create a book. |
| UpdateBook | [UpdateBookRequest](#gid.UpdateBookRequest) | [Book](#gid.Book) |  |
| Update | [ABitOfEverything](#gid.ABitOfEverything) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| UpdateV2 | [UpdateV2Request](#gid.UpdateV2Request) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| GetQuery | [ABitOfEverything](#gid.ABitOfEverything) | [.google.protobuf.Empty](#google.protobuf.Empty) | rpc Delete(grpc.gateway.examples.internal.proto.sub2.IdMessage) returns (google.protobuf.Empty) { option (google.api.http) = { delete: &#34;/v1/example/a_bit_of_everything/{uuid}&#34; }; option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = { security: { security_requirement: { key: &#34;ApiKeyAuth&#34;; value: {} } security_requirement: { key: &#34;OAuth2&#34;; value: { scope: &#34;read&#34;; scope: &#34;write&#34;; } } } extensions: { key: &#34;x-irreversible&#34;; value { bool_value: true; } } }; } |
| GetRepeatedQuery | [ABitOfEverythingRepeated](#gid.ABitOfEverythingRepeated) | [ABitOfEverythingRepeated](#gid.ABitOfEverythingRepeated) |  |
| DeepPathEcho | [ABitOfEverything](#gid.ABitOfEverything) | [ABitOfEverything](#gid.ABitOfEverything) | Echo allows posting a StringMessage value.

It also exposes multiple bindings.

This makes it useful when validating that the OpenAPI v2 API description exposes documentation correctly on all paths defined as additional_bindings in the proto. rpc Echo(grpc.gateway.examples.internal.proto.sub.StringMessage) returns (grpc.gateway.examples.internal.proto.sub.StringMessage) { option (google.api.http) = { get: &#34;/v1/example/a_bit_of_everything/echo/{value}&#34; additional_bindings { post: &#34;/v2/example/echo&#34; body: &#34;value&#34; } additional_bindings { get: &#34;/v2/example/echo&#34; } }; option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = { description: &#34;Description Echo&#34;; summary: &#34;Summary: Echo rpc&#34;; tags: &#34;echo rpc&#34;; external_docs: { url: &#34;https://github.com/grpc-ecosystem/grpc-gateway&#34;; description: &#34;Find out more Echo&#34;; } responses: { key: &#34;200&#34; value: { examples: { key: &#34;application/json&#34; value: &#39;{&#34;value&#34;: &#34;the input value&#34;}&#39; } } } responses: { key: &#34;503&#34;; value: { description: &#34;Returned when the resource is temporarily unavailable.&#34;; extensions: { key: &#34;x-number&#34;; value { number_value: 100; } } } } responses: { // Overwrites global definition. key: &#34;404&#34;; value: { description: &#34;Returned when the resource does not exist.&#34;; schema: { json_schema: { type: INTEGER; } } } } }; } |
| NoBindings | [.google.protobuf.Duration](#google.protobuf.Duration) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| Timeout | [.google.protobuf.Empty](#google.protobuf.Empty) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| ErrorWithDetails | [.google.protobuf.Empty](#google.protobuf.Empty) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| GetMessageWithBody | [MessageWithBody](#gid.MessageWithBody) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| PostWithEmptyBody | [Body](#gid.Body) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| CheckGetQueryParams | [ABitOfEverything](#gid.ABitOfEverything) | [ABitOfEverything](#gid.ABitOfEverything) |  |
| CheckNestedEnumGetQueryParams | [ABitOfEverything](#gid.ABitOfEverything) | [ABitOfEverything](#gid.ABitOfEverything) |  |
| CheckPostQueryParams | [ABitOfEverything](#gid.ABitOfEverything) | [ABitOfEverything](#gid.ABitOfEverything) |  |
| OverwriteResponseContentType | [.google.protobuf.Empty](#google.protobuf.Empty) | [.google.protobuf.StringValue](#google.protobuf.StringValue) |  |
| CheckStatus | [.google.protobuf.Empty](#google.protobuf.Empty) | [CheckStatusResponse](#gid.CheckStatusResponse) |  |


<a name="gid.AnotherServiceWithNoBindings"></a>

### AnotherServiceWithNoBindings


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| NoBindings | [.google.protobuf.Empty](#google.protobuf.Empty) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |


<a name="gid.Id"></a>

### Id
Id 生成随机ID服务

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Generate | [GenerateRequest](#gid.GenerateRequest) | [GenerateResponse](#gid.GenerateResponse) | Generate 生成ID |
| Types | [TypesRequest](#gid.TypesRequest) | [TypesResponse](#gid.TypesResponse) | Types id类型 |


<a name="gid.camelCaseServiceName"></a>

### camelCaseServiceName
camelCase and lowercase service names are valid but not recommended (use TitleCase instead)

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Empty | [.google.protobuf.Empty](#google.protobuf.Empty) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |

 



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

