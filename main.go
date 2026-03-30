package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/sophylax/envguard/cmd"
)

var version = "dev"

func main() {
	if err := cmd.Execute(version); err != nil {
		if !errors.Is(err, cmd.ErrFindings) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
