# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [proto/user/user.proto](#proto_user_user-proto)
    - [ABitOfEverything](#gid-ABitOfEverything)
    - [ABitOfEverything.MapValueEntry](#gid-ABitOfEverything-MapValueEntry)
    - [ABitOfEverything.MappedNestedValueEntry](#gid-ABitOfEverything-MappedNestedValueEntry)
    - [ABitOfEverything.MappedStringValueEntry](#gid-ABitOfEverything-MappedStringValueEntry)
    - [ABitOfEverything.Nested](#gid-ABitOfEverything-Nested)
    - [ABitOfEverythingRepeated](#gid-ABitOfEverythingRepeated)
    - [Body](#gid-Body)
    - [Book](#gid-Book)
    - [CheckStatusResponse](#gid-CheckStatusResponse)
    - [CreateBookRequest](#gid-CreateBookRequest)
    - [ErrorObject](#gid-ErrorObject)
    - [ErrorResponse](#gid-ErrorResponse)
    - [GenerateRequest](#gid-GenerateRequest)
    - [GenerateResponse](#gid-GenerateResponse)
    - [MessageWithBody](#gid-MessageWithBody)
    - [Tag](#gid-Tag)
    - [TypesRequest](#gid-TypesRequest)
    - [TypesResponse](#gid-TypesResponse)
    - [UpdateBookRequest](#gid-UpdateBookRequest)
    - [UpdateV2Request](#gid-UpdateV2Request)
  
    - [ABitOfEverything.Nested.DeepEnum](#gid-ABitOfEverything-Nested-DeepEnum)
    - [NumericEnum](#gid-NumericEnum)
  
    - [ABitOfEverythingService](#gid-ABitOfEverythingService)
    - [AnotherServiceWithNoBindings](#gid-AnotherServiceWithNoBindings)
    - [User](#gid-User)
    - [camelCaseServiceName](#gid-camelCaseServiceName)
  
- [Scalar Value Types](#scalar-value-types)



<a name="proto_user_user-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/user/user.proto



<a name="gid-ABitOfEverything"></a>

### ABitOfEverything
Intentionally complicated message type to cover many features of Protobuf.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| single_nested | [ABitOfEverything.Nested](#gid-ABitOfEverything-Nested) |  |  |
| uuid | [string](#string) |  |  |
| nested | [ABitOfEverything.Nested](#gid-ABitOfEverything-Nested) | repeated |  |
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
| enum_value | [NumericEnum](#gid-NumericEnum) |  |  |
| sfixed32_value | [sfixed32](#sfixed32) |  |  |
| sfixed64_value | [sfixed64](#sfixed64) |  |  |
| sint32_value | [sint32](#sint32) |  |  |
| sint64_value | [sint64](#sint64) |  |  |
| repeated_string_value | [string](#string) | repeated |  |
| oneof_empty | [google.protobuf.Empty](#google-protobuf-Empty) |  |  |
| oneof_string | [string](#string) |  |  |
| map_value | [ABitOfEverything.MapValueEntry](#gid-ABitOfEverything-MapValueEntry) | repeated |  |
| mapped_string_value | [ABitOfEverything.MappedStringValueEntry](#gid-ABitOfEverything-MappedStringValueEntry) | repeated |  |
| mapped_nested_value | [ABitOfEverything.MappedNestedValueEntry](#gid-ABitOfEverything-MappedNestedValueEntry) | repeated |  |
| nonConventionalNameValue | [string](#string) |  |  |
| timestamp_value | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| repeated_enum_value | [NumericEnum](#gid-NumericEnum) | repeated | repeated enum value. it is comma-separated in query |
| repeated_enum_annotation | [NumericEnum](#gid-NumericEnum) | repeated | repeated numeric enum comment (This comment is overridden by the field annotation) |
| enum_value_annotation | [NumericEnum](#gid-NumericEnum) |  | numeric enum comment (This comment is overridden by the field annotation) |
| repeated_string_annotation | [string](#string) | repeated | repeated string comment (This comment is overridden by the field annotation) |
| repeated_nested_annotation | [ABitOfEverything.Nested](#gid-ABitOfEverything-Nested) | repeated | repeated nested object comment (This comment is overridden by the field annotation) |
| nested_annotation | [ABitOfEverything.Nested](#gid-ABitOfEverything-Nested) |  | nested object comments (This comment is overridden by the field annotation) |
| int64_override_type | [int64](#int64) |  |  |
| required_string_via_field_behavior_annotation | [string](#string) |  | mark a field as required in Open API definition |
| output_only_string_via_field_behavior_annotation | [string](#string) |  | mark a field as readonly in Open API definition |






<a name="gid-ABitOfEverything-MapValueEntry"></a>

### ABitOfEverything.MapValueEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [NumericEnum](#gid-NumericEnum) |  |  |






<a name="gid-ABitOfEverything-MappedNestedValueEntry"></a>

### ABitOfEverything.MappedNestedValueEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [ABitOfEverything.Nested](#gid-ABitOfEverything-Nested) |  |  |






<a name="gid-ABitOfEverything-MappedStringValueEntry"></a>

### ABitOfEverything.MappedStringValueEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="gid-ABitOfEverything-Nested"></a>

### ABitOfEverything.Nested
Nested is nested type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is nested field. |
| amount | [uint32](#uint32) |  |  |
| ok | [ABitOfEverything.Nested.DeepEnum](#gid-ABitOfEverything-Nested-DeepEnum) |  | DeepEnum comment. |






<a name="gid-ABitOfEverythingRepeated"></a>

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
| path_repeated_enum_value | [NumericEnum](#gid-NumericEnum) | repeated |  |
| path_repeated_sfixed32_value | [sfixed32](#sfixed32) | repeated |  |
| path_repeated_sfixed64_value | [sfixed64](#sfixed64) | repeated |  |
| path_repeated_sint32_value | [sint32](#sint32) | repeated |  |
| path_repeated_sint64_value | [sint64](#sint64) | repeated |  |






<a name="gid-Body"></a>

### Body



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name="gid-Book"></a>

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
| create_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Output only. Creation time of the book. |






<a name="gid-CheckStatusResponse"></a>

### CheckStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [google.rpc.Status](#google-rpc-Status) |  |  |






<a name="gid-CreateBookRequest"></a>

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
| book | [Book](#gid-Book) |  | The book to create. |
| book_id | [string](#string) |  | The ID to use for the book.

This must start with an alphanumeric character. |






<a name="gid-ErrorObject"></a>

### ErrorObject



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int32](#int32) |  |  |
| message | [string](#string) |  |  |






<a name="gid-ErrorResponse"></a>

### ErrorResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| correlationId | [string](#string) |  |  |
| error | [ErrorObject](#gid-ErrorObject) |  |  |






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






<a name="gid-MessageWithBody"></a>

### MessageWithBody



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| data | [Body](#gid-Body) |  |  |






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






<a name="gid-UpdateBookRequest"></a>

### UpdateBookRequest
A standard Update message from AIP-134

See: https://google.aip.dev/134#request-message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| book | [Book](#gid-Book) |  | The book to update.

The book&#39;s `name` field is used to identify the book to be updated. Format: publishers/{publisher}/books/{book} |
| update_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | The list of fields to be updated. |
| allow_missing | [bool](#bool) |  | If set to true, and the book is not found, a new book will be created. In this situation, `update_mask` is ignored. |






<a name="gid-UpdateV2Request"></a>

### UpdateV2Request
UpdateV2Request request for update includes the message and the update mask


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| abe | [ABitOfEverything](#gid-ABitOfEverything) |  |  |
| update_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | The paths to update. |





 


<a name="gid-ABitOfEverything-Nested-DeepEnum"></a>

### ABitOfEverything.Nested.DeepEnum
DeepEnum is one or zero.

| Name | Number | Description |
| ---- | ------ | ----------- |
| FALSE | 0 | FALSE is false. |
| TRUE | 1 | TRUE is true. |



<a name="gid-NumericEnum"></a>

### NumericEnum
NumericEnum is one or zero.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ZERO | 0 | ZERO means 0 |
| ONE | 1 | ONE means 1 |


 

 


<a name="gid-ABitOfEverythingService"></a>

### ABitOfEverythingService
ABitOfEverything service is used to validate that APIs with complicated
proto messages and URL templates are still processed correctly.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Create | [ABitOfEverything](#gid-ABitOfEverything) | [ABitOfEverything](#gid-ABitOfEverything) | Create a new ABitOfEverything

This API creates a new ABitOfEverything |
| CreateBody | [ABitOfEverything](#gid-ABitOfEverything) | [ABitOfEverything](#gid-ABitOfEverything) |  |
| CreateBook | [CreateBookRequest](#gid-CreateBookRequest) | [Book](#gid-Book) | Create a book. |
| UpdateBook | [UpdateBookRequest](#gid-UpdateBookRequest) | [Book](#gid-Book) |  |
| Update | [ABitOfEverything](#gid-ABitOfEverything) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| UpdateV2 | [UpdateV2Request](#gid-UpdateV2Request) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| GetQuery | [ABitOfEverything](#gid-ABitOfEverything) | [.google.protobuf.Empty](#google-protobuf-Empty) | rpc Delete(grpc.gateway.examples.internal.proto.sub2.IdMessage) returns (google.protobuf.Empty) { option (google.api.http) = { delete: &#34;/v1/example/a_bit_of_everything/{uuid}&#34; }; option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = { security: { security_requirement: { key: &#34;ApiKeyAuth&#34;; value: {} } security_requirement: { key: &#34;OAuth2&#34;; value: { scope: &#34;read&#34;; scope: &#34;write&#34;; } } } extensions: { key: &#34;x-irreversible&#34;; value { bool_value: true; } } }; } |
| GetRepeatedQuery | [ABitOfEverythingRepeated](#gid-ABitOfEverythingRepeated) | [ABitOfEverythingRepeated](#gid-ABitOfEverythingRepeated) |  |
| DeepPathEcho | [ABitOfEverything](#gid-ABitOfEverything) | [ABitOfEverything](#gid-ABitOfEverything) | Echo allows posting a StringMessage value.

It also exposes multiple bindings.

This makes it useful when validating that the OpenAPI v2 API description exposes documentation correctly on all paths defined as additional_bindings in the proto. rpc Echo(grpc.gateway.examples.internal.proto.sub.StringMessage) returns (grpc.gateway.examples.internal.proto.sub.StringMessage) { option (google.api.http) = { get: &#34;/v1/example/a_bit_of_everything/echo/{value}&#34; additional_bindings { post: &#34;/v2/example/echo&#34; body: &#34;value&#34; } additional_bindings { get: &#34;/v2/example/echo&#34; } }; option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = { description: &#34;Description Echo&#34;; summary: &#34;Summary: Echo rpc&#34;; tags: &#34;echo rpc&#34;; external_docs: { url: &#34;https://github.com/grpc-ecosystem/grpc-gateway&#34;; description: &#34;Find out more Echo&#34;; } responses: { key: &#34;200&#34; value: { examples: { key: &#34;application/json&#34; value: &#39;{&#34;value&#34;: &#34;the input value&#34;}&#39; } } } responses: { key: &#34;503&#34;; value: { description: &#34;Returned when the resource is temporarily unavailable.&#34;; extensions: { key: &#34;x-number&#34;; value { number_value: 100; } } } } responses: { // Overwrites global definition. key: &#34;404&#34;; value: { description: &#34;Returned when the resource does not exist.&#34;; schema: { json_schema: { type: INTEGER; } } } } }; } |
| NoBindings | [.google.protobuf.Duration](#google-protobuf-Duration) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| Timeout | [.google.protobuf.Empty](#google-protobuf-Empty) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| ErrorWithDetails | [.google.protobuf.Empty](#google-protobuf-Empty) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| GetMessageWithBody | [MessageWithBody](#gid-MessageWithBody) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| PostWithEmptyBody | [Body](#gid-Body) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| CheckGetQueryParams | [ABitOfEverything](#gid-ABitOfEverything) | [ABitOfEverything](#gid-ABitOfEverything) |  |
| CheckNestedEnumGetQueryParams | [ABitOfEverything](#gid-ABitOfEverything) | [ABitOfEverything](#gid-ABitOfEverything) |  |
| CheckPostQueryParams | [ABitOfEverything](#gid-ABitOfEverything) | [ABitOfEverything](#gid-ABitOfEverything) |  |
| OverwriteResponseContentType | [.google.protobuf.Empty](#google-protobuf-Empty) | [.google.protobuf.StringValue](#google-protobuf-StringValue) |  |
| CheckStatus | [.google.protobuf.Empty](#google-protobuf-Empty) | [CheckStatusResponse](#gid-CheckStatusResponse) |  |


<a name="gid-AnotherServiceWithNoBindings"></a>

### AnotherServiceWithNoBindings


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| NoBindings | [.google.protobuf.Empty](#google-protobuf-Empty) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |


<a name="gid-User"></a>

### User
User 用户服务

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Generate | [GenerateRequest](#gid-GenerateRequest) | [GenerateResponse](#gid-GenerateResponse) | Generate 生成ID |
| Types | [TypesRequest](#gid-TypesRequest) | [TypesResponse](#gid-TypesResponse) | Types id类型 |


<a name="gid-camelCaseServiceName"></a>

### camelCaseServiceName
camelCase and lowercase service names are valid but not recommended (use TitleCase instead)

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Empty | [.google.protobuf.Empty](#google-protobuf-Empty) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |

 



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

