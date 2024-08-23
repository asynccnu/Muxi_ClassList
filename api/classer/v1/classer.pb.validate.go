// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: api/classer/v1/classer.proto

package v1

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

// Validate checks the field values on GetClassRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *GetClassRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetClassRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetClassRequestMultiError, or nil if none found.
func (m *GetClassRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetClassRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetWeek() <= 0 {
		err := GetClassRequestValidationError{
			field:  "Week",
			reason: "value must be greater than 0",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetStuId()) != 10 {
		err := GetClassRequestValidationError{
			field:  "StuId",
			reason: "value length must be 10 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)

	}

	if utf8.RuneCountInString(m.GetSemester()) != 1 {
		err := GetClassRequestValidationError{
			field:  "Semester",
			reason: "value length must be 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)

	}

	if utf8.RuneCountInString(m.GetYear()) < 1 {
		err := GetClassRequestValidationError{
			field:  "Year",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return GetClassRequestMultiError(errors)
	}

	return nil
}

// GetClassRequestMultiError is an error wrapping multiple validation errors
// returned by GetClassRequest.ValidateAll() if the designated constraints
// aren't met.
type GetClassRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetClassRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetClassRequestMultiError) AllErrors() []error { return m }

// GetClassRequestValidationError is the validation error returned by
// GetClassRequest.Validate if the designated constraints aren't met.
type GetClassRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetClassRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetClassRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetClassRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetClassRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetClassRequestValidationError) ErrorName() string { return "GetClassRequestValidationError" }

// Error satisfies the builtin error interface
func (e GetClassRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetClassRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetClassRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetClassRequestValidationError{}

// Validate checks the field values on GetClassResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *GetClassResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetClassResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetClassResponseMultiError, or nil if none found.
func (m *GetClassResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *GetClassResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetClasses() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, GetClassResponseValidationError{
						field:  fmt.Sprintf("Classes[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, GetClassResponseValidationError{
						field:  fmt.Sprintf("Classes[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GetClassResponseValidationError{
					field:  fmt.Sprintf("Classes[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return GetClassResponseMultiError(errors)
	}

	return nil
}

// GetClassResponseMultiError is an error wrapping multiple validation errors
// returned by GetClassResponse.ValidateAll() if the designated constraints
// aren't met.
type GetClassResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetClassResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetClassResponseMultiError) AllErrors() []error { return m }

// GetClassResponseValidationError is the validation error returned by
// GetClassResponse.Validate if the designated constraints aren't met.
type GetClassResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetClassResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetClassResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetClassResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetClassResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetClassResponseValidationError) ErrorName() string { return "GetClassResponseValidationError" }

// Error satisfies the builtin error interface
func (e GetClassResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetClassResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetClassResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetClassResponseValidationError{}

// Validate checks the field values on AddClassRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *AddClassRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on AddClassRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// AddClassRequestMultiError, or nil if none found.
func (m *AddClassRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *AddClassRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if utf8.RuneCountInString(m.GetStuId()) != 10 {
		err := AddClassRequestValidationError{
			field:  "StuId",
			reason: "value length must be 10 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)

	}

	if utf8.RuneCountInString(m.GetName()) < 1 {
		err := AddClassRequestValidationError{
			field:  "Name",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetDurClass()) < 1 {
		err := AddClassRequestValidationError{
			field:  "DurClass",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetWhere()) < 1 {
		err := AddClassRequestValidationError{
			field:  "Where",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetTeacher()) < 1 {
		err := AddClassRequestValidationError{
			field:  "Teacher",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if m.GetWeeks() <= 0 {
		err := AddClassRequestValidationError{
			field:  "Weeks",
			reason: "value must be greater than 0",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetSemester()) != 1 {
		err := AddClassRequestValidationError{
			field:  "Semester",
			reason: "value length must be 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)

	}

	if utf8.RuneCountInString(m.GetYear()) < 1 {
		err := AddClassRequestValidationError{
			field:  "Year",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if val := m.GetDay(); val < 1 || val > 7 {
		err := AddClassRequestValidationError{
			field:  "Day",
			reason: "value must be inside range [1, 7]",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	// no validation rules for Credit

	if len(errors) > 0 {
		return AddClassRequestMultiError(errors)
	}

	return nil
}

// AddClassRequestMultiError is an error wrapping multiple validation errors
// returned by AddClassRequest.ValidateAll() if the designated constraints
// aren't met.
type AddClassRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m AddClassRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m AddClassRequestMultiError) AllErrors() []error { return m }

// AddClassRequestValidationError is the validation error returned by
// AddClassRequest.Validate if the designated constraints aren't met.
type AddClassRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e AddClassRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e AddClassRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e AddClassRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e AddClassRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e AddClassRequestValidationError) ErrorName() string { return "AddClassRequestValidationError" }

// Error satisfies the builtin error interface
func (e AddClassRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sAddClassRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = AddClassRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = AddClassRequestValidationError{}

// Validate checks the field values on AddClassResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *AddClassResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on AddClassResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// AddClassResponseMultiError, or nil if none found.
func (m *AddClassResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *AddClassResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Id

	// no validation rules for Msg

	if len(errors) > 0 {
		return AddClassResponseMultiError(errors)
	}

	return nil
}

// AddClassResponseMultiError is an error wrapping multiple validation errors
// returned by AddClassResponse.ValidateAll() if the designated constraints
// aren't met.
type AddClassResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m AddClassResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m AddClassResponseMultiError) AllErrors() []error { return m }

// AddClassResponseValidationError is the validation error returned by
// AddClassResponse.Validate if the designated constraints aren't met.
type AddClassResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e AddClassResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e AddClassResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e AddClassResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e AddClassResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e AddClassResponseValidationError) ErrorName() string { return "AddClassResponseValidationError" }

// Error satisfies the builtin error interface
func (e AddClassResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sAddClassResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = AddClassResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = AddClassResponseValidationError{}

// Validate checks the field values on DeleteClassRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *DeleteClassRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeleteClassRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// DeleteClassRequestMultiError, or nil if none found.
func (m *DeleteClassRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *DeleteClassRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if utf8.RuneCountInString(m.GetId()) < 1 {
		err := DeleteClassRequestValidationError{
			field:  "Id",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetStuId()) != 10 {
		err := DeleteClassRequestValidationError{
			field:  "StuId",
			reason: "value length must be 10 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)

	}

	if utf8.RuneCountInString(m.GetYear()) < 1 {
		err := DeleteClassRequestValidationError{
			field:  "Year",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetSemester()) != 1 {
		err := DeleteClassRequestValidationError{
			field:  "Semester",
			reason: "value length must be 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)

	}

	if len(errors) > 0 {
		return DeleteClassRequestMultiError(errors)
	}

	return nil
}

// DeleteClassRequestMultiError is an error wrapping multiple validation errors
// returned by DeleteClassRequest.ValidateAll() if the designated constraints
// aren't met.
type DeleteClassRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeleteClassRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeleteClassRequestMultiError) AllErrors() []error { return m }

// DeleteClassRequestValidationError is the validation error returned by
// DeleteClassRequest.Validate if the designated constraints aren't met.
type DeleteClassRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteClassRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteClassRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteClassRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteClassRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteClassRequestValidationError) ErrorName() string {
	return "DeleteClassRequestValidationError"
}

// Error satisfies the builtin error interface
func (e DeleteClassRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteClassRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteClassRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteClassRequestValidationError{}

// Validate checks the field values on DeleteClassResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *DeleteClassResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeleteClassResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// DeleteClassResponseMultiError, or nil if none found.
func (m *DeleteClassResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *DeleteClassResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Msg

	if len(errors) > 0 {
		return DeleteClassResponseMultiError(errors)
	}

	return nil
}

// DeleteClassResponseMultiError is an error wrapping multiple validation
// errors returned by DeleteClassResponse.ValidateAll() if the designated
// constraints aren't met.
type DeleteClassResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeleteClassResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeleteClassResponseMultiError) AllErrors() []error { return m }

// DeleteClassResponseValidationError is the validation error returned by
// DeleteClassResponse.Validate if the designated constraints aren't met.
type DeleteClassResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteClassResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteClassResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteClassResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteClassResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteClassResponseValidationError) ErrorName() string {
	return "DeleteClassResponseValidationError"
}

// Error satisfies the builtin error interface
func (e DeleteClassResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteClassResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteClassResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteClassResponseValidationError{}

// Validate checks the field values on UpdateClassRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *UpdateClassRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on UpdateClassRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// UpdateClassRequestMultiError, or nil if none found.
func (m *UpdateClassRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *UpdateClassRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if utf8.RuneCountInString(m.GetStuId()) != 10 {
		err := UpdateClassRequestValidationError{
			field:  "StuId",
			reason: "value length must be 10 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)

	}

	if utf8.RuneCountInString(m.GetName()) < 1 {
		err := UpdateClassRequestValidationError{
			field:  "Name",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetDurClass()) < 1 {
		err := UpdateClassRequestValidationError{
			field:  "DurClass",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetWhere()) < 1 {
		err := UpdateClassRequestValidationError{
			field:  "Where",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetTeacher()) < 1 {
		err := UpdateClassRequestValidationError{
			field:  "Teacher",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if m.GetWeeks() <= 0 {
		err := UpdateClassRequestValidationError{
			field:  "Weeks",
			reason: "value must be greater than 0",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetSemester()) != 1 {
		err := UpdateClassRequestValidationError{
			field:  "Semester",
			reason: "value length must be 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)

	}

	if utf8.RuneCountInString(m.GetYear()) < 1 {
		err := UpdateClassRequestValidationError{
			field:  "Year",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if val := m.GetDay(); val < 1 || val > 7 {
		err := UpdateClassRequestValidationError{
			field:  "Day",
			reason: "value must be inside range [1, 7]",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if utf8.RuneCountInString(m.GetClassId()) < 1 {
		err := UpdateClassRequestValidationError{
			field:  "ClassId",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return UpdateClassRequestMultiError(errors)
	}

	return nil
}

// UpdateClassRequestMultiError is an error wrapping multiple validation errors
// returned by UpdateClassRequest.ValidateAll() if the designated constraints
// aren't met.
type UpdateClassRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m UpdateClassRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m UpdateClassRequestMultiError) AllErrors() []error { return m }

// UpdateClassRequestValidationError is the validation error returned by
// UpdateClassRequest.Validate if the designated constraints aren't met.
type UpdateClassRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e UpdateClassRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e UpdateClassRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e UpdateClassRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e UpdateClassRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e UpdateClassRequestValidationError) ErrorName() string {
	return "UpdateClassRequestValidationError"
}

// Error satisfies the builtin error interface
func (e UpdateClassRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sUpdateClassRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = UpdateClassRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = UpdateClassRequestValidationError{}

// Validate checks the field values on UpdateClassResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *UpdateClassResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on UpdateClassResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// UpdateClassResponseMultiError, or nil if none found.
func (m *UpdateClassResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *UpdateClassResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Msg

	// no validation rules for ClassId

	if len(errors) > 0 {
		return UpdateClassResponseMultiError(errors)
	}

	return nil
}

// UpdateClassResponseMultiError is an error wrapping multiple validation
// errors returned by UpdateClassResponse.ValidateAll() if the designated
// constraints aren't met.
type UpdateClassResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m UpdateClassResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m UpdateClassResponseMultiError) AllErrors() []error { return m }

// UpdateClassResponseValidationError is the validation error returned by
// UpdateClassResponse.Validate if the designated constraints aren't met.
type UpdateClassResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e UpdateClassResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e UpdateClassResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e UpdateClassResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e UpdateClassResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e UpdateClassResponseValidationError) ErrorName() string {
	return "UpdateClassResponseValidationError"
}

// Error satisfies the builtin error interface
func (e UpdateClassResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sUpdateClassResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = UpdateClassResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = UpdateClassResponseValidationError{}

// Validate checks the field values on ClassInfo with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ClassInfo) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ClassInfo with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ClassInfoMultiError, or nil
// if none found.
func (m *ClassInfo) ValidateAll() error {
	return m.validate(true)
}

func (m *ClassInfo) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Day

	// no validation rules for Teacher

	// no validation rules for Where

	// no validation rules for ClassWhen

	// no validation rules for WeekDuration

	// no validation rules for Classname

	// no validation rules for Credit

	// no validation rules for Weeks

	// no validation rules for Semester

	// no validation rules for Year

	// no validation rules for Id

	if len(errors) > 0 {
		return ClassInfoMultiError(errors)
	}

	return nil
}

// ClassInfoMultiError is an error wrapping multiple validation errors returned
// by ClassInfo.ValidateAll() if the designated constraints aren't met.
type ClassInfoMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ClassInfoMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ClassInfoMultiError) AllErrors() []error { return m }

// ClassInfoValidationError is the validation error returned by
// ClassInfo.Validate if the designated constraints aren't met.
type ClassInfoValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ClassInfoValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ClassInfoValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ClassInfoValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ClassInfoValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ClassInfoValidationError) ErrorName() string { return "ClassInfoValidationError" }

// Error satisfies the builtin error interface
func (e ClassInfoValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sClassInfo.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ClassInfoValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ClassInfoValidationError{}

// Validate checks the field values on Class with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Class) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Class with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in ClassMultiError, or nil if none found.
func (m *Class) ValidateAll() error {
	return m.validate(true)
}

func (m *Class) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetInfo()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ClassValidationError{
					field:  "Info",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ClassValidationError{
					field:  "Info",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetInfo()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ClassValidationError{
				field:  "Info",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for Thisweek

	if len(errors) > 0 {
		return ClassMultiError(errors)
	}

	return nil
}

// ClassMultiError is an error wrapping multiple validation errors returned by
// Class.ValidateAll() if the designated constraints aren't met.
type ClassMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ClassMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ClassMultiError) AllErrors() []error { return m }

// ClassValidationError is the validation error returned by Class.Validate if
// the designated constraints aren't met.
type ClassValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ClassValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ClassValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ClassValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ClassValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ClassValidationError) ErrorName() string { return "ClassValidationError" }

// Error satisfies the builtin error interface
func (e ClassValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sClass.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ClassValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ClassValidationError{}