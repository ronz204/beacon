package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "target missing")
		os.Exit(2)
	}

	switch os.Args[1] {
	case "pass":
		fmt.Println("all assertions hold")
	case "fail":
		fmt.Fprintln(os.Stderr, "assertion failed: lockout after 5 attempts")
		os.Exit(1)
	default:
		fmt.Fprintln(os.Stderr, "unknown target", os.Args[1])
		os.Exit(2)
	}
}
