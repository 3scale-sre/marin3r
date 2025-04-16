//go:build !disable_pgv
// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: envoy/config/grpc_credential/v3/aws_iam.proto

package grpc_credentialv3

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

// Validate checks the field values on AwsIamConfig with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *AwsIamConfig) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on AwsIamConfig with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in AwsIamConfigMultiError, or
// nil if none found.
func (m *AwsIamConfig) ValidateAll() error {
	return m.validate(true)
}

func (m *AwsIamConfig) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if utf8.RuneCountInString(m.GetServiceName()) < 1 {
		err := AwsIamConfigValidationError{
			field:  "ServiceName",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	// no validation rules for Region

	if len(errors) > 0 {
		return AwsIamConfigMultiError(errors)
	}

	return nil
}

// AwsIamConfigMultiError is an error wrapping multiple validation errors
// returned by AwsIamConfig.ValidateAll() if the designated constraints aren't met.
type AwsIamConfigMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m AwsIamConfigMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m AwsIamConfigMultiError) AllErrors() []error { return m }

// AwsIamConfigValidationError is the validation error returned by
// AwsIamConfig.Validate if the designated constraints aren't met.
type AwsIamConfigValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e AwsIamConfigValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e AwsIamConfigValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e AwsIamConfigValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e AwsIamConfigValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e AwsIamConfigValidationError) ErrorName() string { return "AwsIamConfigValidationError" }

// Error satisfies the builtin error interface
func (e AwsIamConfigValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sAwsIamConfig.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = AwsIamConfigValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = AwsIamConfigValidationError{}
