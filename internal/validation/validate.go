package validation

import (
	"strings"
)

type FieldError struct {
	Field string
	Reason string
}

type Errors []FieldError

func (e Errors) ToInvalidParams() []InvalidParam {
	var out []InvalidParam
	for _, fe := range e {
		out = append(out, InvalidParam{Name: fe.Field, Reason: fe.Reason})
	}
	return out
}

func NotBlank(field, value string) *FieldError {
	if strings.TrimSpace(value) == "" { return &FieldError{Field: field, Reason: "must not be blank"} }
	return nil
}

func MinLen(field, value string, n int) *FieldError {
	if len(value) < n { return &FieldError{Field: field, Reason: "length must be >="} }
	return nil
}
