package resty

import (
	"bytes"
	"context"
	"encoding/gob"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/pubgo/funk/log"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/lava"
)

func NewJar(log log.Logger) *Jar {
	return &Jar{
		log:     log.WithName("cookie-jar"),
		cookies: make(map[string]*fasthttp.Cookie),
	}
}

var _ lava.Middleware = (*Jar)(nil)

type Jar struct {
	mu  sync.Mutex
	log log.Logger

	cookies map[string]*fasthttp.Cookie
}

func (j *Jar) Middleware(next lava.HandlerFunc) lava.HandlerFunc {
	return func(ctx context.Context, req lava.Request) (lava.Response, error) {
		for _, c := range j.cookies {
			req.Header().SetCookieBytesKV(c.Key(), c.Value())
		}

		rsp, err := next(ctx, req)
		if err != nil {
			return nil, err
		}

		rsp.Header().VisitAllCookie(func(key, value []byte) {
			acquireCookie := fasthttp.AcquireCookie()
			if err := acquireCookie.ParseBytes(value); err != nil {
				j.log.Err(err, ctx).Msg("failed to parse cookie")
			} else {
				j.cookies[string(acquireCookie.Key())] = acquireCookie
			}
		})

		return rsp, err
	}
}

func (j *Jar) String() string {
	return "cookie-jar"
}

func (j *Jar) PeekValue(key string) []byte {
	c, ok := j.cookies[key]
	if ok {
		return c.Value()
	}

	return nil
}

func (j *Jar) Peek(key string) *fasthttp.Cookie {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.cookies[key]
}

func (j *Jar) ReleaseCookie(key string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	c, ok := j.cookies[key]
	if ok {
		fasthttp.ReleaseCookie(c)
		delete(j.cookies, key)
	}
}

func (j *Jar) MarshalJSON() ([]byte, error) {
	cookies, err := j.makeEncodable()
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(cookies)

	return data, err
}

func (j *Jar) UnmarshalJSON(data []byte) error {
	cooks := cookies{}

	err := json.Unmarshal(data, &cooks)
	if err != nil {
		return err
	}

	return err
}

func (j *Jar) EncodeGOB() ([]byte, error) {
	cookies, err := j.makeEncodable()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	err = gob.NewEncoder(&buf).Encode(cookies)

	return buf.Bytes(), err
}

func (j *Jar) DecodeGOB(data []byte) error {
	cooks := &cookies{}

	var buf bytes.Buffer
	buf.Write(data)

	err := gob.NewDecoder(&buf).Decode(cooks)
	if err != nil {
		return err
	}

	err = j.decode(*cooks)

	return err
}

func (j *Jar) decode(cooks cookies) error {
	for _, v := range cooks {
		expire := new(time.Time)
		err := expire.UnmarshalText([]byte(v.Expire))
		if err != nil {
			return err
		}

		c := fasthttp.AcquireCookie()

		c.SetKey(v.Key)
		c.SetValue(v.Value)
		c.SetExpire(*expire)
		c.SetMaxAge(v.MaxAge)
		c.SetDomain(v.Domain)
		c.SetPath(v.Path)
		c.SetHTTPOnly(v.HTTPOnly)
		c.SetSecure(v.Secure)
		c.SetSameSite(v.SameSite)

		j.cookies[v.Key] = c
	}
	return nil
}

func (j *Jar) makeEncodable() (cookies, error) {
	cookies := cookies{}

	for _, v := range j.cookies {
		expire, err := v.Expire().MarshalText()
		if err != nil {
			return nil, err
		}

		c := cookie{
			Key:      string(v.Key()),
			Value:    string(v.Value()),
			Expire:   string(expire),
			MaxAge:   v.MaxAge(),
			Domain:   string(v.Domain()),
			Path:     string(v.Path()),
			HTTPOnly: v.HTTPOnly(),
			Secure:   v.Secure(),
			SameSite: v.SameSite(),
		}

		cookies = append(cookies, c)
	}

	return cookies, nil
}

//easyjson:json
type cookies []cookie

type cookie struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Expire string `json:"time"`

	MaxAge int    `json:"max_age"`
	Domain string `json:"domain"`
	Path   string `json:"path"`

	HTTPOnly bool                    `json:"http_only"`
	Secure   bool                    `json:"secure"`
	SameSite fasthttp.CookieSameSite `json:"same_site"`
}
