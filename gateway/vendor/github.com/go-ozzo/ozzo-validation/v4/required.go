// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

var (
	// ErrRequired is the error that returns when a value is required.
	ErrRequired = NewError("validation_required", "cannot be blank")
	// ErrNilOrNotEmpty is the error that returns when a value is not nil and is empty.
	ErrNilOrNotEmpty = NewError("validation_nil_or_not_empty_required", "cannot be blank")
)

// Required is a validation rule that checks if a value is not empty.
// A value is considered not empty if
// - integer, float: not zero
// - bool: true
// - string, array, slice, map: len() > 0
// - interface, pointer: not nil and the referenced value is not empty
// - any other types
var Required = RequiredRule{skipNil: false, condition: true}

// NilOrNotEmpty checks if a value is a nil pointer or a value that is not empty.
// NilOrNotEmpty differs from Required in that it treats a nil pointer as valid.
var NilOrNotEmpty = RequiredRule{skipNil: true, condition: true}

// RequiredRule is a rule that checks if a value is not empty.
type RequiredRule struct {
	condition bool
	skipNil   bool
	err       Error
}

// Validate checks if the given value is valid or not.
func (r RequiredRule) Validate(value interface{}) error {
	if r.condition {
		value, isNil := Indirect(value)
		if r.skipNil && !isNil && IsEmpty(value) || !r.skipNil && (isNil || IsEmpty(value)) {
			if r.err != nil {
				return r.err
			}
			if r.skipNil {
				return ErrNilOrNotEmpty
			}
			return ErrRequired
		}
	}
	return nil
}

// When sets the condition that determines if the validation should be performed.
func (r RequiredRule) When(condition bool) RequiredRule {
	r.condition = condition
	return r
}

// Error sets the error message for the rule.
func (r RequiredRule) Error(message string) RequiredRule {
	if r.err == nil {
		if r.skipNil {
			r.err = ErrNilOrNotEmpty
		} else {
			r.err = ErrRequired
		}
	}
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r RequiredRule) ErrorObject(err Error) RequiredRule {
	r.err = err
	return r
}
