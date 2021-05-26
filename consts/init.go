package consts

const Default = "default"
const Unknown = "unknown"
const Driver = "driver"

var NotFound = new(struct{})

func GetDefault(names ...string) string {
	var name = Default
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}
	return name
}
