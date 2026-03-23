package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/internal/logger"
	"github.com/a-cordier/sew/internal/registry"
	internalschema "github.com/a-cordier/sew/internal/schema"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [path...]",
	Short: "Validate sew.yaml and context flag files against the configuration schema",
	Long: `Validate one or more sew.yaml files against the sew configuration schema.

Each argument can be a path to a sew.yaml file or a directory. When a
directory is given, all sew.yaml and sew--*.yaml (context flag) files
under it are validated recursively. Context flag files are additionally
checked for a valid naming convention and a non-empty description field.

When no argument is given, validates ./sew.yaml in the current directory.`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func isFlagFile(name string) bool {
	return strings.HasPrefix(name, "sew--") && strings.HasSuffix(name, ".yaml")
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

	var configFiles []string
	var flagFiles []string
	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			return fmt.Errorf("cannot access %s: %w", target, err)
		}
		if !info.IsDir() {
			if isFlagFile(filepath.Base(target)) {
				flagFiles = append(flagFiles, target)
			} else {
				configFiles = append(configFiles, target)
			}
			continue
		}
		if err := filepath.Walk(target, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return nil
			}
			name := fi.Name()
			if name == "sew.yaml" {
				configFiles = append(configFiles, path)
			} else if isFlagFile(name) {
				flagFiles = append(flagFiles, path)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("walking %s: %w", target, err)
		}
	}

	total := len(configFiles) + len(flagFiles)
	if total == 0 {
		return fmt.Errorf("no sew.yaml or flag files found")
	}

	var failed int
	for _, f := range configFiles {
		if err := internalschema.ValidateFile(sch, f); err != nil {
			logger.Error("%s: %v", f, err)
			failed++
		}
	}
	for _, f := range flagFiles {
		if err := validateFlagFile(sch, f); err != nil {
			logger.Error("%s: %v", f, err)
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d file(s) failed validation", failed)
	}

	logger.Success("%d file(s) valid", len(configFiles)+len(flagFiles))
	return nil
}

// validateFlagFile validates a context flag file: naming convention,
// description required, and schema compliance.
func validateFlagFile(sch *jsonschema.Schema, path string) error {
	name := filepath.Base(path)
	if _, err := registry.FlagNameFromFile(name); err != nil {
		return err
	}

	if err := internalschema.ValidateFile(sch, path); err != nil {
		return fmt.Errorf("schema: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	if err := registry.ValidateFlagDescription(data); err != nil {
		return err
	}

	return nil
}
