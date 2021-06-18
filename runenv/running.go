package runenv

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
