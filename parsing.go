package configs

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// MustLoad wraps Load, but panics if an error occurs.
func MustLoad(prefix string, container interface{}) {
	err := Load(prefix, container)
	if err != nil {
		panic(err)
	}
}

// Load populates the container with environment variables.
// The container must be a pointer to a struct whose properties have
// "environment" tags.
func Load(prefix string, container interface{}) *LoadError {
	return visit(container, prefix, loadEnvironmentVisitor)
}

// LoadError is returned by Load() if anything went wrong while
// initializing the config value.
type LoadError struct {
	invalidKeys map[string]error
}

// IsValid returns true if the environment variable had a valid value,
// and false if it was invalid. For example, the value "foo" would be
// invalid if the struct value for that variable was defined with an int.
func (p *LoadError) IsValid(environment string) bool {
	if p == nil {
		return true
	}

	_, ok := p.invalidKeys[environment]
	return !ok
}

// Append marks an environment variable as invalid,
// with the given error message.
//
// This can be used after Parse() to aggregate "extra" validation errors
// (like "int must be positive" or "string can't be empty") alongside
// those produced by this library.
//
// If err is nil, a new one will be returned
func Append(err *LoadError, environment string, msg string) *LoadError {
	if err == nil {
		return &LoadError{
			invalidKeys: map[string]error{
				environment: errors.New(msg),
			},
		}
	}

	// Defensive in case someone creates an empty LoadError{} manually
	if err.invalidKeys == nil {
		err.invalidKeys = make(map[string]error)
	}

	existing, ok := err.invalidKeys[environment]
	if ok {
		err.invalidKeys[environment] = fmt.Errorf("%v: %s", existing, msg)
	} else {
		err.invalidKeys[environment] = errors.New(msg)
	}
	return err
}

// Error returns an error message describing all the invalid environment variables.
func (p *LoadError) Error() string {
	if p == nil {
		return ""
	}

	msg := strings.Builder{}
	msg.WriteString("Invalid environment variables for app config:\n")
	for env, err := range p.invalidKeys {
		msg.WriteString(fmt.Sprintf("  %s: %v\n", env, err))
	}
	return msg.String()
}

// Visit calls the visitor function on each property on container,
// unless that property is a struct itself. It will recurse through any
// any structs until it eventually gets finds the leaves.
func visit(container interface{}, prefix string, visitor func(value reflect.Value, environment string) error) *LoadError {
	return visitReflectValue(prefix, reflect.ValueOf(container), visitor, nil)
}

func visitReflectValue(environmentSoFar string, theValue reflect.Value, visitor func(value reflect.Value, environment string) error, errs *LoadError) *LoadError {
	theType := theValue.Type().Elem()

	for i := 0; i < theType.NumField(); i++ {
		thisField := theType.Field(i)
		thisFieldValue := theValue.Elem().Field(i)
		environment := environmentSoFar + "_" + thisField.Tag.Get("environment")
		switch thisField.Type.Kind() {
		case reflect.Ptr:
			errs = visitReflectValue(environment, thisFieldValue, visitor, errs)
		default:
			if err := visitor(thisFieldValue, environment); err != nil {
				if errs == nil {
					errs = &LoadError{
						invalidKeys: map[string]error{
							environment: err,
						},
					}
				} else {
					errs.invalidKeys[environment] = err
				}
			}
		}
	}
	return errs
}

func loadEnvironmentVisitor(value reflect.Value, environment string) error {
	environmentValue, isSet := os.LookupEnv(environment)
	if !isSet {
		return nil
	}

	switch value.Kind() {
	case reflect.Bool:
		return parseAndSetBool(environment, value, environmentValue)
	case reflect.Int:
		return parseAndSetInt(environment, value, environmentValue)
	case reflect.String:
		value.SetString(environmentValue)
		return nil
	case reflect.Slice:
		switch value.Type().Elem().Kind() {
		case reflect.String:
			value.Set(reflect.ValueOf(parseCommaSeparatedStrings(environmentValue)))
			return nil
		case reflect.Int:
			return parseAndSetIntSlice(environment, value, environmentValue)
		default:
			panic(fmt.Sprintf("loadEnvironmentVisitor() is not yet implement for slices of type %v", value.Type().Elem().Kind()))
		}
	default:
		panic("loadEnvironmentVisitor() hasn't yet implemented parsing for type " + value.String())
	}
}

func parseAndSetBool(env string, toSet reflect.Value, value string) error {
	switch value {
	case "true":
		toSet.SetBool(true)
	case "false":
		toSet.SetBool(false)
	default:
		return fmt.Errorf(`%s must be "true" or "false". Got "%s"`, env, value)
	}
	return nil
}

func parseAndSetInt(env string, toSet reflect.Value, value string) error {
	parsed, err := parseInt(value)
	if err != nil {
		return fmt.Errorf("%s must be an int. Got \"%s\"", env, value)
	}
	toSet.SetInt(parsed)
	return nil
}

func parseInt(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func parseCommaSeparatedStrings(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}

func parseAndSetIntSlice(env string, toSet reflect.Value, value string) error {
	parsed, err := parseCommaSeparatedInts(value)
	if err != nil {
		return err
	}
	toSet.Set(reflect.ValueOf(parsed))
	return nil
}

func parseCommaSeparatedInts(value string) ([]int, error) {
	if value == "" {
		return nil, nil
	}
	stringSlice := strings.Split(value, ",")
	intSlice := make([]int, len(stringSlice))
	for i := 0; i < len(stringSlice); i++ {
		parsed, err := strconv.Atoi(stringSlice[i])
		if err != nil {
			return nil, fmt.Errorf(`value "%s" contains a non-int at index %d`, value, i)
		}
		intSlice[i] = parsed
	}
	return intSlice, nil
}
