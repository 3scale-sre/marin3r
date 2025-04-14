//go:build !disable_pgv
// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: envoy/extensions/router/cluster_specifiers/lua/v3/lua.proto

package luav3

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

// Validate checks the field values on LuaConfig with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *LuaConfig) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on LuaConfig with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in LuaConfigMultiError, or nil
// if none found.
func (m *LuaConfig) ValidateAll() error {
	return m.validate(true)
}

func (m *LuaConfig) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetSourceCode() == nil {
		err := LuaConfigValidationError{
			field:  "SourceCode",
			reason: "value is required",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if all {
		switch v := interface{}(m.GetSourceCode()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, LuaConfigValidationError{
					field:  "SourceCode",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, LuaConfigValidationError{
					field:  "SourceCode",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetSourceCode()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return LuaConfigValidationError{
				field:  "SourceCode",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for DefaultCluster

	if len(errors) > 0 {
		return LuaConfigMultiError(errors)
	}

	return nil
}

// LuaConfigMultiError is an error wrapping multiple validation errors returned
// by LuaConfig.ValidateAll() if the designated constraints aren't met.
type LuaConfigMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m LuaConfigMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m LuaConfigMultiError) AllErrors() []error { return m }

// LuaConfigValidationError is the validation error returned by
// LuaConfig.Validate if the designated constraints aren't met.
type LuaConfigValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e LuaConfigValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e LuaConfigValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e LuaConfigValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e LuaConfigValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e LuaConfigValidationError) ErrorName() string { return "LuaConfigValidationError" }

// Error satisfies the builtin error interface
func (e LuaConfigValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sLuaConfig.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = LuaConfigValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = LuaConfigValidationError{}
