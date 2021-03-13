package registry

import (
	"strings"
)

func SplitService(service string) (name, id string) {
	var separator string
	if strings.Contains(service, ".") {
		separator = "."
	}

	if strings.Contains(service, "-") {
		separator = "-"
	}

	services := strings.Split(service, separator)
	switch len(services) {
	case 1:
		return services[0], ""
	case 2:
		return services[0], services[1]
	case 3:
		return services[0], services[2]
	default:
		return "", ""
	}
}
