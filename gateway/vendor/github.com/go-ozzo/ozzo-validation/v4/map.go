package validation

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrNotMap is the error that the value being validated is not a map.
	ErrNotMap = errors.New("only a map can be validated")

	// ErrKeyWrongType is the error returned in case of an incorrect key type.
	ErrKeyWrongType = NewError("validation_key_wrong_type", "key not the correct type")

	// ErrKeyMissing is the error returned in case of a missing key.
	ErrKeyMissing = NewError("validation_key_missing", "required key is missing")

	// ErrKeyUnexpected is the error returned in case of an unexpected key.
	ErrKeyUnexpected = NewError("validation_key_unexpected", "key not expected")
)

type (
	// MapRule represents a rule set associated with a map.
	MapRule struct {
		keys           []*KeyRules
		allowExtraKeys bool
	}

	// KeyRules represents a rule set associated with a map key.
	KeyRules struct {
		key      interface{}
		optional bool
		rules    []Rule
	}
)

// Map returns a validation rule that checks the keys and values of a map.
// This rule should only be used for validating maps, or a validation error will be reported.
// Use Key() to specify map keys that need to be validated. Each Key() call specifies a single key which can
// be associated with multiple rules.
// For example,
//    validation.Map(
//        validation.Key("Name", validation.Required),
//        validation.Key("Value", validation.Required, validation.Length(5, 10)),
//    )
//
// A nil value is considered valid. Use the Required rule to make sure a map value is present.
func Map(keys ...*KeyRules) MapRule {
	return MapRule{keys: keys}
}

// AllowExtraKeys configures the rule to ignore extra keys.
func (r MapRule) AllowExtraKeys() MapRule {
	r.allowExtraKeys = true
	return r
}

// Validate checks if the given value is valid or not.
func (r MapRule) Validate(m interface{}) error {
	return r.ValidateWithContext(nil, m)
}

// ValidateWithContext checks if the given value is valid or not.
func (r MapRule) ValidateWithContext(ctx context.Context, m interface{}) error {
	value := reflect.ValueOf(m)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Map {
		// must be a map
		return NewInternalError(ErrNotMap)
	}
	if value.IsNil() {
		// treat a nil map as valid
		return nil
	}

	errs := Errors{}
	kt := value.Type().Key()

	var extraKeys map[interface{}]bool
	if !r.allowExtraKeys {
		extraKeys = make(map[interface{}]bool, value.Len())
		for _, k := range value.MapKeys() {
			extraKeys[k.Interface()] = true
		}
	}

	for _, kr := range r.keys {
		var err error
		if kv := reflect.ValueOf(kr.key); !kt.AssignableTo(kv.Type()) {
			err = ErrKeyWrongType
		} else if vv := value.MapIndex(kv); !vv.IsValid() {
			if !kr.optional {
				err = ErrKeyMissing
			}
		} else if ctx == nil {
			err = Validate(vv.Interface(), kr.rules...)
		} else {
			err = ValidateWithContext(ctx, vv.Interface(), kr.rules...)
		}
		if err != nil {
			if ie, ok := err.(InternalError); ok && ie.InternalError() != nil {
				return err
			}
			errs[getErrorKeyName(kr.key)] = err
		}
		if !r.allowExtraKeys {
			delete(extraKeys, kr.key)
		}
	}

	if !r.allowExtraKeys {
		for key := range extraKeys {
			errs[getErrorKeyName(key)] = ErrKeyUnexpected
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Key specifies a map key and the corresponding validation rules.
func Key(key interface{}, rules ...Rule) *KeyRules {
	return &KeyRules{
		key:   key,
		rules: rules,
	}
}

// Optional configures the rule to ignore the key if missing.
func (r *KeyRules) Optional() *KeyRules {
	r.optional = true
	return r
}

// getErrorKeyName returns the name that should be used to represent the validation error of a map key.
func getErrorKeyName(key interface{}) string {
	return fmt.Sprintf("%v", key)
}
