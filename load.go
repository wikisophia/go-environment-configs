package configs

import (
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// MustLoadWithPrefix loads the environment variables into a struct.
// It panics if any of the environment variables' values can't be
// coerced into the type defined on the struct.
func MustLoadWithPrefix(container interface{}, prefix string) {
	err := LoadWithPrefix(container, prefix)
	if err != nil {
		panic(err)
	}
}

// LoadWithPrefix loads the values of environment variables into a struct.
// It returns an error if any of the environment variable values don't match
// the type defined on the struct.
func LoadWithPrefix(container interface{}, prefix string) error {
	return visit(container, loader(prefix))
}

// loader returns a visitor which populates the struct's properties with
// environment variables.
func loader(prefix string) visitor {
	return visitor(func(environment string, value reflect.Value) *visitError {
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

func parseAndSetBool(env string, toSet reflect.Value, value string) *visitError {
	switch value {
	case "true":
		toSet.SetBool(true)
	case "false":
		toSet.SetBool(false)
	default:
		return &visitError{
			error: fmt.Errorf(`must be "true" or "false": got "%s"`, value),
			Key:   env,
		}
	}
	return nil
}

func parseAndSetInt(env string, toSet reflect.Value, value string) *visitError {
	parsed, err := parseInt(value)
	if err != nil {
		return &visitError{
			error: fmt.Errorf("must be an int: got \"%s\"", value),
			Key:   env,
		}
	}
	toSet.SetInt(parsed)
	return nil
}

func parseInt(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func parseAndSetBigInt(env string, toSet reflect.Value, value string) *visitError {
	parsed, ok := parseBigInt(value)
	if !ok {
		return &visitError{
			error: fmt.Errorf("must be a base-10 big.Int: got \"%s\"", value),
			Key:   env,
		}
	}
	toSet.Set(reflect.ValueOf(parsed))
	return nil
}

func parseAndSetBigIntPointer(env string, toSet reflect.Value, value string) *visitError {
	parsed, ok := parseBigInt(value)
	if !ok {
		return &visitError{
			error: fmt.Errorf("must be a base-10 big.Int: got \"%s\"", value),
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

func parseAndSetIntSlice(env string, toSet reflect.Value, value string) *visitError {
	parsed, err := parseCommaSeparatedInts(value)
	if err != nil {
		return &visitError{
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
			return nil, fmt.Errorf(`must be a comma-separated list of ints: got "%s" which contains a non-int at index %d`, value, i)
		}
		intSlice[i] = parsed
	}
	return intSlice, nil
}
