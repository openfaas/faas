// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

var (
	// ErrNil is the error that returns when a value is not nil.
	ErrNil = NewError("validation_nil", "must be blank")
	// ErrEmpty is the error that returns when a not nil value is not empty.
	ErrEmpty = NewError("validation_empty", "must be blank")
)

// Nil is a validation rule that checks if a value is nil.
// It is the opposite of NotNil rule
var Nil = absentRule{condition: true, skipNil: false}

// Empty checks if a not nil value is empty.
var Empty = absentRule{condition: true, skipNil: true}

type absentRule struct {
	condition bool
	err       Error
	skipNil   bool
}

// Validate checks if the given value is valid or not.
func (r absentRule) Validate(value interface{}) error {
	if r.condition {
		value, isNil := Indirect(value)
		if !r.skipNil && !isNil || r.skipNil && !isNil && !IsEmpty(value) {
			if r.err != nil {
				return r.err
			}
			if r.skipNil {
				return ErrEmpty
			}
			return ErrNil
		}
	}
	return nil
}

// When sets the condition that determines if the validation should be performed.
func (r absentRule) When(condition bool) absentRule {
	r.condition = condition
	return r
}

// Error sets the error message for the rule.
func (r absentRule) Error(message string) absentRule {
	if r.err == nil {
		if r.skipNil {
			r.err = ErrEmpty
		} else {
			r.err = ErrNil
		}
	}
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r absentRule) ErrorObject(err Error) absentRule {
	r.err = err
	return r
}
