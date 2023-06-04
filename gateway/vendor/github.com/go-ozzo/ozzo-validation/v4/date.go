// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"time"
)

var (
	// ErrDateInvalid is the error that returns in case of an invalid date.
	ErrDateInvalid = NewError("validation_date_invalid", "must be a valid date")
	// ErrDateOutOfRange is the error that returns in case of an invalid date.
	ErrDateOutOfRange = NewError("validation_date_out_of_range", "the date is out of range")
)

// DateRule is a validation rule that validates date/time string values.
type DateRule struct {
	layout        string
	min, max      time.Time
	err, rangeErr Error
}

// Date returns a validation rule that checks if a string value is in a format that can be parsed into a date.
// The format of the date should be specified as the layout parameter which accepts the same value as that for time.Parse.
// For example,
//    validation.Date(time.ANSIC)
//    validation.Date("02 Jan 06 15:04 MST")
//    validation.Date("2006-01-02")
//
// By calling Min() and/or Max(), you can let the Date rule to check if a parsed date value is within
// the specified date range.
//
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func Date(layout string) DateRule {
	return DateRule{
		layout:   layout,
		err:      ErrDateInvalid,
		rangeErr: ErrDateOutOfRange,
	}
}

// Error sets the error message that is used when the value being validated is not a valid date.
func (r DateRule) Error(message string) DateRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct that is used when the value being validated is not a valid date..
func (r DateRule) ErrorObject(err Error) DateRule {
	r.err = err
	return r
}

// RangeError sets the error message that is used when the value being validated is out of the specified Min/Max date range.
func (r DateRule) RangeError(message string) DateRule {
	r.rangeErr = r.rangeErr.SetMessage(message)
	return r
}

// RangeErrorObject sets the error struct that is used when the value being validated is out of the specified Min/Max date range.
func (r DateRule) RangeErrorObject(err Error) DateRule {
	r.rangeErr = err
	return r
}

// Min sets the minimum date range. A zero value means skipping the minimum range validation.
func (r DateRule) Min(min time.Time) DateRule {
	r.min = min
	return r
}

// Max sets the maximum date range. A zero value means skipping the maximum range validation.
func (r DateRule) Max(max time.Time) DateRule {
	r.max = max
	return r
}

// Validate checks if the given value is a valid date.
func (r DateRule) Validate(value interface{}) error {
	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	str, err := EnsureString(value)
	if err != nil {
		return err
	}

	date, err := time.Parse(r.layout, str)
	if err != nil {
		return r.err
	}

	if !r.min.IsZero() && r.min.After(date) || !r.max.IsZero() && date.After(r.max) {
		return r.rangeErr
	}

	return nil
}
