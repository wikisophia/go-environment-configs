package configs_test

import (
	"os"
	"testing"

	configs "github.com/wikisophia/go-environment-configs"
)

type Config struct {
	Boolean     bool     `environment:"BOOLEAN"`
	Int         int      `environment:"INT"`
	String      string   `environment:"STRING"`
	IntSlice    []int    `environment:"INT_SLICE"`
	StringSlice []string `environment:"STRING_SLICE"`
	Nested      *Nested  `environment:"NESTED"`
}

type Nested struct {
	Value int `environment:"VALUE"`
}

func TestWellFormedValues(t *testing.T) {
	defer setEnv(t, "MY_BOOLEAN", "true")()
	defer setEnv(t, "MY_INT", "10")()
	defer setEnv(t, "MY_STRING", "someString")()
	defer setEnv(t, "MY_INT_SLICE", "1,2")()
	defer setEnv(t, "MY_STRING_SLICE", "abc,def")()
	defer setEnv(t, "MY_NESTED_VALUE", "20")()

	cfg := Config{
		Nested: &Nested{},
	}
	if err := configs.Parse("MY", &cfg); err != nil {
		t.Errorf("Got unexpected Parse() error: %v", err)
		return
	}
	assertBoolsEqual(t, true, cfg.Boolean)
	assertStringsEqual(t, "someString", cfg.String)
	assertIntsEqual(t, 10, cfg.Int)
	assertIntSlicesEqual(t, []int{1, 2}, cfg.IntSlice)
	assertStringSlicesEqual(t, []string{"abc", "def"}, cfg.StringSlice)
	assertIntsEqual(t, 20, cfg.Nested.Value)
}

func TestBadValues(t *testing.T) {
	defer setEnv(t, "MY_INT", "foo")()
	defer setEnv(t, "MY_BOOLEAN", "3")()
	defer setEnv(t, "MY_INT_SLICE", "1,foo,2")()
	defer setEnv(t, "MY_NESTED_VALUE", "bar")()
	cfg := Config{
		Nested: &Nested{},
	}
	err := configs.Parse("MY", &cfg)
	if err == nil {
		t.Errorf("Missing expected Parse() error: %v", err)
		return
	}

	assertBoolsEqual(t, false, err.IsValid("MY_INT"))
	assertBoolsEqual(t, false, err.IsValid("MY_BOOLEAN"))
	assertBoolsEqual(t, false, err.IsValid("MY_INT_SLICE"))
	assertBoolsEqual(t, false, err.IsValid("MY_NESTED_VALUE"))
}

func assertStringsEqual(t *testing.T, expected string, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf(`Expected "%s" does not match actual "%s"`, expected, actual)
	}
}

func assertStringSlicesEqual(t *testing.T, expected []string, actual []string) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf(`Expected "%v" does not match actual "%v". The number of elements differ`, expected, actual)
		return
	}
	for i := 0; i < len(expected); i++ {
		assertStringsEqual(t, expected[i], actual[i])
	}
}

func assertIntSlicesEqual(t *testing.T, expected []int, actual []int) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf(`Expected "%v" does not match actual "%v". The number of elements differ`, expected, actual)
		return
	}
	for i := 0; i < len(expected); i++ {
		assertIntsEqual(t, expected[i], actual[i])
	}
}

func assertIntsEqual(t *testing.T, expected int, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf(`Expected "%d" does not match actual "%d"`, expected, actual)
	}
}

func assertBoolsEqual(t *testing.T, expected bool, actual bool) {
	t.Helper()
	if expected != actual {
		t.Errorf(`Expected "%t" does not match actual "%t"`, expected, actual)
	}
}

// setEnv acts as a wrapper around os.Setenv, returning a function that resets the environment
// back to its original value. This prevents tests from setting environment variables as a side-effect.
func setEnv(t *testing.T, key string, val string) func() {
	t.Helper()
	orig, set := os.LookupEnv(key)
	if err := os.Setenv(key, val); err != nil {
		t.Errorf("Error setting environment value %s: %v", key, err)
		return func() {}
	}

	if set {
		return func() {
			os.Setenv(key, orig)
		}
	} else {
		return func() {
			os.Unsetenv(key)
		}
	}
}
