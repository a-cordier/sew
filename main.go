package main

import (
	_ "embed"
	"os"

	"github.com/a-cordier/sew/cmd"
)

var version = "dev"

//go:embed sew.yaml
var defaultConfigData []byte

//go:embed schema/sew.schema.yaml
var schemaData []byte

func main() {
	cmd.Version = version
	cmd.DefaultConfigData = defaultConfigData
	cmd.SchemaData = schemaData
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
