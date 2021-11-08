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
	RunmodeDev     RunMode = 0
	RunmodeTest    RunMode = 1
	RunmodeStag    RunMode = 2
	RunmodeProd    RunMode = 3
	RunmodeRelease RunMode = 4
	RunModeUnknown RunMode = 5
)

var RunmodeValue = map[string]int32{
	"dev":     0,
	"test":    1,
	"stag":    2,
	"prod":    3,
	"release": 4,
	"unknown": 5,
}

// CheckMode 运行环境检查
func CheckMode() bool {
	if _, ok := RunmodeValue[Mode]; ok {
		return ok
	}

	return false
}

func IsDev() bool {
	return Mode == RunmodeDev.String()
}

func IsTest() bool {
	return Mode == RunmodeTest.String()
}

func IsStag() bool {
	return Mode == RunmodeStag.String()
}

func IsProd() bool {
	return Mode == RunmodeProd.String()
}

func IsRelease() bool {
	return Mode == RunmodeRelease.String()
}
