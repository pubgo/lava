// Code generated by protoc-gen-resty. DO NOT EDIT.
// versions:
// - protoc-gen-resty v0.1.0
// - protoc           v3.17.3
// source: proto/hello/helloworld.proto

package hello

import (
	context "context"
	v2 "github.com/go-resty/resty/v2"
	go_json "github.com/goccy/go-json"
	reflect "reflect"
)

type GreeterResty interface {
	SayHello(ctx context.Context, in *HelloRequest, opts ...func(req *v2.Request)) (*HelloReply, error)
}

func NewGreeterResty(client *v2.Client) GreeterResty {
	client.SetContentLength(true)
	return &greeterResty{client: client}
}

type greeterResty struct {
	client *v2.Client
}

func (c *greeterResty) SayHello(ctx context.Context, in *HelloRequest, opts ...func(req *v2.Request)) (*HelloReply, error) {
	var req = c.client.R()
	if ctx != nil {
		req.SetContext(ctx)
	}
	for i := range opts {
		opts[i](req)
	}
	if in != nil {
		var rv = reflect.ValueOf(in).Elem()
		var rt = reflect.TypeOf(in).Elem()
		for i := 0; i < rt.NumField(); i++ {
			if val, ok := rt.Field(i).Tag.Lookup("param"); ok && val != "" {
				req.SetPathParam(val, rv.Field(i).String())
				continue
			}
			if val, ok := rt.Field(i).Tag.Lookup("query"); ok && val != "" {
				req.SetQueryParam(val, rv.Field(i).String())
				continue
			}
			if val, ok := rt.Field(i).Tag.Lookup("json"); ok && val != "" {
				req.SetQueryParam(val, rv.Field(i).String())
			}
		}
	}
	var resp, err = req.Execute("GET", "/say/{name}")
	if err != nil {
		return nil, err
	}
	out := new(HelloReply)
	if err := go_json.Unmarshal(resp.Body(), out); err != nil {
		return nil, err
	}
	return out, nil
}