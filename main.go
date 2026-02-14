package main

import (
	"os"

	"github.com/a-cordier/sew/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
