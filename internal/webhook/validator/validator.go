package validator

type Validator interface {
	Source() string
	Validate(payload []byte, headers map[string]string) error
}
