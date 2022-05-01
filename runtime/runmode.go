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
	case 5:
		return "release"
	default:
		return ""
	}
}

const (
	RunModeLocal   RunMode = 0
	RunModeDev     RunMode = 1
	RunModeTest    RunMode = 2
	RunModeStag    RunMode = 3
	RunModeProd    RunMode = 4
	RunModeRelease RunMode = 5
)

func IsK8s() bool {
	return Namespace != ""
}

func IsLocal() bool {
	return Mode == RunModeLocal
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
