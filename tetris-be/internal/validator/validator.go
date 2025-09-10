package validator

import "regexp"

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string // Errors[key]: message
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func In(value string, list ...string) bool {
	for _, element := range list {
		if value == element {
			return true
		}
	}
	return false
}

func Match(value string, pattern *regexp.Regexp) bool {
	return pattern.MatchString(value)
}

func Unique(values []string) bool {
	set := make(map[string]struct{})

	for _, val := range values {
		if _, exists := set[val]; exists {
			return false
		}
		set[val] = struct{}{}
	}
	return true
}
