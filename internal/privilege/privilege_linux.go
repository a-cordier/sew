//go:build linux

package privilege

// Elevate is a no-op on Linux where the operations that require
// privilege escalation on macOS work without root.
func Elevate(_ string) error {
	return nil
}
