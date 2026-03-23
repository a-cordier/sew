package main

import (
	_ "embed"
	"os"
	"runtime/debug"

	"github.com/a-cordier/sew/cmd"
)

var version = "dev"

func init() {
	if version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
}

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
