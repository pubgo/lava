package mdns

import "time"

type Cfg struct {
	Text []string `json:"text"`
	TTL  time.Duration `json:"ttl"`
}
