package configs

import (
	"log"
	"reflect"
	"strings"
)

// Logger returns a Visitor that logs each value, except for ones with
// "password" somewhere in the key,
//
// This can be used to print config values on app startup, without
// compromising any credentials.
func Logger(prefix string) Visitor {
	return Visitor(func(environment string, value reflect.Value) *VisitError {
		logUnlessPassword(prefix+environment, value)
		return nil
	})
}

func logUnlessPassword(environment string, value reflect.Value) {
	if strings.Contains(strings.ToLower(environment), "password") {
		log.Printf("%s: <redacted>", environment)
	} else {
		log.Printf("%s: %#v", environment, value)
	}
}
