package main

import (
	"fmt"
	"os"
	"time"

	"blueprint/internal/app"
)

var Version = "dev"

func main() {
	instance := app.App{In: os.Stdin, Out: os.Stdout, Err: os.Stderr, Now: time.Now, Version: Version}
	code, err := instance.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if code != 0 {
		os.Exit(code)
	}
}
