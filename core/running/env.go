package running

// Well-known runtime environment names (see EnvFlag / --runenv).
const (
	EnvDev   = "dev"
	EnvTest  = "test"
	EnvStage = "stage" // staging
	EnvProd  = "prod"
)

// EnvNames lists all supported environment values in promotion order.
var EnvNames = []string{EnvDev, EnvTest, EnvStage, EnvProd}

// EnvName returns the current runtime environment.
func EnvName() string {
	return Env.String()
}

// IsDev reports whether the process runs in the development environment.
func IsDev() bool {
	return EnvName() == EnvDev
}

// IsTest reports whether the process runs in the test environment.
func IsTest() bool {
	return EnvName() == EnvTest
}

// IsStage reports whether the process runs in the staging environment.
func IsStage() bool {
	return EnvName() == EnvStage
}

// IsProd reports whether the process runs in the production environment.
func IsProd() bool {
	return EnvName() == EnvProd
}

// IsNonProd reports whether the process runs outside production.
func IsNonProd() bool {
	switch EnvName() {
	case EnvDev, EnvTest, EnvStage:
		return true
	default:
		return false
	}
}
