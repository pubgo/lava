package runmode

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
	Local   RunMode = 0
	Dev     RunMode = 1
	Test    RunMode = 2
	Stag    RunMode = 3
	Prod    RunMode = 4
	Release RunMode = 5
)

func IsK8s() bool {
	return Namespace != ""
}

func IsLocal() bool {
	return Mode == Local
}

func IsDev() bool {
	return Mode == Dev
}

func IsTest() bool {
	return Mode == Test
}

func IsStag() bool {
	return Mode == Stag
}

func IsProd() bool {
	return Mode == Prod
}

func IsRelease() bool {
	return Mode == Release
}
