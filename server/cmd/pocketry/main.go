package main

import (
	"github.com/CDX-1/pocketry/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		panic(err)
	}
}