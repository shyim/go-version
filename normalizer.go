package version

import (
	"fmt"
	"regexp"
	"strings"
)

// Static patterns used by version normalization. Compiling a regexp is far more
// expensive than running it, so these are built once at package load instead of
// on every normalizeVersion call (which runs for every NewVersion).
var (
	reAlias     = regexp.MustCompile(`^([^,\s]+)\s+as\s+([^,\s]+)$`)
	reStability = regexp.MustCompile(`(?i)@(?:stable|beta|dev|alpha|RC)$`)
	reBuild     = regexp.MustCompile(`^([^,\s+]+)\+[^\s]+$`)

	modifierPattern = `(?:[._-]?(stable|beta|b|alpha|a|RC|patch|p|pl|dev)([._-]?\d+(?:[._-][0-9A-Za-z-]+)*)?(?:[._-]?(dev))?)?`
	reClassical     = regexp.MustCompile("(?i)" + fmt.Sprintf(`^v?(\d{1,5})(\.\d+)?(\.\d+)?(\.\d+)?%s$`, modifierPattern))
	reDate          = regexp.MustCompile("(?i)" + fmt.Sprintf(`^v?(\d{4}(?:[.:-]?\d{2}){1,6}(?:[.:-]?\d{1,3}){0,2})%s$`, modifierPattern))
	reNonDigit      = regexp.MustCompile(`\D`)
	reDevBranch     = regexp.MustCompile(`(?i)(.*?)[.-]?dev$`)

	reDevSuffixWildcard = regexp.MustCompile(`(?i)^v?\d+(?:\.(?:\d+|x|\*)){0,3}$`)
	reNumericBranch     = regexp.MustCompile(`(?i)^v?(\d+)(\.(\d+|[xX*]))?(\.(\d+|[xX*]))?(\.(\d+|[xX*]))?$`)
)

func normalizeVersion(version string) (string, error) {
	return normalizeVersionWithContext(version, version)
}

// fastNumericNormalize handles the common case of a plain numeric version
// ("1.2.3", "v1.2", "01.02") without invoking the regexp pipeline. It returns
// ok=false (falling through to the full normalizer) for anything that is not a
// strict 1-4 segment, all-digit, <=5-digit-first-segment version. The guards
// are deliberately conservative so the output is always identical to the
// classical regexp path (verified by TestNormalizeFastPathParity).
func fastNumericNormalize(version string) (string, bool) {
	body := version
	if len(body) > 1 && (body[0] == 'v' || body[0] == 'V') {
		body = body[1:]
	}
	if body == "" {
		return "", false
	}

	segments := make([]string, 0, 4)
	start := 0
	for i := 0; i <= len(body); i++ {
		if i == len(body) || body[i] == '.' {
			seg := body[start:i]
			if seg == "" {
				return "", false // empty segment, e.g. "1..2" or trailing dot
			}
			for j := 0; j < len(seg); j++ {
				if seg[j] < '0' || seg[j] > '9' {
					return "", false // non-digit: let the regexp path handle it
				}
			}
			segments = append(segments, seg)
			if len(segments) > 4 {
				return "", false // too many segments
			}
			start = i + 1
		}
	}

	// The classical regexp caps the first segment at 5 digits; longer numbers
	// are date-style and must not be padded here.
	if len(segments[0]) > 5 {
		return "", false
	}

	for len(segments) < 4 {
		segments = append(segments, "0")
	}
	return strings.Join(segments, "."), true
}

func normalizeVersionWithContext(version, fullVersion string) (string, error) {
	version = strings.TrimSpace(version)
	invalidVersion := version
	fullVersion = strings.TrimSpace(fullVersion)
	if fullVersion == "" {
		fullVersion = invalidVersion
	}

	// Fast path for plain numeric versions like "1.2.3", "v1.2", "01.02".
	// This is a strict subset of the classical regexp path below: it only fires
	// for 1-4 all-ASCII-digit segments with a <=5-digit first segment (so
	// 6+-digit date-style numbers still route through the date heuristic), and
	// it copies the raw digit substrings verbatim so leading zeros are
	// preserved exactly as the regexp path would. Anything else falls through.
	if normalized, ok := fastNumericNormalize(version); ok {
		return normalized, nil
	}

	// Strip off aliasing e.g. "1.2.3 as 1.2.3-alias"
	if match := reAlias.FindStringSubmatch(version); match != nil {
		version = match[1]
	}

	// Strip off stability flag e.g. "1.2.3@beta"
	if match := reStability.FindStringSubmatch(version); match != nil {
		version = version[:len(version)-len(match[0])]
	}

	// Normalize master/trunk/default branches to dev-branch.
	lowerVersion := strings.ToLower(version)
	if lowerVersion == "master" || lowerVersion == "trunk" || lowerVersion == "default" {
		version = "dev-" + version
	}

	// If the requirement is branch-like (starts with dev-), use a normalized branch name.
	if strings.HasPrefix(strings.ToLower(version), "dev-") {
		return "dev-" + version[4:], nil
	}

	// Strip off build metadata: e.g. "1.2.3+buildinfo"
	if match := reBuild.FindStringSubmatch(version); match != nil {
		version = match[1]
	}

	var matches []string
	modifierIndex := 0 // will indicate where modifiers are found in the matches

	// Match classical versioning like 1.2.3.4 with optional modifiers.
	if matches = reClassical.FindStringSubmatch(version); matches != nil {
		major := matches[1]
		minor := ".0"
		patch := ".0"
		build := ".0" // fourth part defaults to .0 if not provided.
		if matches[2] != "" {
			minor = matches[2]
		}
		if matches[3] != "" {
			patch = matches[3]
		}
		if matches[4] != "" {
			build = matches[4]
		}
		version = major + minor + patch + build
		// Modifier (if any) is expected at index 5 in the matches slice.
		modifierIndex = 5
	} else {
		// Match date(time) based versioning such as 2020.01.01 with optional modifiers.
		if matches = reDate.FindStringSubmatch(version); matches != nil {
			// Replace any non-digit character with a dot.
			version = reNonDigit.ReplaceAllString(matches[1], ".")
			// Modifier (if any) is expected at index 2.
			modifierIndex = 2
		}
	}

	// If a version was matched with modifiers, process them.
	if len(matches) > 0 && modifierIndex < len(matches) {
		// If there's a modifier string
		if len(matches) > modifierIndex && matches[modifierIndex] != "" {
			// If the modifier equals "stable", just return the version.
			if strings.ToLower(matches[modifierIndex]) == "stable" {
				return version, nil
			}
			// Append the expanded stability and any extra numeric part.
			version += "-" + expandStability(matches[modifierIndex])
			extra := ""
			if len(matches) > modifierIndex+1 && matches[modifierIndex+1] != "" {
				// Remove any leading dots or hyphens.
				extra = strings.TrimLeft(matches[modifierIndex+1], "._-")
			}
			version += extra
		}

		// If there is an additional modifier (e.g. a dev indicator) append it.
		if len(matches) > modifierIndex+2 && matches[modifierIndex+2] != "" {
			version += "-dev"
		}
		return version, nil
	}

	// Match dev branches such as "feature-dev" or "feature.dev"
	if match := reDevBranch.FindStringSubmatch(version); match != nil {
		base := match[1]
		// Suffix-style arbitrary branches are accepted, but Composer only
		// applies this conversion to simple strings.
		if canConvertDevSuffix(base) {
			return normalizeBranch(base), nil
		}
	}

	// If no match was found, prepare an appropriate error message.
	// These patterns embed the (untrusted) version string, so they are compiled
	// defensively: invalid input (e.g. non-UTF-8 bytes) only costs the richer
	// error message rather than panicking via MustCompile.
	extraMessage := ""
	// Check if the alias in fullVersion must be an exact version.
	aliasExactPattern := fmt.Sprintf(` +as +%s(?:@(?:%s))?$`, regexp.QuoteMeta(version), `stable|beta|dev|alpha|RC`)
	if reAliasExact, err := regexp.Compile(aliasExactPattern); err == nil && reAliasExact.MatchString(fullVersion) {
		extraMessage = fmt.Sprintf(` in "%s", the alias must be an exact version`, fullVersion)
	} else {
		aliasSourcePattern := fmt.Sprintf(`^%s(?:@(?:%s))? +as +`, regexp.QuoteMeta(version), `stable|beta|dev|alpha|RC`)
		if reAliasSource, err := regexp.Compile(aliasSourcePattern); err == nil && reAliasSource.MatchString(fullVersion) {
			extraMessage = fmt.Sprintf(` in "%s", the alias source must be an exact version, if it is a branch name you should prefix it with dev-`, fullVersion)
		}
	}

	return "", fmt.Errorf(`invalid version string "%s"%s`, invalidVersion, extraMessage)
}

func canConvertDevSuffix(base string) bool {
	if strings.ContainsAny(base, " <>!=") || strings.Contains(strings.ToLower(base), "dev-") {
		return false
	}
	if strings.Contains(base, "*") {
		if base == "*" {
			return false
		}
		return reDevSuffixWildcard.MatchString(base)
	}
	return true
}

func normalizeBranch(name string) string {
	name = strings.TrimSpace(name)
	if m := reNumericBranch.FindStringSubmatch(name); m != nil {
		seg1 := m[1]
		seg2 := m[3]
		seg3 := m[5]
		seg4 := m[7]
		if seg2 == "" {
			seg2 = "x"
		}
		if seg3 == "" {
			seg3 = "x"
		}
		if seg4 == "" {
			seg4 = "x"
		}
		seg2 = strings.ReplaceAll(strings.ReplaceAll(seg2, "*", "x"), "X", "x")
		seg3 = strings.ReplaceAll(strings.ReplaceAll(seg3, "*", "x"), "X", "x")
		seg4 = strings.ReplaceAll(strings.ReplaceAll(seg4, "*", "x"), "X", "x")
		normalized := seg1 + "." + seg2 + "." + seg3 + "." + seg4
		normalized = strings.ReplaceAll(normalized, "x", "9999999")
		return normalized + "-dev"
	}
	return "dev-" + name
}

func parseNumericAliasPrefix(input string) (string, bool) {
	input = strings.TrimSpace(input)
	if !strings.HasSuffix(strings.ToLower(input), "-dev") {
		return "", false
	}

	base := input[:len(input)-4]
	if len(base) > 1 && (base[0] == 'v' || base[0] == 'V') {
		base = base[1:]
	}
	if base == "" {
		return "", false
	}

	parts := strings.Split(base, ".")
	if len(parts) > 3 {
		return "", false
	}

	prefix := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return "", false
		}
		if part == "*" || strings.EqualFold(part, "x") {
			break
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return "", false
			}
		}
		prefix = append(prefix, part)
	}

	if len(prefix) == 0 {
		return "", false
	}
	return strings.Join(prefix, ".") + ".", true
}

func expandStability(stability string) string {
	s := strings.ToLower(stability)
	switch s {
	case "a":
		return "alpha"
	case "b":
		return "beta"
	case "p", "pl":
		return "patch"
	case "rc":
		return "RC"
	default:
		return s
	}
}

func normalizeStability(stability string) string {
	return expandStability(stability)
}
