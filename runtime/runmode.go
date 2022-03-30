package runtime

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
	RunModeDev     RunMode = 0
	RunModeTest    RunMode = 1
	RunModeStag    RunMode = 2
	RunModeProd    RunMode = 3
	RunModeRelease RunMode = 4
	RunModeUnknown RunMode = 5
)

var RunModeValue = map[string]int32{
	"dev":     0,
	"test":    1,
	"stag":    2,
	"prod":    3,
	"release": 4,
	"unknown": 5,
}

func IsDev() bool {
	return Mode == RunModeDev
}

func IsTest() bool {
	return Mode == RunModeTest
}

func IsStag() bool {
	return Mode == RunModeStag
}

func IsProd() bool {
	return Mode == RunModeProd
}

func IsRelease() bool {
	return Mode == RunModeRelease
}
