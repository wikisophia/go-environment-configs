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
func Visit(container interface{}, visitor Visitor) error {
	return doVisit("", reflect.ValueOf(container), visitor, nil)
}

var s struct{}
var terminalTypes = map[string]struct{}{
	"big.Int":  s,
	"*big.Int": s,
}

func doVisit(environmentSoFar string, theValue reflect.Value, visitor Visitor, errs error) error {
	theType := theValue.Type().Elem()

	for i := 0; i < theType.NumField(); i++ {
		thisField := theType.Field(i)
		thisFieldValue := theValue.Elem().Field(i)
		environment := environmentSoFar + "_" + thisField.Tag.Get("environment")
		switch thisField.Type.Kind() {
		case reflect.Ptr:
			if _, ok := terminalTypes[thisField.Type.String()]; ok {
				if err := visitor(environment, thisFieldValue); err != nil {
					errs = Append(errs, err.Key, err)
				}
			} else {
				errs = doVisit(environment, thisFieldValue, visitor, errs)
			}
		default:
			if err := visitor(environment, thisFieldValue); err != nil {
				errs = Append(errs, err.Key, err)
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

// Append adds a custom key/error to the TraversalError. If the input error is nil,
// a new *TraversalError will be returned.
//
// This can be used after Parse() to aggregate "extra" validation errors
// (like "int must be positive" or "string can't be empty") alongside
// those produced by this library.
//
// If err is not a *TraversalError, this will panic.
func Append(err error, key string, msg error) error {
	if err == nil {
		return &TraversalError{
			invalidKeys: map[string]error{
				key: msg,
			},
		}
	}

	if casted, ok := err.(*TraversalError); ok {
		// Defensive in case someone creates an empty LoadError{} manually
		if casted.invalidKeys == nil {
			casted.invalidKeys = make(map[string]error)
		}

		existing, ok := casted.invalidKeys[key]
		if ok {
			casted.invalidKeys[key] = fmt.Errorf("%v: %s", existing, msg)
		} else {
			casted.invalidKeys[key] = msg
		}
		return casted
	}

	panic("Append is only intended to work on *TraversalError types")
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
