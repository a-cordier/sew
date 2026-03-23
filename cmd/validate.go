package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/a-cordier/sew/internal/logger"
	internalschema "github.com/a-cordier/sew/internal/schema"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [path...]",
	Short: "Validate sew.yaml files against the configuration schema",
	Long: `Validate one or more sew.yaml files against the sew configuration schema.

Each argument can be a path to a sew.yaml file or a directory. When a
directory is given, all sew.yaml files under it are validated recursively.
When no argument is given, validates ./sew.yaml in the current directory.`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(_ *cobra.Command, args []string) error {
	sch, err := internalschema.Compile(SchemaData)
	if err != nil {
		return fmt.Errorf("compiling schema: %w", err)
	}

	targets := args
	if len(targets) == 0 {
		targets = []string{"sew.yaml"}
	}

	var files []string
	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			return fmt.Errorf("cannot access %s: %w", target, err)
		}
		if !info.IsDir() {
			files = append(files, target)
			continue
		}
		if err := filepath.Walk(target, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !fi.IsDir() && fi.Name() == "sew.yaml" {
				files = append(files, path)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("walking %s: %w", target, err)
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("no sew.yaml files found")
	}

	var failed int
	for _, f := range files {
		if err := internalschema.ValidateFile(sch, f); err != nil {
			logger.Error("%s: %v", f, err)
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d file(s) failed validation", failed)
	}

	logger.Success("%d file(s) valid", len(files))
	return nil
}
