// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"unicode/utf8"
)

var (
	// ErrLengthTooLong is the error that returns in case of too long length.
	ErrLengthTooLong = NewError("validation_length_too_long", "the length must be no more than {{.max}}")
	// ErrLengthTooShort is the error that returns in case of too short length.
	ErrLengthTooShort = NewError("validation_length_too_short", "the length must be no less than {{.min}}")
	// ErrLengthInvalid is the error that returns in case of an invalid length.
	ErrLengthInvalid = NewError("validation_length_invalid", "the length must be exactly {{.min}}")
	// ErrLengthOutOfRange is the error that returns in case of out of range length.
	ErrLengthOutOfRange = NewError("validation_length_out_of_range", "the length must be between {{.min}} and {{.max}}")
	// ErrLengthEmptyRequired is the error that returns in case of non-empty value.
	ErrLengthEmptyRequired = NewError("validation_length_empty_required", "the value must be empty")
)

// Length returns a validation rule that checks if a value's length is within the specified range.
// If max is 0, it means there is no upper bound for the length.
// This rule should only be used for validating strings, slices, maps, and arrays.
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func Length(min, max int) LengthRule {
	return LengthRule{min: min, max: max, err: buildLengthRuleError(min, max)}
}

// RuneLength returns a validation rule that checks if a string's rune length is within the specified range.
// If max is 0, it means there is no upper bound for the length.
// This rule should only be used for validating strings, slices, maps, and arrays.
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
// If the value being validated is not a string, the rule works the same as Length.
func RuneLength(min, max int) LengthRule {
	r := Length(min, max)
	r.rune = true

	return r
}

// LengthRule is a validation rule that checks if a value's length is within the specified range.
type LengthRule struct {
	err Error

	min, max int
	rune     bool
}

// Validate checks if the given value is valid or not.
func (r LengthRule) Validate(value interface{}) error {
	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	var (
		l   int
		err error
	)
	if s, ok := value.(string); ok && r.rune {
		l = utf8.RuneCountInString(s)
	} else if l, err = LengthOfValue(value); err != nil {
		return err
	}

	if r.min > 0 && l < r.min || r.max > 0 && l > r.max || r.min == 0 && r.max == 0 && l > 0 {
		return r.err
	}

	return nil
}

// Error sets the error message for the rule.
func (r LengthRule) Error(message string) LengthRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r LengthRule) ErrorObject(err Error) LengthRule {
	r.err = err
	return r
}

func buildLengthRuleError(min, max int) (err Error) {
	if min == 0 && max > 0 {
		err = ErrLengthTooLong
	} else if min > 0 && max == 0 {
		err = ErrLengthTooShort
	} else if min > 0 && max > 0 {
		if min == max {
			err = ErrLengthInvalid
		} else {
			err = ErrLengthOutOfRange
		}
	} else {
		err = ErrLengthEmptyRequired
	}

	return err.SetParams(map[string]interface{}{"min": min, "max": max})
}
