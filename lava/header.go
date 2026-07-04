package lava

// RequestHeader is the transport-agnostic HTTP request header surface used by lava.
type RequestHeader interface {
	Peek(key string) []byte
	PeekAll(key string) [][]byte
	Set(key, value string)
	Add(key, value string)
	Method() []byte
	SetMethod(method string)
	SetMethodBytes(method []byte)
	RequestURI() []byte
	SetRequestURI(uri string)
	SetRequestURIBytes(uri []byte)
	ContentType() []byte
	SetContentType(ct string)
	SetContentTypeBytes(ct []byte)
	Referer() []byte
	UserAgent() []byte
	SetCookieBytesKV(key, value []byte)
	VisitAll(fn func(key, value []byte))
	String() string
}

// ResponseHeader is the transport-agnostic HTTP response header surface used by lava.
type ResponseHeader interface {
	Peek(key string) []byte
	PeekAll(key string) [][]byte
	Set(key, value string)
	Add(key, value string)
	VisitAll(fn func(key, value []byte))
	String() string
}
