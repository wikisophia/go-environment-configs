
[![BuildStatus](https://travis-ci.org/wikisophia/go-environment-configs.svg?branch=master)](https://travis-ci.org/wikisophia/go-environment-configs)
[![ReportCard](https://goreportcard.com/badge/github.com/wikisophia/go-environment-configs)](https://goreportcard.com/report/github.com/wikisophia/go-environment-configs)
[![GoDoc](https://godoc.org/github.com/wikisophia/go-environment-configs?status.svg)](https://godoc.org/github.com/wikisophia/go-environment-configs)

# Overview

This library helps applications work with configs through environment variables.
It's a more opinionated and much smaller version of
[Viper](https://github.com/spf13/viper). It assumes you're following the
12 factor app recommendations for [configs](https://12factor.net/config) and
[logging](https://12factor.net/logs).

# Goals

Applications generally want to do the following:

1. Define some default config values which work in most cases.
2. Overwrite defaults with environment variables.
3. Validate the config values.
4. Log any config values _except_ for credentials.

This library's goal is to make this as painless as possible.

# Usage

Define structs and tag them with the environment variables you want:

```go
type struct Config {
  Main Server `environment:"MAIN"`
  Admin Server `environment:"ADMIN"`
  Password string `environment:"PASSWORD"`
}

type struct Server {
  Port int `environment:"PORT"`
}
```

Set some environment variables in your shell.

```sh
export MYAPP_MAIN_PORT=80
export MYAPP_ADMIN_PORT=81
export MYAPP_PASSWORD=boo
```

Use them like this:

```go
import (
	"fmt"
	"log"
	"os"

	"github.com/wikisophia/go-environment-configs"
)

func Parse() Config {
  // Intiial values serve as defaults.
  cfg := Config{
    Main: Server{
      Port: 80
    },
    Admin: Server{
      Port: 81
    }
  }

  // Overwrite the defaults with environment variables.
  // Panic with a descriptive error message if the values don't match the types.
  configs.Visit(&cfg, configs.MustLoader("MYAPP"))

  // Print the config values.
  // Anything named "password" will be logged as "<redacted>"
	configs.Visit(&cfg, configs.Logger("MYAPP"))
}
```

# Contributing

This library doesn't yet support all the struct property types...
just the ones that I've needed for projects so far.
This covers the common ones, but certainly not all of them.
Feel free to contribute support for any new types you need.
