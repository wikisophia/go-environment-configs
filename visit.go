package configs

import (
	"fmt"
	"reflect"
	"strings"
)

// Visitor is a function which acts on struct leaf properties.
type Visitor func(environment string, value reflect.Value) *VisitError

// VisitError is an error which can be returned by Visitors if something
// went wrong while running the function.
type VisitError struct {
	error
	// Key describes the leaf node. In general, this can just be the
	// "environment" argument.
	Key string
}

// Visit calls the visitor function on each property on container,
// unless that property is a struct itself. It will recurse through any
// any structs until it eventually gets finds the leaves.
func Visit(container interface{}, visitor Visitor) *TraversalError {
	return doVisit("", reflect.ValueOf(container), visitor, nil)
}

func doVisit(environmentSoFar string, theValue reflect.Value, visitor Visitor, errs *TraversalError) *TraversalError {
	theType := theValue.Type().Elem()

	for i := 0; i < theType.NumField(); i++ {
		thisField := theType.Field(i)
		thisFieldValue := theValue.Elem().Field(i)
		environment := environmentSoFar + "_" + thisField.Tag.Get("environment")
		switch thisField.Type.Kind() {
		case reflect.Ptr:
			errs = doVisit(environment, thisFieldValue, visitor, errs)
		default:
			if err := visitor(environment, thisFieldValue); err != nil {
				if errs == nil {
					errs = &TraversalError{
						invalidKeys: map[string]error{
							err.Key: err,
						},
					}
				} else {
					errs.invalidKeys[err.Key] = err
				}
			}
		}
	}
	return errs
}

// TraversalError is returned by Visit() if the Visitor returned any errors
type TraversalError struct {
	summary     string
	invalidKeys map[string]error
}

// IsValid returns false if the Visitor returned an error at the given
// key during the traversal. It returns true if the Visitor succeeded
// or was never run on this key.
func (p *TraversalError) IsValid(key string) bool {
	if p == nil {
		return true
	}
	_, ok := p.invalidKeys[key]
	return !ok
}

// Append adds a custom key/error to the Traversal set.
//
// This can be used after Parse() to aggregate "extra" validation errors
// (like "int must be positive" or "string can't be empty") alongside
// those produced by this library.
//
// If err is nil, a new one will be returned.
func Append(err *TraversalError, key string, msg error) *TraversalError {
	if err == nil {
		return &TraversalError{
			invalidKeys: map[string]error{
				key: msg,
			},
		}
	}

	// Defensive in case someone creates an empty LoadError{} manually
	if err.invalidKeys == nil {
		err.invalidKeys = make(map[string]error)
	}

	existing, ok := err.invalidKeys[key]
	if ok {
		err.invalidKeys[key] = fmt.Errorf("%v: %s", existing, msg)
	} else {
		err.invalidKeys[key] = msg
	}
	return err
}

// Error returns an error message describing all the invalid environment variables.
func (p *TraversalError) Error() string {
	if p == nil {
		return ""
	}

	msg := strings.Builder{}
	msg.WriteString("Errors occurred while acting on the struct:\n")
	for env, err := range p.invalidKeys {
		msg.WriteString(fmt.Sprintf("  %s: %v\n", env, err))
	}
	return msg.String()
}
