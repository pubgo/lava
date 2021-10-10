package consts

import "time"

const Default = "default"
const DefaultDateFormat = "2006-01-02"
const DefaultTimeFormat = "2006-01-02 15:04:05"
const Unknown = "unknown"
const Driver = "driver"
const DefaultTimeout = time.Second * 2

func GetDefault(names ...string) string {
	var name = Default
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}
	return name
}
