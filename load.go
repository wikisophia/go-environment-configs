package configs

import (
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// MustLoader behaves like Loader, except it panics instead of returning errors.
func MustLoader(prefix string) Visitor {
	delegate := Loader(prefix)
	return Visitor(func(environment string, value reflect.Value) *VisitError {
		if err := delegate(environment, value); err != nil {
			panic(err)
		}
		return nil
	})
}

// Loader returns a Visitor which populates the struct's properties with
// environment variables.
func Loader(prefix string) Visitor {
	return Visitor(func(environment string, value reflect.Value) *VisitError {
		environment = prefix + environment
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
		case reflect.Struct:
			switch value.Type().String() {
			case "big.Int":
				return parseAndSetBigInt(environment, value, environmentValue)
			default:
				panic("loadEnvironmentVisitor() hasn't yet implemented parsing for type " + value.Type().String())
			}
		case reflect.Ptr:
			switch value.Type().String() {
			case "*big.Int":
				return parseAndSetBigIntPointer(environment, value, environmentValue)
			default:
				panic("loadEnvironmentVisitor() hasn't yet implemented parsing for type " + value.Type().String())
			}
		default:
			panic("loadEnvironmentVisitor() hasn't yet implemented parsing for type " + value.String())
		}
	})
}

func parseAndSetBool(env string, toSet reflect.Value, value string) *VisitError {
	switch value {
	case "true":
		toSet.SetBool(true)
	case "false":
		toSet.SetBool(false)
	default:
		return &VisitError{
			error: fmt.Errorf(`%s must be "true" or "false". Got "%s"`, env, value),
			Key:   env,
		}
	}
	return nil
}

func parseAndSetInt(env string, toSet reflect.Value, value string) *VisitError {
	parsed, err := parseInt(value)
	if err != nil {
		return &VisitError{
			error: fmt.Errorf("%s must be an int. Got \"%s\"", env, value),
			Key:   env,
		}
	}
	toSet.SetInt(parsed)
	return nil
}

func parseInt(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func parseAndSetBigInt(env string, toSet reflect.Value, value string) *VisitError {
	parsed, ok := parseBigInt(value)
	if !ok {
		return &VisitError{
			error: fmt.Errorf("%s must be a base-10 big.Int. Got \"%s\"", env, value),
			Key:   env,
		}
	}
	toSet.Set(reflect.ValueOf(parsed))
	return nil
}

func parseAndSetBigIntPointer(env string, toSet reflect.Value, value string) *VisitError {
	parsed, ok := parseBigInt(value)
	if !ok {
		return &VisitError{
			error: fmt.Errorf("%s must be a base-10 big.Int. Got \"%s\"", env, value),
			Key:   env,
		}
	}
	toSet.Set(reflect.ValueOf(&parsed))
	return nil
}

func parseBigInt(value string) (big.Int, bool) {
	parsed := big.Int{}
	_, ok := parsed.SetString(value, 10)
	return parsed, ok
}

func parseCommaSeparatedStrings(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}

func parseAndSetIntSlice(env string, toSet reflect.Value, value string) *VisitError {
	parsed, err := parseCommaSeparatedInts(value)
	if err != nil {
		return &VisitError{
			error: err,
			Key:   env,
		}
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
