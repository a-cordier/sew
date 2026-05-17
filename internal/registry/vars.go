package registry

import (
	"strings"
)

// SetOverrides holds --set values split into scoped (dotted key targeting
// a specific context path) and broadcast (plain key applied to every
// context that declares the variable).
type SetOverrides struct {
	Scoped    map[string]string // keys containing "." (e.g. "mysql.standalone.imageTag")
	Broadcast map[string]string // keys without "." (e.g. "imageTag")
}

// SplitSetOverrides partitions a flat --set map into scoped and broadcast
// buckets based on the presence of a dot in the key.
func SplitSetOverrides(raw map[string]string) SetOverrides {
	s := SetOverrides{
		Scoped:    make(map[string]string),
		Broadcast: make(map[string]string),
	}
	for k, v := range raw {
		if strings.Contains(k, ".") {
			s.Scoped[k] = v
		} else {
			s.Broadcast[k] = v
		}
	}
	return s
}

// contextVarDefaults maps contextPath -> varName -> default value.
type contextVarDefaults map[string]map[string]string

// resolveScopedOverride matches a dotted --set key against known context
// paths using longest-prefix matching. It returns the context path and var
// name, or empty strings if no match is found.
//
// Example: key "mysql.standalone.imageTag" with known paths
// {"mysql/standalone"} -> contextPath="mysql/standalone", varName="imageTag".
func resolveScopedOverride(dottedKey string, knownPaths map[string]bool) (contextPath, varName string) {
	parts := strings.Split(dottedKey, ".")
	// Try longest prefix first (all but last segment is path, last is var).
	// Then progressively shorten the path prefix.
	for i := len(parts) - 1; i >= 1; i-- {
		candidate := strings.Join(parts[:i], "/")
		if knownPaths[candidate] {
			return candidate, strings.Join(parts[i:], ".")
		}
	}
	return "", ""
}

// computeEffectiveVars builds the final var map for a context by applying
// the priority chain: own defaults < broadcast --set < path-scoped overrides < scoped --set.
//
// Path-scoped overrides (set by a child context targeting this parent) take
// priority over broadcast --set because they express specific intent from
// the composition author. Without this, a broadcast like --set imageTag=X
// would leak into datastores composed alongside a product context that also
// declares imageTag.
func computeEffectiveVars(
	ownDefaults map[string]string,
	pathOverrides map[string]string,
	set SetOverrides,
) map[string]string {
	vars := make(map[string]string, len(ownDefaults))

	for k, v := range ownDefaults {
		vars[k] = v
	}

	for k, v := range set.Broadcast {
		if _, declared := ownDefaults[k]; declared {
			vars[k] = v
		}
	}

	for k, v := range pathOverrides {
		vars[k] = v
	}

	// Scoped --set overrides are applied by the caller after resolving
	// dotted keys to context paths.

	return vars
}
