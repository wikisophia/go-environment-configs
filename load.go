package configs

import (
	"errors"
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
		case reflect.Uint64:
			return parseAndSetUInt(environment, value, environmentValue, 64)
		case reflect.Uint32:
			return parseAndSetUInt(environment, value, environmentValue, 32)
		case reflect.Uint16:
			return parseAndSetUInt(environment, value, environmentValue, 16)
		case reflect.Uint8:
			return parseAndSetUInt(environment, value, environmentValue, 8)
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
			error: errors.New(`must be "true" or "false"`),
			Key:   env,
		}
	}
	return nil
}

func parseAndSetInt(env string, toSet reflect.Value, value string) *visitError {
	parsed, err := parseInt(value)
	if err != nil {
		return &visitError{
			error: errors.New("must be an int"),
			Key:   env,
		}
	}
	toSet.SetInt(parsed)
	return nil
}

func parseAndSetUInt(env string, toSet reflect.Value, value string, bitSize int) *visitError {
	parsed, err := strconv.ParseUint(value, 10, bitSize)
	if casted, ok := err.(*strconv.NumError); ok && casted != nil {
		if casted.Err == strconv.ErrRange {
			return &visitError{
				error: fmt.Errorf("has a max value of %d", parsed),
				Key:   env,
			}
		}
		if _, err := strconv.ParseInt(value, 10, 64); err == nil {
			return &visitError{
				error: errors.New("has a min value of 0"),
				Key:   env,
			}
		}
		return &visitError{
			error: errors.New("must be a uint" + strconv.FormatInt(int64(bitSize), 10)),
			Key:   env,
		}
	}
	toSet.SetUint(parsed)
	return nil
}

func parseInt(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func parseAndSetBigInt(env string, toSet reflect.Value, value string) *visitError {
	parsed, ok := parseBigInt(value)
	if !ok {
		return &visitError{
			error: errors.New("must be a base-10 big.Int"),
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
			error: errors.New("must be a base-10 big.Int"),
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
			return nil, fmt.Errorf(`must be a comma-separated list of ints: index %d is invalid`, i)
		}
		intSlice[i] = parsed
	}
	return intSlice, nil
}
