package version

import "strings"

const (
	StabilityDev    = "dev"
	StabilityAlpha  = "alpha"
	StabilityBeta   = "beta"
	StabilityRC     = "RC"
	StabilityStable = "stable"
)

// Satisfies reports whether versionString matches constraintString.
func Satisfies(versionString, constraintString string) (bool, error) {
	v, err := NewVersion(versionString)
	if err != nil {
		return false, err
	}

	constraintString = strings.TrimSpace(constraintString)
	if constraintString == "" || constraintString == "*" {
		return true, nil
	}

	constraint, err := NewConstraint(constraintString)
	if err != nil {
		return false, err
	}

	return constraint.Check(v), nil
}

// NormalizeComposerVersion normalizes a Composer version string for comparison.
func NormalizeComposerVersion(versionString string) (string, error) {
	return normalizeVersion(versionString)
}

// Stability returns the Composer stability for a version string.
func Stability(versionString string) string {
	v, err := NewVersion(versionString)
	if err != nil {
		normalized, normalizeErr := normalizeVersion(versionString)
		if normalizeErr == nil && strings.HasPrefix(strings.ToLower(normalized), "dev-") {
			return StabilityDev
		}
		return StabilityDev
	}

	stability := getVersionStability(v)
	if stability == "rc" {
		return StabilityRC
	}
	return stability
}
