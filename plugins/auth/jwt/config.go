package jwt

import (
	"net/http"
	"time"
)

type Cfg struct {
	ExpireAfter string
	SecretFunc  SecretFunc //required
	Expire      time.Duration
	Realm       string //required

	//optional. Authorize check whether this request could access some resource or API based on json claims.
	//Typically, this method should communicate with a RBAC, ABAC system
	Authorize func(payload map[string]interface{}, req *http.Request) error

	//optional.
	// this function control whether a request should be validate or not
	// if this func is nil, validate all requests.
	MustAuth func(req *http.Request) bool
}

type SigningMethod int

//const
const (
	RS256 SigningMethod = 1
	RS512 SigningMethod = 2
	HS256 SigningMethod = 3
)

//Options is options
type Options struct {
	Expire        string
	SigningMethod SigningMethod
}

//Option is option
type Option func(options *Options)

//WithExpTime generate a token which expire after a duration
//for example 5s,1m,24h
func WithExpTime(exp string) Option {
	return func(options *Options) {
		options.Expire = exp
	}
}

//WithSigningMethod specify the sign method
func WithSigningMethod(m SigningMethod) Option {
	return func(options *Options) {
		options.SigningMethod = m
	}
}

////Use put a custom auth logic
////then register handler to chassis
//func Use(middleware *Auth) {
//	auth = middleware
//	if auth.Expire == 0 {
//		openlog.Warn("token issued by service will not expire")
//	}
//	if auth.MustAuth == nil {
//		openlog.Info("auth all requests")
//	} else {
//		openlog.Warn("under some condition, no auth")
//	}
//}
//
////SetExpire reset the expire time
//func SetExpire(duration time.Duration) {
//	auth.Expire = duration
//}
