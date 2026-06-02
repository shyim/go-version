package version

import (
	"strconv"
	"testing"
)

// regexpOnlyNormalize reproduces the classical-versioning branch of the full
// normalizer without the fast path, serving as the behavioral oracle for
// fastNumericNormalize. It mirrors normalizer.go's reClassical handling: 1-4
// digit segments (first <=5 digits), padded to four with ".0", leading zeros
// preserved verbatim.
func regexpOnlyNormalize(t *testing.T, version string) (string, bool) {
	t.Helper()
	matches := reClassical.FindStringSubmatch(version)
	if matches == nil {
		return "", false
	}
	// Reject if a stability modifier was captured; the fast path only covers
	// pure numeric versions.
	if matches[5] != "" || matches[6] != "" || matches[7] != "" {
		return "", false
	}
	major := matches[1]
	minor, patch, build := ".0", ".0", ".0"
	if matches[2] != "" {
		minor = matches[2]
	}
	if matches[3] != "" {
		patch = matches[3]
	}
	if matches[4] != "" {
		build = matches[4]
	}
	return major + minor + patch + build, true
}

// TestNormalizeFastPathParity is the safety net for the numeric fast path: for
// every generated input, whenever fastNumericNormalize fires it must produce the
// exact same string as the classical regexp path, and the full normalizeVersion
// result must match too. It also checks explicit edge cases.
func TestNormalizeFastPathParity(t *testing.T) {
	var inputs []string
	// Explicit edge cases called out in the analysis.
	inputs = append(inputs,
		"1", "1.2", "1.2.3", "1.2.3.4",
		"v1", "v1.2", "V2.0.0",
		"01", "01.02", "01.02.03", "00.00", "000000",
		"99999", "99999.1", "100000", "100000.0.0", // 6-digit first segment must fall through
		"1.2.3.4.5", // 5 segments: must fall through (to an error)
		"1..2", "1.", ".1", "", "v",
		"1.2.3-beta", "1.2.3@stable", "1.2.3+build", "dev-main", // non-numeric: must fall through
	)
	// Generated numeric combinations.
	for a := 0; a < 12; a++ {
		for b := 0; b < 6; b++ {
			inputs = append(inputs,
				strconv.Itoa(a),
				strconv.Itoa(a)+"."+strconv.Itoa(b),
				strconv.Itoa(a)+"."+strconv.Itoa(b)+".0",
				"v"+strconv.Itoa(a)+"."+strconv.Itoa(b),
				"0"+strconv.Itoa(a)+"."+strconv.Itoa(b), // leading zero
			)
		}
	}

	for _, in := range inputs {
		fast, fastOK := fastNumericNormalize(trimmed(in))

		// 1) When the fast path fires, it must equal the regexp classical path.
		if fastOK {
			ref, refOK := regexpOnlyNormalize(t, trimmed(in))
			if !refOK {
				t.Errorf("fast path fired for %q -> %q but regexp classical path rejected it", in, fast)
				continue
			}
			if fast != ref {
				t.Errorf("fast path %q -> %q, regexp path -> %q", in, fast, ref)
			}
		}

		// 2) The full normalizeVersion result must be consistent regardless of
		//    which path produced it: if the fast path fired, the public result
		//    must equal the fast output.
		got, err := normalizeVersion(in)
		if fastOK {
			if err != nil {
				t.Errorf("normalizeVersion(%q) errored %v but fast path produced %q", in, err, fast)
			} else if got != fast {
				t.Errorf("normalizeVersion(%q) = %q, fast path = %q", in, got, fast)
			}
		}
	}
}

func trimmed(s string) string {
	// normalizeVersionWithContext trims before calling the fast path, so the
	// test must too.
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
