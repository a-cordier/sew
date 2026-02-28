// Package privilege provides platform-specific privilege escalation.
//
// On macOS, CPK requires root for network setup; Elevate uses osascript
// to prompt the user via a GUI dialog. On Linux, no escalation is needed.
package privilege
