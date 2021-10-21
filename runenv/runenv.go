package runenv

type RunMode int32

func (x RunMode) String() string {
	switch x {
	case 0:
		return "dev"
	case 1:
		return "test"
	case 2:
		return "stag"
	case 3:
		return "prod"
	case 4:
		return "release"
	default:
		return "unknown"
	}
}

const (
	RunMode_dev     RunMode = 0
	RunMode_test    RunMode = 1
	RunMode_stag    RunMode = 2
	RunMode_prod    RunMode = 3
	RunMode_release RunMode = 4
	RunMode_unknown RunMode = 5
)

var RunMode_value = map[string]int32{
	"dev":     0,
	"test":    1,
	"stag":    2,
	"prod":    3,
	"release": 4,
	"unknown": 5,
}

// CheckMode 运行环境检查
func CheckMode() bool {
	if _, ok := RunMode_value[Mode]; ok {
		return ok
	}

	return false
}

func IsDev() bool {
	return Mode == RunMode_dev.String()
}

func IsTest() bool {
	return Mode == RunMode_test.String()
}

func IsStag() bool {
	return Mode == RunMode_stag.String()
}

func IsProd() bool {
	return Mode == RunMode_prod.String()
}

func IsRelease() bool {
	return Mode == RunMode_release.String()
}
