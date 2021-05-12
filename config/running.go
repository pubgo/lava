package config

func parseRunMode(mode string) RunMode {
	if val, ok := RunMode_value[mode]; ok {
		return RunMode(val)
	}

	return RunMode_unknown
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
