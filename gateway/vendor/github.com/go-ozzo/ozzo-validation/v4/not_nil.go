// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

// ErrNotNilRequired is the error that returns when a value is Nil.
var ErrNotNilRequired = NewError("validation_not_nil_required", "is required")

// NotNil is a validation rule that checks if a value is not nil.
// NotNil only handles types including interface, pointer, slice, and map.
// All other types are considered valid.
var NotNil = notNilRule{}

type notNilRule struct {
	err Error
}

// Validate checks if the given value is valid or not.
func (r notNilRule) Validate(value interface{}) error {
	_, isNil := Indirect(value)
	if isNil {
		if r.err != nil {
			return r.err
		}
		return ErrNotNilRequired
	}
	return nil
}

// Error sets the error message for the rule.
func (r notNilRule) Error(message string) notNilRule {
	if r.err == nil {
		r.err = ErrNotNilRequired
	}
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r notNilRule) ErrorObject(err Error) notNilRule {
	r.err = err
	return r
}
