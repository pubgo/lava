package consts

const Default = "default"
const DefaultTimeFormat = "2006-01-02 15:04:05"
const Unknown = "unknown"
const Driver = "driver"

func GetDefault(names ...string) string {
	var name = Default
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}
	return name
}
