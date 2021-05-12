package gocache

type Options struct {
	onEvicted func(k, v []byte)
}

const Name = "gocache"

type Option func(o *Options)
