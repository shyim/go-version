package version

import (
	"fmt"
	"regexp"
	"strings"
)

func normalizeVersion(version string) (string, error) {
	version = strings.TrimSpace(version)
	origVersion := version

	// Strip off aliasing e.g. "1.2.3 as 1.2.3-alias"
	aliasPattern := `^([^,\s]+)\s+as\s+([^,\s]+)$`
	reAlias := regexp.MustCompile(aliasPattern)
	if match := reAlias.FindStringSubmatch(version); match != nil {
		version = match[1]
	}

	// Strip off stability flag e.g. "1.2.3@beta"
	stabilityPattern := fmt.Sprintf("@(?:%s)$", `stable|beta|dev|alpha|RC`)
	reStability := regexp.MustCompile("(?i)" + stabilityPattern)
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
	buildPattern := `^([^,\s+]+)\+[^\s]+$`
	reBuild := regexp.MustCompile(buildPattern)
	if match := reBuild.FindStringSubmatch(version); match != nil {
		version = match[1]
	}

	var matches []string
	modifierIndex := 0 // will indicate where modifiers are found in the matches

	// Match classical versioning like 1.2.3.4 with optional modifiers.
	classicalPattern := fmt.Sprintf(`^v?(\d{1,5})(\.\d+)?(\.\d+)?(\.\d+)?%s$`, `(?:-?(stable|beta|b|alpha|a|RC)(?:[.-]?(\d+))?(?:[.-]?(dev))?)?`)
	reClassical := regexp.MustCompile("(?i)" + classicalPattern)
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
		datePattern := fmt.Sprintf(`^v?(\d{4}(?:[.:-]?\d{2}){1,6}(?:[.:-]?\d{1,3}){0,2})%s$`, `(?:-(stable|beta|b|alpha|a|RC)(?:[.-]?(\d+))?(?:[.-]?(dev))?)?`)
		reDate := regexp.MustCompile("(?i)" + datePattern)
		if matches = reDate.FindStringSubmatch(version); matches != nil {
			// Replace any non-digit character with a dot.
			version = regexp.MustCompile(`\D`).ReplaceAllString(matches[1], ".")
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
				extra = strings.TrimLeft(matches[modifierIndex+1], ".-")
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
	devBranchPattern := `(?i)(.*?)[.-]?dev$`
	reDevBranch := regexp.MustCompile(devBranchPattern)
	if match := reDevBranch.FindStringSubmatch(version); match != nil {
		normalized := normalizeBranch(match[1])
		// A branch ending with "-dev" is only valid if it does not already contain a "dev-" prefix.
		if !strings.Contains(normalized, "dev-") {
			return normalized, nil
		}
	}

	// If no match was found, prepare an appropriate error message.
	extraMessage := ""
	// Check if the alias in fullVersion must be an exact version.
	aliasExactPattern := fmt.Sprintf(` +as +%s(?:@(?:%s))?$`, regexp.QuoteMeta(version), `stable|beta|alpha|RC`)
	reAliasExact := regexp.MustCompile(aliasExactPattern)
	if reAliasExact.MatchString(origVersion) {
		extraMessage = fmt.Sprintf(` in "%s", the alias must be an exact version`, origVersion)
	} else {
		aliasSourcePattern := fmt.Sprintf(`^%s(?:@(?:%s))? +as +`, regexp.QuoteMeta(version), `stable|beta|alpha|RC`)
		reAliasSource := regexp.MustCompile(aliasSourcePattern)
		if reAliasSource.MatchString(origVersion) {
			extraMessage = fmt.Sprintf(` in "%s", the alias source must be an exact version, if it is a branch name you should prefix it with dev-`, origVersion)
		}
	}

	return "", fmt.Errorf(`invalid version string "%s"%s`, origVersion, extraMessage)
}

func normalizeBranch(name string) string {
	name = strings.TrimSpace(name)
	re := regexp.MustCompile(`(?i)^v?(\d+)(\.(\d+|[xX*]))?(\.(\d+|[xX*]))?(\.(\d+|[xX*]))?$`)
	if m := re.FindStringSubmatch(name); m != nil {
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
		return "rc"
	default:
		return s
	}
}
