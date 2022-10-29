package validate

type validator interface {
	Validate() error
}
