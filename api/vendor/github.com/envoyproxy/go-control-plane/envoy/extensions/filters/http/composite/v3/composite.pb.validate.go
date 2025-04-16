//go:build !disable_pgv
// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: envoy/extensions/filters/http/composite/v3/composite.proto

package compositev3

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

// Validate checks the field values on Composite with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Composite) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Composite with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in CompositeMultiError, or nil
// if none found.
func (m *Composite) ValidateAll() error {
	return m.validate(true)
}

func (m *Composite) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return CompositeMultiError(errors)
	}

	return nil
}

// CompositeMultiError is an error wrapping multiple validation errors returned
// by Composite.ValidateAll() if the designated constraints aren't met.
type CompositeMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CompositeMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CompositeMultiError) AllErrors() []error { return m }

// CompositeValidationError is the validation error returned by
// Composite.Validate if the designated constraints aren't met.
type CompositeValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CompositeValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CompositeValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CompositeValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CompositeValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CompositeValidationError) ErrorName() string { return "CompositeValidationError" }

// Error satisfies the builtin error interface
func (e CompositeValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sComposite.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CompositeValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CompositeValidationError{}

// Validate checks the field values on DynamicConfig with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *DynamicConfig) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DynamicConfig with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in DynamicConfigMultiError, or
// nil if none found.
func (m *DynamicConfig) ValidateAll() error {
	return m.validate(true)
}

func (m *DynamicConfig) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if utf8.RuneCountInString(m.GetName()) < 1 {
		err := DynamicConfigValidationError{
			field:  "Name",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if all {
		switch v := interface{}(m.GetConfigDiscovery()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, DynamicConfigValidationError{
					field:  "ConfigDiscovery",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, DynamicConfigValidationError{
					field:  "ConfigDiscovery",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetConfigDiscovery()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return DynamicConfigValidationError{
				field:  "ConfigDiscovery",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return DynamicConfigMultiError(errors)
	}

	return nil
}

// DynamicConfigMultiError is an error wrapping multiple validation errors
// returned by DynamicConfig.ValidateAll() if the designated constraints
// aren't met.
type DynamicConfigMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DynamicConfigMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DynamicConfigMultiError) AllErrors() []error { return m }

// DynamicConfigValidationError is the validation error returned by
// DynamicConfig.Validate if the designated constraints aren't met.
type DynamicConfigValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DynamicConfigValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DynamicConfigValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DynamicConfigValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DynamicConfigValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DynamicConfigValidationError) ErrorName() string { return "DynamicConfigValidationError" }

// Error satisfies the builtin error interface
func (e DynamicConfigValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDynamicConfig.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DynamicConfigValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DynamicConfigValidationError{}

// Validate checks the field values on ExecuteFilterAction with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ExecuteFilterAction) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ExecuteFilterAction with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ExecuteFilterActionMultiError, or nil if none found.
func (m *ExecuteFilterAction) ValidateAll() error {
	return m.validate(true)
}

func (m *ExecuteFilterAction) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetTypedConfig()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ExecuteFilterActionValidationError{
					field:  "TypedConfig",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ExecuteFilterActionValidationError{
					field:  "TypedConfig",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetTypedConfig()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ExecuteFilterActionValidationError{
				field:  "TypedConfig",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if all {
		switch v := interface{}(m.GetDynamicConfig()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ExecuteFilterActionValidationError{
					field:  "DynamicConfig",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ExecuteFilterActionValidationError{
					field:  "DynamicConfig",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetDynamicConfig()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ExecuteFilterActionValidationError{
				field:  "DynamicConfig",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if all {
		switch v := interface{}(m.GetSamplePercent()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ExecuteFilterActionValidationError{
					field:  "SamplePercent",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ExecuteFilterActionValidationError{
					field:  "SamplePercent",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetSamplePercent()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ExecuteFilterActionValidationError{
				field:  "SamplePercent",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return ExecuteFilterActionMultiError(errors)
	}

	return nil
}

// ExecuteFilterActionMultiError is an error wrapping multiple validation
// errors returned by ExecuteFilterAction.ValidateAll() if the designated
// constraints aren't met.
type ExecuteFilterActionMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ExecuteFilterActionMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ExecuteFilterActionMultiError) AllErrors() []error { return m }

// ExecuteFilterActionValidationError is the validation error returned by
// ExecuteFilterAction.Validate if the designated constraints aren't met.
type ExecuteFilterActionValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ExecuteFilterActionValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ExecuteFilterActionValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ExecuteFilterActionValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ExecuteFilterActionValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ExecuteFilterActionValidationError) ErrorName() string {
	return "ExecuteFilterActionValidationError"
}

// Error satisfies the builtin error interface
func (e ExecuteFilterActionValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sExecuteFilterAction.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ExecuteFilterActionValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ExecuteFilterActionValidationError{}
