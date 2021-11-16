# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [proto/sqlx/code.proto](#proto/sqlx/code.proto)
    - [GetSendStatusRequest](#sqlx.GetSendStatusRequest)
    - [GetSendStatusResponse](#sqlx.GetSendStatusResponse)
    - [IsCheckImageCodeRequest](#sqlx.IsCheckImageCodeRequest)
    - [IsCheckImageCodeResponse](#sqlx.IsCheckImageCodeResponse)
    - [SendCodeRequest](#sqlx.SendCodeRequest)
    - [SendCodeResponse](#sqlx.SendCodeResponse)
    - [SendCodeResponse.DataEntry](#sqlx.SendCodeResponse.DataEntry)
    - [SendStatus](#sqlx.SendStatus)
    - [VerifyImageCodeRequest](#sqlx.VerifyImageCodeRequest)
    - [VerifyImageCodeResponse](#sqlx.VerifyImageCodeResponse)
    - [VerifyRequest](#sqlx.VerifyRequest)
    - [VerifyResponse](#sqlx.VerifyResponse)
    - [VerifyResponse.DataEntry](#sqlx.VerifyResponse.DataEntry)
  
    - [Code](#sqlx.Code)
  
- [Scalar Value Types](#scalar-value-types)



<a name="proto/sqlx/code.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/sqlx/code.proto



<a name="sqlx.GetSendStatusRequest"></a>

### GetSendStatusRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| sendType | [string](#string) |  | 发送类型 |
| template | [string](#string) |  | 模板 |
| signR | [int64](#int64) |  | 是否越狱标示 |
| ip | [string](#string) |  | ip |






<a name="sqlx.GetSendStatusResponse"></a>

### GetSendStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [SendStatus](#sqlx.SendStatus) |  | 数据 |






<a name="sqlx.IsCheckImageCodeRequest"></a>

### IsCheckImageCodeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| scene | [string](#string) |  | 场景 |






<a name="sqlx.IsCheckImageCodeResponse"></a>

### IsCheckImageCodeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [bool](#bool) |  | 数据 |






<a name="sqlx.SendCodeRequest"></a>

### SendCodeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 电话 |
| sendType | [string](#string) |  | 发送类型,call ,sms |
| ip | [string](#string) |  | ip |
| template | [string](#string) |  | 模板 |






<a name="sqlx.SendCodeResponse"></a>

### SendCodeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 @gotags: valid:&#34;ip&#34; custom_tag:&#34;custom_value&#34; |
| data | [SendCodeResponse.DataEntry](#sqlx.SendCodeResponse.DataEntry) | repeated | 数据 |






<a name="sqlx.SendCodeResponse.DataEntry"></a>

### SendCodeResponse.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="sqlx.SendStatus"></a>

### SendStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| needImageCode | [bool](#bool) |  | 需要图形验证码 |
| forceCall | [bool](#bool) |  | 强制语音 |
| isForbidden | [bool](#bool) |  | 被禁止 |
| numberLimit | [bool](#bool) |  | 数量超限制 |






<a name="sqlx.VerifyImageCodeRequest"></a>

### VerifyImageCodeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| ticket | [string](#string) |  | 图形验证码ticket |
| randStr | [string](#string) |  | 图形验证码randStr |
| ip | [string](#string) |  | 图形验证码ip |
| scene | [string](#string) |  | 场景 |






<a name="sqlx.VerifyImageCodeResponse"></a>

### VerifyImageCodeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 |






<a name="sqlx.VerifyRequest"></a>

### VerifyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| code | [string](#string) |  | 验证码 |
| template | [string](#string) |  | 模板 |






<a name="sqlx.VerifyResponse"></a>

### VerifyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [VerifyResponse.DataEntry](#sqlx.VerifyResponse.DataEntry) | repeated | 数据 |






<a name="sqlx.VerifyResponse.DataEntry"></a>

### VerifyResponse.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 


<a name="sqlx.Code"></a>

### Code
验证码

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SendCode | [SendCodeRequest](#sqlx.SendCodeRequest) | [SendCodeResponse](#sqlx.SendCodeResponse) | 发送 |
| Verify | [VerifyRequest](#sqlx.VerifyRequest) | [VerifyResponse](#sqlx.VerifyResponse) | 校验 |
| IsCheckImageCode | [IsCheckImageCodeRequest](#sqlx.IsCheckImageCodeRequest) | [IsCheckImageCodeResponse](#sqlx.IsCheckImageCodeResponse) | 是否校验图片验证码 |
| VerifyImageCode | [VerifyImageCodeRequest](#sqlx.VerifyImageCodeRequest) | [VerifyImageCodeResponse](#sqlx.VerifyImageCodeResponse) | 校验图片验证码 |
| GetSendStatus | [GetSendStatusRequest](#sqlx.GetSendStatusRequest) | [GetSendStatusResponse](#sqlx.GetSendStatusResponse) | 获取发送状态 |

 



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

