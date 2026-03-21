package logger

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

var logFilePath string

func SetLogFile(path string) {
	logFilePath = path
}

func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", color.GreenString("✓"), color.BlueString(msg))
}

func Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", color.YellowString("⚠"), color.YellowString(msg))
}

func WithSpinner(message string, fn func() error) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Start()

	start := time.Now()
	err := fn()
	s.Stop()

	if err != nil {
		fmt.Printf("  %s %s\n", color.RedString("✗"), color.RedString("%s (failed after %s)", message, time.Since(start).Round(time.Millisecond)))
		if logFilePath != "" {
			color.Yellow("    See logs: %s", logFilePath)
		}
		return err
	}

	Success("%s (%s)", message, time.Since(start).Round(time.Millisecond))
	return nil
}
