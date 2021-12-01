# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [proto/yuque/yuque.proto](#proto/yuque/yuque.proto)
    - [CreateGroupReq](#yuque.v2.CreateGroupReq)
    - [CreateGroupResp](#yuque.v2.CreateGroupResp)
    - [CreateGroupResp.Data](#yuque.v2.CreateGroupResp.Data)
    - [UserInfoReq](#yuque.v2.UserInfoReq)
    - [UserInfoResp](#yuque.v2.UserInfoResp)
    - [UserInfoResp.Data](#yuque.v2.UserInfoResp.Data)
  
    - [Yuque](#yuque.v2.Yuque)
  
- [Scalar Value Types](#scalar-value-types)



<a name="proto/yuque/yuque.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/yuque/yuque.proto



<a name="yuque.v2.CreateGroupReq"></a>

### CreateGroupReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| login | [string](#string) |  | login |
| name | [string](#string) |  | 组织名称 |
| description | [string](#string) |  | 介绍 |






<a name="yuque.v2.CreateGroupResp"></a>

### CreateGroupResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [CreateGroupResp.Data](#yuque.v2.CreateGroupResp.Data) | repeated |  |
| response | [lava.Response](#lava.Response) |  |  |






<a name="yuque.v2.CreateGroupResp.Data"></a>

### CreateGroupResp.Data



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [uint32](#uint32) |  |  |
| login | [string](#string) |  |  |
| name | [string](#string) |  |  |
| avatar_url | [string](#string) |  |  |
| description | [string](#string) |  |  |
| created_at | [string](#string) |  |  |
| updated_at | [string](#string) |  |  |






<a name="yuque.v2.UserInfoReq"></a>

### UserInfoReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| login | [string](#string) |  |  |
| id | [string](#string) |  |  |






<a name="yuque.v2.UserInfoResp"></a>

### UserInfoResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [UserInfoResp.Data](#yuque.v2.UserInfoResp.Data) |  |  |
| response | [lava.Response](#lava.Response) |  |  |






<a name="yuque.v2.UserInfoResp.Data"></a>

### UserInfoResp.Data



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [uint32](#uint32) |  |  |
| type | [string](#string) |  |  |
| login | [string](#string) |  |  |
| name | [string](#string) |  |  |
| description | [string](#string) |  |  |
| avatar_url | [string](#string) |  |  |
| created_at | [string](#string) |  |  |
| updated_at | [string](#string) |  |  |





 

 

 


<a name="yuque.v2.Yuque"></a>

### Yuque


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| UserInfo | [.google.protobuf.Empty](#google.protobuf.Empty) | [UserInfoResp](#yuque.v2.UserInfoResp) | 获取认证的用户的个人信息 |
| UserInfoByLogin | [UserInfoReq](#yuque.v2.UserInfoReq) | [UserInfoResp](#yuque.v2.UserInfoResp) | 获取单个用户信息 |
| UserInfoById | [UserInfoReq](#yuque.v2.UserInfoReq) | [UserInfoResp](#yuque.v2.UserInfoResp) | 获取单个用户信息 |
| CreateGroup | [CreateGroupReq](#yuque.v2.CreateGroupReq) | [CreateGroupResp](#yuque.v2.CreateGroupResp) | 创建 Group |

 



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

