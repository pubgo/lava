package mdns

import "time"

const Name = "mdns"

type Cfg struct {
	TTL time.Duration `json:"ttl"`
}
