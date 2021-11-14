# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [proto/login/bind.proto](#proto/login/bind.proto)
    - [AutomaticBindRequest](#login.AutomaticBindRequest)
    - [AutomaticBindResponse](#login.AutomaticBindResponse)
    - [BindChangeRequest](#login.BindChangeRequest)
    - [BindChangeResponse](#login.BindChangeResponse)
    - [BindData](#login.BindData)
    - [BindPhoneParseByOneClickRequest](#login.BindPhoneParseByOneClickRequest)
    - [BindPhoneParseByOneClickResponse](#login.BindPhoneParseByOneClickResponse)
    - [BindPhoneParseByOneClickResponse.DataEntry](#login.BindPhoneParseByOneClickResponse.DataEntry)
    - [BindPhoneParseRequest](#login.BindPhoneParseRequest)
    - [BindPhoneParseResponse](#login.BindPhoneParseResponse)
    - [BindPhoneParseResponse.DataEntry](#login.BindPhoneParseResponse.DataEntry)
    - [BindVerifyRequest](#login.BindVerifyRequest)
    - [BindVerifyResponse](#login.BindVerifyResponse)
    - [BindVerifyResponse.DataEntry](#login.BindVerifyResponse.DataEntry)
    - [CheckRequest](#login.CheckRequest)
    - [CheckResponse](#login.CheckResponse)
    - [CheckResponse.DataEntry](#login.CheckResponse.DataEntry)
  
    - [BindTelephone](#login.BindTelephone)
  
- [proto/login/code.proto](#proto/login/code.proto)
    - [GetSendStatusRequest](#login.GetSendStatusRequest)
    - [GetSendStatusResponse](#login.GetSendStatusResponse)
    - [IsCheckImageCodeRequest](#login.IsCheckImageCodeRequest)
    - [IsCheckImageCodeResponse](#login.IsCheckImageCodeResponse)
    - [SendCodeRequest](#login.SendCodeRequest)
    - [SendCodeResponse](#login.SendCodeResponse)
    - [SendCodeResponse.DataEntry](#login.SendCodeResponse.DataEntry)
    - [SendStatus](#login.SendStatus)
    - [VerifyImageCodeRequest](#login.VerifyImageCodeRequest)
    - [VerifyImageCodeResponse](#login.VerifyImageCodeResponse)
    - [VerifyRequest](#login.VerifyRequest)
    - [VerifyResponse](#login.VerifyResponse)
    - [VerifyResponse.DataEntry](#login.VerifyResponse.DataEntry)
  
    - [Code](#login.Code)
  
- [proto/login/login.proto](#proto/login/login.proto)
    - [AuthenticateRequest](#login.AuthenticateRequest)
    - [AuthenticateRequest.CredentialsEntry](#login.AuthenticateRequest.CredentialsEntry)
    - [AuthenticateResponse](#login.AuthenticateResponse)
    - [Credentials](#login.Credentials)
    - [Data](#login.Data)
    - [LoginRequest](#login.LoginRequest)
    - [LoginRequest.DataEntry](#login.LoginRequest.DataEntry)
    - [LoginResponse](#login.LoginResponse)
    - [PlatformInfo](#login.PlatformInfo)
  
    - [Login](#login.Login)
  
- [proto/login/merge.proto](#proto/login/merge.proto)
    - [Reply](#login.Reply)
    - [Reply.DataEntry](#login.Reply.DataEntry)
    - [TelephoneRequest](#login.TelephoneRequest)
    - [WeChatRequest](#login.WeChatRequest)
    - [WeChatUnMergeRequest](#login.WeChatUnMergeRequest)
  
    - [Merge](#login.Merge)
  
- [Scalar Value Types](#scalar-value-types)



<a name="proto/login/bind.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/login/bind.proto



<a name="login.AutomaticBindRequest"></a>

### AutomaticBindRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| uid | [int64](#int64) |  | uid |
| origin | [string](#string) |  | 前缀,通常为空,抖音必须为DY- |






<a name="login.AutomaticBindResponse"></a>

### AutomaticBindResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | 消息 |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [BindData](#login.BindData) |  | 数据 |






<a name="login.BindChangeRequest"></a>

### BindChangeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| uid | [int64](#int64) |  | uid |
| code | [string](#string) |  | 验证码 |
| origin | [string](#string) |  | 前缀,通常为空,抖音必须为DY- |






<a name="login.BindChangeResponse"></a>

### BindChangeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [BindData](#login.BindData) |  | 数据 |






<a name="login.BindData"></a>

### BindData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bindId | [int64](#int64) |  | uid |






<a name="login.BindPhoneParseByOneClickRequest"></a>

### BindPhoneParseByOneClickRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [string](#string) |  | 用于解析手机号加密数据 |
| platformId | [int64](#int64) |  | platformId |
| telephone | [string](#string) |  | telephone 有手机号即验证手机号 |






<a name="login.BindPhoneParseByOneClickResponse"></a>

### BindPhoneParseByOneClickResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | 消息 |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [BindPhoneParseByOneClickResponse.DataEntry](#login.BindPhoneParseByOneClickResponse.DataEntry) | repeated | 数据 |






<a name="login.BindPhoneParseByOneClickResponse.DataEntry"></a>

### BindPhoneParseByOneClickResponse.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="login.BindPhoneParseRequest"></a>

### BindPhoneParseRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [string](#string) |  | 用于解析手机号加密数据 |
| encryptedData | [string](#string) |  | 用于解析手机号加密数据 |
| iv | [string](#string) |  | 用于解析手机号加密数据 |
| platformId | [int64](#int64) |  | platformId |
| uid | [int64](#int64) |  | uid，有uid的情况下不使用code |






<a name="login.BindPhoneParseResponse"></a>

### BindPhoneParseResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | 消息 |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [BindPhoneParseResponse.DataEntry](#login.BindPhoneParseResponse.DataEntry) | repeated | 数据 |






<a name="login.BindPhoneParseResponse.DataEntry"></a>

### BindPhoneParseResponse.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="login.BindVerifyRequest"></a>

### BindVerifyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| uid | [int64](#int64) |  | uid |
| code | [string](#string) |  | 验证码 |
| origin | [string](#string) |  | 前缀,通常为空,抖音必须为DY- |






<a name="login.BindVerifyResponse"></a>

### BindVerifyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | 消息 |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [BindVerifyResponse.DataEntry](#login.BindVerifyResponse.DataEntry) | repeated | 数据 |






<a name="login.BindVerifyResponse.DataEntry"></a>

### BindVerifyResponse.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="login.CheckRequest"></a>

### CheckRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| uid | [int64](#int64) |  | uid |
| origin | [string](#string) |  | 前缀,通常为空,抖音必须为DY- |






<a name="login.CheckResponse"></a>

### CheckResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code,不为0为错误 |
| msg | [string](#string) |  | 错误信息 |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [CheckResponse.DataEntry](#login.CheckResponse.DataEntry) | repeated | 数据 |






<a name="login.CheckResponse.DataEntry"></a>

### CheckResponse.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 


<a name="login.BindTelephone"></a>

### BindTelephone
绑定手机号

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Check | [CheckRequest](#login.CheckRequest) | [CheckResponse](#login.CheckResponse) | 检查是否可以绑定 |
| BindVerify | [BindVerifyRequest](#login.BindVerifyRequest) | [BindVerifyResponse](#login.BindVerifyResponse) | 通过验证码,校验手机号是否可以接收验证码 |
| BindChange | [BindChangeRequest](#login.BindChangeRequest) | [BindChangeResponse](#login.BindChangeResponse) | 通过验证码,进行手机号绑定,换绑 |
| AutomaticBind | [AutomaticBindRequest](#login.AutomaticBindRequest) | [AutomaticBindResponse](#login.AutomaticBindResponse) | 手机号绑定,不通过验证码 |
| BindPhoneParse | [BindPhoneParseRequest](#login.BindPhoneParseRequest) | [BindPhoneParseResponse](#login.BindPhoneParseResponse) | 绑定手机号解析，通过第三方小程序code换取手机号 |
| BindPhoneParseByOneClick | [BindPhoneParseByOneClickRequest](#login.BindPhoneParseByOneClickRequest) | [BindPhoneParseByOneClickResponse](#login.BindPhoneParseByOneClickResponse) | 绑定手机号解析，通过阿里一键 |

 



<a name="proto/login/code.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/login/code.proto



<a name="login.GetSendStatusRequest"></a>

### GetSendStatusRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| sendType | [string](#string) |  | 发送类型 |
| template | [string](#string) |  | 模板 |
| signR | [int64](#int64) |  | 是否越狱标示 |
| ip | [string](#string) |  | ip |






<a name="login.GetSendStatusResponse"></a>

### GetSendStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [SendStatus](#login.SendStatus) |  | 数据 |






<a name="login.IsCheckImageCodeRequest"></a>

### IsCheckImageCodeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| scene | [string](#string) |  | 场景 |






<a name="login.IsCheckImageCodeResponse"></a>

### IsCheckImageCodeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [bool](#bool) |  | 数据 |






<a name="login.SendCodeRequest"></a>

### SendCodeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 电话 |
| sendType | [string](#string) |  | 发送类型,call ,sms |
| ip | [string](#string) |  | ip |
| template | [string](#string) |  | 模板 |






<a name="login.SendCodeResponse"></a>

### SendCodeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [SendCodeResponse.DataEntry](#login.SendCodeResponse.DataEntry) | repeated | 数据 |






<a name="login.SendCodeResponse.DataEntry"></a>

### SendCodeResponse.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="login.SendStatus"></a>

### SendStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| needImageCode | [bool](#bool) |  | 需要图形验证码 |
| forceCall | [bool](#bool) |  | 强制语音 |
| isForbidden | [bool](#bool) |  | 被禁止 |
| numberLimit | [bool](#bool) |  | 数量超限制 |






<a name="login.VerifyImageCodeRequest"></a>

### VerifyImageCodeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| ticket | [string](#string) |  | 图形验证码ticket |
| randStr | [string](#string) |  | 图形验证码randStr |
| ip | [string](#string) |  | 图形验证码ip |
| scene | [string](#string) |  | 场景 |






<a name="login.VerifyImageCodeResponse"></a>

### VerifyImageCodeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 |






<a name="login.VerifyRequest"></a>

### VerifyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nationCode | [string](#string) |  | 区号 |
| telephone | [string](#string) |  | 手机号 |
| code | [string](#string) |  | 验证码 |
| template | [string](#string) |  | 模板 |






<a name="login.VerifyResponse"></a>

### VerifyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [VerifyResponse.DataEntry](#login.VerifyResponse.DataEntry) | repeated | 数据 |






<a name="login.VerifyResponse.DataEntry"></a>

### VerifyResponse.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 


<a name="login.Code"></a>

### Code
验证码

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SendCode | [SendCodeRequest](#login.SendCodeRequest) | [SendCodeResponse](#login.SendCodeResponse) | 发送 |
| Verify | [VerifyRequest](#login.VerifyRequest) | [VerifyResponse](#login.VerifyResponse) | 校验 |
| IsCheckImageCode | [IsCheckImageCodeRequest](#login.IsCheckImageCodeRequest) | [IsCheckImageCodeResponse](#login.IsCheckImageCodeResponse) | 是否校验图片验证码 |
| VerifyImageCode | [VerifyImageCodeRequest](#login.VerifyImageCodeRequest) | [VerifyImageCodeResponse](#login.VerifyImageCodeResponse) | 校验图片验证码 |
| GetSendStatus | [GetSendStatusRequest](#login.GetSendStatusRequest) | [GetSendStatusResponse](#login.GetSendStatusResponse) | 获取发送状态 |

 



<a name="proto/login/login.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/login/login.proto



<a name="login.AuthenticateRequest"></a>

### AuthenticateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| credentials | [AuthenticateRequest.CredentialsEntry](#login.AuthenticateRequest.CredentialsEntry) | repeated | 凭证,cookie:string or token:sting |






<a name="login.AuthenticateRequest.CredentialsEntry"></a>

### AuthenticateRequest.CredentialsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="login.AuthenticateResponse"></a>

### AuthenticateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | 错误码,0 为正常 |
| msg | [string](#string) |  | 错误信息 |
| nowTime | [int64](#int64) |  | 请求响应时间戳 |
| data | [Data](#login.Data) |  | 数据 |






<a name="login.Credentials"></a>

### Credentials



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uid | [int64](#int64) |  | userinfoId 对应 bindId |
| uri | [string](#string) |  | uri |
| openid | [string](#string) |  | openid |
| isNew | [bool](#bool) |  | isNew |
| isFirstRegister | [bool](#bool) |  | 是否首次注册 |
| isBindTelephone | [bool](#bool) |  | 是否绑定手机号 |
| platformInfo | [PlatformInfo](#login.PlatformInfo) |  | platformId |






<a name="login.Data"></a>

### Data



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uid | [int64](#int64) |  | userinfoId |
| uri | [string](#string) |  | uri |
| nickname | [string](#string) |  | 个人昵称,没有店铺昵称覆盖逻辑 |
| headImgUrl | [string](#string) |  | 个人头像,没有店铺头像覆盖逻辑 |
| signature | [string](#string) |  | 签名 |
| sex | [int64](#int64) |  | 性别, 性别 0未知,1男,2女 |
| region | [string](#string) |  | 区域 |
| country | [string](#string) |  | 国家 |
| province | [string](#string) |  | 省市 |
| city | [string](#string) |  | 城市 |
| lang | [string](#string) |  | 语言类型,默认 &#34;&#34; |
| createTime | [int64](#int64) |  | 注册时间戳 |
| modifyTime | [int64](#int64) |  | 更新时间戳 |
| currentlyLoggedPlatformId | [int64](#int64) |  | 当前登录平台id ,对应 center 表 type 字段 |






<a name="login.LoginRequest"></a>

### LoginRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| platformId | [int64](#int64) |  | 平台id ,对应 center 表 type 字段 |
| data | [LoginRequest.DataEntry](#login.LoginRequest.DataEntry) | repeated | 登录相关信息,json,手机号登录参数 UserType int64 `json:&#34;userType&#34;` 	VerifyType string `json:&#34;verifyType&#34;` 	NationCode string `json:&#34;nationCode&#34;` 	Telephone string `json:&#34;telephone&#34;` 	Code string `json:&#34;code&#34;` 	LoginToken string `json:&#34;loginToken&#34;` 	DeviceId string `json:&#34;deviceId&#34;` 	SysMessageNum int64 `json:&#34;sysMessageNum&#34;` |
| scope | [string](#string) |  | 凭据类型,普通用户 base, 特权?超级? super |






<a name="login.LoginRequest.DataEntry"></a>

### LoginRequest.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="login.LoginResponse"></a>

### LoginResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | 错误码,0 为正常 |
| msg | [string](#string) |  | 错误信息 |
| nowTime | [int64](#int64) |  | 请求响应时间戳 |
| data | [Credentials](#login.Credentials) |  | 数据 |






<a name="login.PlatformInfo"></a>

### PlatformInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| platformId | [int64](#int64) |  | platformId |
| originalUid | [int64](#int64) |  | originalId 原始ID,platformId 对应的user |
| originalUri | [string](#string) |  | originalUri 原始uri,platformId 对应的user |
| originalOpenid | [string](#string) |  | originalOpenid 原始openid,platformId 对应的user |





 

 

 


<a name="login.Login"></a>

### Login
统一登录入口

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Login | [LoginRequest](#login.LoginRequest) | [LoginResponse](#login.LoginResponse) | 登录注册获取凭证,cookie,token |
| Authenticate | [AuthenticateRequest](#login.AuthenticateRequest) | [AuthenticateResponse](#login.AuthenticateResponse) | 使用凭证获取用户信息 |

 



<a name="proto/login/merge.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proto/login/merge.proto



<a name="login.Reply"></a>

### Reply



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [int64](#int64) |  | code |
| msg | [string](#string) |  | msg |
| nowTime | [int64](#int64) |  | 时间戳 |
| data | [Reply.DataEntry](#login.Reply.DataEntry) | repeated | 数据 |






<a name="login.Reply.DataEntry"></a>

### Reply.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="login.TelephoneRequest"></a>

### TelephoneRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uid | [int64](#int64) |  | 登陆用户 |
| targetTelephone | [string](#string) |  | 新手机号 |
| isNewProcess | [bool](#bool) |  | 是否走新流程 |






<a name="login.WeChatRequest"></a>

### WeChatRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uid | [int64](#int64) |  | 登陆用户 |
| targetUid | [int64](#int64) |  | 要合并的用户 |






<a name="login.WeChatUnMergeRequest"></a>

### WeChatUnMergeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uid | [int64](#int64) |  | 登陆用户 |





 

 

 


<a name="login.Merge"></a>

### Merge
账户合并

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Telephone | [TelephoneRequest](#login.TelephoneRequest) | [Reply](#login.Reply) | 手机号合并,换绑,手机号更换 |
| TelephoneCheck | [TelephoneRequest](#login.TelephoneRequest) | [Reply](#login.Reply) | 手机号账户合并检查 |
| WeChat | [WeChatRequest](#login.WeChatRequest) | [Reply](#login.Reply) | 微信账户绑定 |
| WeChatCheck | [WeChatRequest](#login.WeChatRequest) | [Reply](#login.Reply) | 微信合并检查 |
| WeChatUnMerge | [WeChatUnMergeRequest](#login.WeChatUnMergeRequest) | [Reply](#login.Reply) | 解除微信绑定, 必须拥有手机号 |

 



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

