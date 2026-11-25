package main

import (
	"os"

	"github.com/ronz204/beacon/source/router"
)

func main() {
	os.Exit(router.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
