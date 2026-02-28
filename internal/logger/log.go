package logger

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

var logFilePath string

func SetLogFile(path string) {
	logFilePath = path
}

func WithSpinner(message string, fn func() error) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Start()

	start := time.Now()
	err := fn()
	s.Stop()

	if err != nil {
		color.Red("  ✗ %s (failed after %s)", message, time.Since(start).Round(time.Millisecond))
		if logFilePath != "" {
			color.Yellow("    See logs: %s", logFilePath)
		}
		return err
	}

	color.Blue("  ✓ %s (%s)", message, time.Since(start).Round(time.Millisecond))
	return nil
}
