//go:build !disable_pgv
// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: envoy/extensions/filters/http/dynamic_modules/v3/dynamic_modules.proto

package dynamic_modulesv3

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on DynamicModuleFilter with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *DynamicModuleFilter) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DynamicModuleFilter with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// DynamicModuleFilterMultiError, or nil if none found.
func (m *DynamicModuleFilter) ValidateAll() error {
	return m.validate(true)
}

func (m *DynamicModuleFilter) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetDynamicModuleConfig()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, DynamicModuleFilterValidationError{
					field:  "DynamicModuleConfig",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, DynamicModuleFilterValidationError{
					field:  "DynamicModuleConfig",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetDynamicModuleConfig()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return DynamicModuleFilterValidationError{
				field:  "DynamicModuleConfig",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for FilterName

	// no validation rules for FilterConfig

	if len(errors) > 0 {
		return DynamicModuleFilterMultiError(errors)
	}

	return nil
}

// DynamicModuleFilterMultiError is an error wrapping multiple validation
// errors returned by DynamicModuleFilter.ValidateAll() if the designated
// constraints aren't met.
type DynamicModuleFilterMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DynamicModuleFilterMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DynamicModuleFilterMultiError) AllErrors() []error { return m }

// DynamicModuleFilterValidationError is the validation error returned by
// DynamicModuleFilter.Validate if the designated constraints aren't met.
type DynamicModuleFilterValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DynamicModuleFilterValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DynamicModuleFilterValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DynamicModuleFilterValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DynamicModuleFilterValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DynamicModuleFilterValidationError) ErrorName() string {
	return "DynamicModuleFilterValidationError"
}

// Error satisfies the builtin error interface
func (e DynamicModuleFilterValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDynamicModuleFilter.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DynamicModuleFilterValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DynamicModuleFilterValidationError{}
