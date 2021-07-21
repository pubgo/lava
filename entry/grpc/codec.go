package grpc

//func init() {
//	encoding.RegisterCodec(&uriCodec{})
//}
//
//// 解析http get请求的query参数
//type uriCodec struct{}
//
//func (c *uriCodec) Name() string { return "uri" }
//func (c *uriCodec) Marshal(v interface{}) ([]byte, error) {
//	return json.Marshal(v)
//}
//
//func (c *uriCodec) Unmarshal(data []byte, v interface{}) error {
//	var u, err = url.ParseQuery(string(data))
//	if err != nil {
//		return err
//	}
//
//	return gutil.MapFormByTag(v, u, "json")
//}
