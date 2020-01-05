package configs

import (
	"log"
	"reflect"
	"strconv"
	"strings"
)

// LogWithPrefix prints all the environment variables and their values on
// container to stdout, excluding any which include the name "password" (for security)
func LogWithPrefix(container interface{}, prefix string) {
	visit(container, logger(prefix))
}

// logger returns a Visitor that logs each value, except for ones with
// "password" somewhere in the key,
//
// This can be used to print config values on app startup, without
// compromising any credentials.
func logger(prefix string) visitor {
	return visitor(func(environment string, value reflect.Value) *visitError {
		logUnlessPassword(prefix+environment, value)
		return nil
	})
}

func logUnlessPassword(environment string, value reflect.Value) {
	if strings.Contains(strings.ToLower(environment), "password") {
		log.Printf("%s: <redacted>", environment)
	} else {
		switch value.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			log.Printf("%s: %s", environment, strconv.FormatUint(value.Uint(), 10))
		default:
			log.Printf("%s: %#v", environment, value)
		}
	}
}
