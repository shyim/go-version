package version

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

// Some fixtures are derived from an external MIT-licensed semver test suite.
// See EXTERNAL_TESTS_LICENSE.md for the source notice.
func TestSemverExtendedSatisfies(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		constraint string
		expected   bool
	}{
		{name: "x wildcard major", version: "2.1.3", constraint: "2.x.x", expected: true},
		{name: "x wildcard minor", version: "1.2.3", constraint: "1.2.x", expected: true},
		{name: "x wildcard disjunction right", version: "2.1.3", constraint: "1.2.x || 2.x", expected: true},
		{name: "x wildcard disjunction left", version: "1.2.3", constraint: "1.2.x || 2.x", expected: true},
		{name: "x wildcard any", version: "1.2.3", constraint: "x", expected: true},
		{name: "repeated star wildcard", version: "2.1.3", constraint: "2.*.*", expected: true},
		{name: "segment star wildcard any", version: "1.2.3", constraint: "*.*", expected: true},
		{name: "x wildcard wrong major", version: "1.1.3", constraint: "2.x.x", expected: false},
		{name: "x wildcard wrong minor", version: "1.3.3", constraint: "1.2.x", expected: false},
		{name: "repeated star wildcard wrong major", version: "1.1.3", constraint: "2.*.*", expected: false},
		{name: "dev branch exact", version: "dev-master", constraint: "dev-master", expected: true},
		{name: "dev branch arbitrary exact", version: "dev-feature-a", constraint: "dev-feature-a", expected: true},
		{name: "numeric branch constraint", version: "1.0.9999999.9999999-dev", constraint: "1.0.x-dev", expected: true},
		{name: "less-than excludes implicit stable prerelease", version: "1.0.0beta", constraint: "<1", expected: false},
		{name: "less-than excludes prerelease of bound", version: "1.2.3-beta", constraint: "<1.2.3", expected: false},
		{name: "tilde with v prefix and prerelease", version: "0.5.4-alpha", constraint: "~v0.5.4-beta", expected: false},
		{name: "hyphen v major", version: "2.5.0", constraint: "v1 - v2", expected: true},
		{name: "hyphen prerelease lower partial upper", version: "2.3.9", constraint: "1.2-beta - 2.3", expected: true},
		{name: "hyphen prerelease lower partial upper stop", version: "2.4.0", constraint: "1.2-beta - 2.3", expected: false},
		{name: "caret zero minor includes patch", version: "0.2.1", constraint: "^0.2.0", expected: true},
		{name: "caret zero minor excludes next minor", version: "0.3.0", constraint: "^0.2.0", expected: false},
		{name: "caret zero zero excludes next patch", version: "0.0.4", constraint: "^0.0.3", expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := checkSatisfiesForTest(tc.version, tc.constraint)
			if err != nil {
				t.Fatalf("unexpected parse error for version %q constraint %q: %v", tc.version, tc.constraint, err)
			}
			if actual != tc.expected {
				t.Fatalf("constraint %q with version %q: expected %v, got %v", tc.constraint, tc.version, tc.expected, actual)
			}
		})
	}
}

func TestSemverSatisfiesCases(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		expected   bool
	}{
		{"1.2.3", "1.0.0 - 2.0.0", true},
		{"1.2.3", "^1.2.3+build", true},
		{"1.3.0", "^1.2.3+build", true},
		{"2.4.3-alpha", "1.2.3+asdf - 2.4.3+asdf", true},
		{"1.3.0-beta", ">1.2", true},
		{"1.2.3-beta", "<=1.2.3", true},
		{"1.2.3-beta", "^1.2.3", true},
		{"1.2.3", "1.2.3+asdf - 2.4.3+asdf", true},
		{"1.0.0", "1.0.0", true},
		{"1.2.3", "*", true},
		{"v1.2.3", "*", true},
		{"1.0.0", ">=1.0.0", true},
		{"1.0.1", ">=1.0.0", true},
		{"1.1.0", ">=1.0.0", true},
		{"1.0.1", ">1.0.0", true},
		{"1.1.0", ">1.0.0", true},
		{"2.0.0", "<=2.0.0", true},
		{"1.9999.9999", "<=2.0.0", true},
		{"0.2.9", "<=2.0.0", true},
		{"1.9999.9999", "<2.0.0", true},
		{"0.2.9", "<2.0.0", true},
		{"1.0.0", ">= 1.0.0", true},
		{"1.0.1", ">=  1.0.0", true},
		{"1.1.0", ">=   1.0.0", true},
		{"1.0.1", "> 1.0.0", true},
		{"1.1.0", ">  1.0.0", true},
		{"2.0.0", "<=   2.0.0", true},
		{"1.9999.9999", "<= 2.0.0", true},
		{"0.2.9", "<=  2.0.0", true},
		{"1.9999.9999", "<    2.0.0", true},
		{"0.2.9", "<\t2.0.0", true},
		{"v0.1.97", ">=0.1.97", true},
		{"0.1.97", ">=0.1.97", true},
		{"1.2.4", "0.1.20 || 1.2.4", true},
		{"0.0.0", ">=0.2.3 || <0.0.1", true},
		{"0.2.3", ">=0.2.3 || <0.0.1", true},
		{"0.2.4", ">=0.2.3 || <0.0.1", true},
		{"2.1.3", "2.x.x", true},
		{"1.2.3", "1.2.x", true},
		{"2.1.3", "1.2.x || 2.x", true},
		{"1.2.3", "1.2.x || 2.x", true},
		{"1.2.3", "x", true},
		{"2.1.3", "2.*.*", true},
		{"1.2.3", "1.2.*", true},
		{"2.1.3", "1.2.* || 2.*", true},
		{"1.2.3", "1.2.* || 2.*", true},
		{"1.2.3", "*", true},
		{"2.9.0", "~2.4", true},
		{"2.4.5", "~2.4", true},
		{"1.2.3", "~1", true},
		{"1.4.7", "~1.0", true},
		{"1.0.0", ">=1", true},
		{"1.0.0", ">= 1", true},
		{"1.2.8", ">1.2", true},
		{"1.1.1", "<1.2", true},
		{"1.1.1", "< 1.2", true},
		{"1.2.3", "~1.2.1 >=1.2.3", true},
		{"1.2.3", "~1.2.1 =1.2.3", true},
		{"1.2.3", "~1.2.1 1.2.3", true},
		{"1.2.3", "~1.2.1 >=1.2.3 1.2.3", true},
		{"1.2.3", "~1.2.1 1.2.3 >=1.2.3", true},
		{"1.2.3", "~1.2.1 1.2.3", true},
		{"1.2.3", ">=1.2.1 1.2.3", true},
		{"1.2.3", "1.2.3 >=1.2.1", true},
		{"1.2.3", ">=1.2.3 >=1.2.1", true},
		{"1.2.3", ">=1.2.1 >=1.2.3", true},
		{"1.2.8", ">=1.2", true},
		{"1.8.1", "^1.2.3", true},
		{"0.1.2", "^0.1.2", true},
		{"0.1.2", "^0.1", true},
		{"1.4.2", "^1.2", true},
		{"1.4.2", "^1.2 ^1", true},
		{"0.0.1-beta", "^0.0.1-alpha", true},
		{"2.2.3", "1.0.0 - 2.0.0", false},
		{"2.0.0", "^1.2.3+build", false},
		{"1.2.0", "^1.2.3+build", false},
		{"1.0.0beta", "1", false},
		{"1.0.0beta", "<1", false},
		{"1.0.0beta", "< 1", false},
		{"1.0.1", "1.0.0", false},
		{"0.0.0", ">=1.0.0", false},
		{"0.0.1", ">=1.0.0", false},
		{"0.1.0", ">=1.0.0", false},
		{"0.0.1", ">1.0.0", false},
		{"0.1.0", ">1.0.0", false},
		{"3.0.0", "<=2.0.0", false},
		{"2.9999.9999", "<=2.0.0", false},
		{"2.2.9", "<=2.0.0", false},
		{"2.9999.9999", "<2.0.0", false},
		{"2.2.9", "<2.0.0", false},
		{"v0.1.93", ">=0.1.97", false},
		{"0.1.93", ">=0.1.97", false},
		{"1.2.3", "0.1.20 || 1.2.4", false},
		{"0.0.3", ">=0.2.3 || <0.0.1", false},
		{"0.2.2", ">=0.2.3 || <0.0.1", false},
		{"1.1.3", "2.x.x", false},
		{"3.1.3", "2.x.x", false},
		{"1.3.3", "1.2.x", false},
		{"3.1.3", "1.2.x || 2.x", false},
		{"1.1.3", "1.2.x || 2.x", false},
		{"1.1.3", "2.*.*", false},
		{"3.1.3", "2.*.*", false},
		{"1.3.3", "1.2.*", false},
		{"3.1.3", "1.2.* || 2.*", false},
		{"1.1.3", "1.2.* || 2.*", false},
		{"1.1.2", "2", false},
		{"2.4.1", "2.3", false},
		{"3.0.0", "~2.4", false},
		{"2.3.9", "~2.4", false},
		{"0.2.3", "~1", false},
		{"1.0.0", "<1", false},
		{"1.1.1", ">=1.2", false},
		{"2.0.0beta", "1", false},
		{"0.5.4-alpha", "~v0.5.4-beta", false},
		{"1.2.3-beta", "<1.2.3", false},
		{"2.0.0-alpha", "^1.2.3", false},
		{"1.2.2", "^1.2.3", false},
		{"1.1.9", "^1.2", false},
	}

	for _, tc := range tests {
		actual, err := checkSatisfiesForTest(tc.version, tc.constraint)
		if err != nil {
			t.Errorf("Satisfies(%q, %q) unexpected error: %v", tc.version, tc.constraint, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("Satisfies(%q, %q): expected %v, got %v", tc.version, tc.constraint, tc.expected, actual)
		}
	}
}

func TestSemverNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "none", input: "1.0.0", expected: "1.0.0.0"},
		{name: "none/2", input: "1.2.3.4", expected: "1.2.3.4"},
		{name: "parses state", input: "1.0.0RC1dev", expected: "1.0.0.0-RC1-dev"},
		{name: "rc uppercase", input: "1.0.0-rc1", expected: "1.0.0.0-RC1"},
		{name: "rc case insensitive", input: "1.0.0-rC15-dev", expected: "1.0.0.0-RC15-dev"},
		{name: "rc dotted delimiters", input: "1.0.0.RC.15-dev", expected: "1.0.0.0-RC15-dev"},
		{name: "patch shorthand", input: "1.0.0.pl3-dev", expected: "1.0.0.0-patch3-dev"},
		{name: "forces w.x.y.z", input: "1.0-dev", expected: "1.0.0.0-dev"},
		{name: "forces w.x.y.z/2", input: "0", expected: "0.0.0.0"},
		{name: "forces w.x.y.z/maximum major", input: "99999", expected: "99999.0.0.0"},
		{name: "parses long", input: "10.4.13-beta", expected: "10.4.13.0-beta"},
		{name: "parses long/2", input: "10.4.13beta2", expected: "10.4.13.0-beta2"},
		{name: "parses long/semver", input: "10.4.13beta.2", expected: "10.4.13.0-beta2"},
		{name: "parses long/semver2", input: "v1.13.11-beta.0", expected: "1.13.11.0-beta0"},
		{name: "parses long/semver3", input: "1.13.11.0-beta0", expected: "1.13.11.0-beta0"},
		{name: "expand shorthand", input: "10.4.13-b", expected: "10.4.13.0-beta"},
		{name: "expand shorthand/2", input: "10.4.13-b5", expected: "10.4.13.0-beta5"},
		{name: "strips leading v", input: "v1.0.0", expected: "1.0.0.0"},
		{name: "parses dates y-m as classical", input: "2010.01", expected: "2010.01.0.0"},
		{name: "parses dates w/ . as classical", input: "2010.01.02", expected: "2010.01.02.0"},
		{name: "parses dates y.m.Y as classical", input: "2010.1.555", expected: "2010.1.555.0"},
		{name: "parses dates y.m.Y/2 as classical", input: "2010.10.200", expected: "2010.10.200.0"},
		{name: "parses CalVer YYYYMMDD (as MAJOR) versions", input: "20230131.0.0", expected: "20230131.0.0"},
		{name: "parses CalVer YYYYMMDDhhmm (as MAJOR) versions", input: "202301310000.0.0", expected: "202301310000.0.0"},
		{name: "strips v/datetime", input: "v20100102", expected: "20100102"},
		{name: "parses dates no delimiter", input: "20100102", expected: "20100102"},
		{name: "parses dates no delimiter/2", input: "20100102.0", expected: "20100102.0"},
		{name: "parses dates no delimiter/3", input: "20100102.1.0", expected: "20100102.1.0"},
		{name: "parses dates no delimiter/4", input: "20100102.0.3", expected: "20100102.0.3"},
		{name: "parses dates no delimiter/earliest year", input: "100000", expected: "100000"},
		{name: "parses dates w/ - and .", input: "2010-01-02-10-20-30.0.3", expected: "2010.01.02.10.20.30.0.3"},
		{name: "parses dates w/ - and ./2", input: "2010-01-02-10-20-30.5", expected: "2010.01.02.10.20.30.5"},
		{name: "parses dates w/ -", input: "2010-01-02", expected: "2010.01.02"},
		{name: "parses dates w/ .", input: "2012.06.07", expected: "2012.06.07.0"},
		{name: "parses numbers", input: "2010-01-02.5", expected: "2010.01.02.5"},
		{name: "parses dates y.m.Y", input: "2010.1.555", expected: "2010.1.555.0"},
		{name: "parses datetime", input: "20100102-203040", expected: "20100102.203040"},
		{name: "parses date dev", input: "20100102.x-dev", expected: "20100102.9999999.9999999.9999999-dev"},
		{name: "parses datetime dev", input: "20100102.203040.x-dev", expected: "20100102.203040.9999999.9999999-dev"},
		{name: "parses dt+number", input: "20100102203040-10", expected: "20100102203040.10"},
		{name: "parses dt+patch", input: "20100102-203040-p1", expected: "20100102.203040-patch1"},
		{name: "parses dt Ym", input: "201903.0", expected: "201903.0"},
		{name: "parses dt Ym dev", input: "201903.x-dev", expected: "201903.9999999.9999999.9999999-dev"},
		{name: "parses dt Ym+patch", input: "201903.0-p2", expected: "201903.0-patch2"},
		{name: "parses master", input: "dev-master", expected: "dev-master"},
		{name: "parses master w/o dev", input: "master", expected: "dev-master"},
		{name: "parses trunk", input: "dev-trunk", expected: "dev-trunk"},
		{name: "parses branches", input: "1.x-dev", expected: "1.9999999.9999999.9999999-dev"},
		{name: "parses arbitrary", input: "dev-feature-foo", expected: "dev-feature-foo"},
		{name: "parses arbitrary/2", input: "DEV-FOOBAR", expected: "dev-FOOBAR"},
		{name: "parses arbitrary/3", input: "dev-feature/foo", expected: "dev-feature/foo"},
		{name: "parses arbitrary/4", input: "dev-feature+issue-1", expected: "dev-feature+issue-1"},
		{name: "ignores aliases", input: "dev-master as 1.0.0", expected: "dev-master"},
		{name: "ignores aliases/2", input: "dev-load-varnish-only-when-used as ^2.0", expected: "dev-load-varnish-only-when-used"},
		{name: "ignores aliases/3", input: "dev-load-varnish-only-when-used@dev as ^2.0@dev", expected: "dev-load-varnish-only-when-used"},
		{name: "ignores stability", input: "1.0.0+foo@dev", expected: "1.0.0.0"},
		{name: "ignores stability/2", input: "dev-load-varnish-only-when-used@stable", expected: "dev-load-varnish-only-when-used"},
		{name: "semver beta metadata", input: "1.0.0-beta.5+foo", expected: "1.0.0.0-beta5"},
		{name: "semver metadata/3", input: "1.0.0+foo", expected: "1.0.0.0"},
		{name: "semver alpha dotted metadata", input: "1.0.0-alpha.3.1+foo", expected: "1.0.0.0-alpha3.1"},
		{name: "semver metadata/5", input: "1.0.0-alpha2.1+foo", expected: "1.0.0.0-alpha2.1"},
		{name: "semver alpha dashed metadata", input: "1.0.0-alpha-2.1-3+foo", expected: "1.0.0.0-alpha2.1-3"},
		{name: "metadata w/ alias", input: "1.0.0+foo as 2.0", expected: "1.0.0.0"},
		{name: "keep zero-padding", input: "00.01.03.04", expected: "00.01.03.04"},
		{name: "keep zero-padding/2", input: "000.001.003.004", expected: "000.001.003.004"},
		{name: "keep zero-padding/3", input: "0.000.103.204", expected: "0.000.103.204"},
		{name: "keep zero-padding/4", input: "0700", expected: "0700.0.0.0"},
		{name: "keep zero-padding/5", input: "041.x-dev", expected: "041.9999999.9999999.9999999-dev"},
		{name: "keep zero-padding/6", input: "dev-041.003", expected: "dev-041.003"},
		{name: "dev with mad name", input: "dev-1.0.0-dev<1.0.5-dev", expected: "dev-1.0.0-dev<1.0.5-dev"},
		{name: "dev prefix with spaces", input: "dev-foo bar", expected: "dev-foo bar"},
		{name: "space padding", input: " 1.0.0", expected: "1.0.0.0"},
		{name: "space padding/2", input: "1.0.0 ", expected: "1.0.0.0"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := normalizeVersion(tc.input)
			if err != nil {
				t.Fatalf("unexpected normalize error for %q: %v", tc.input, err)
			}
			if actual != tc.expected {
				t.Fatalf("normalize %q: expected %q, got %q", tc.input, tc.expected, actual)
			}
		})
	}
}

func TestSemverNormalizeFailsCases(t *testing.T) {
	tests := []string{
		"",
		"a",
		"1.0.0-meh",
		"1.0.0.0.0",
		"feature-foo",
		"1.0.0+foo bar",
		"1.0.1-SNAPSHOT",
		"1.0.0<1.0.5-dev",
		"1.0.0-dev<1.0.5-dev",
		"foo bar-dev",
		"1.0 .2",
		" as ",
		" as 1.2",
		"^",
		"^8 || ^",
		"~",
		"~1 ~",
		"~1",
		"^1",
		"1.*",
		"20100102.0.3.4",
		"100000.0.0.0",
		"2023013.0.0",
		"202301311.0.0",
		"20230131000.0.0",
		"2023013100000.0.0",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if actual, err := normalizeVersion(input); err == nil {
				t.Fatalf("expected normalizeVersion(%q) to fail, got %q", input, actual)
			}
		})
	}
}

func TestSemverNormalizeAliasErrorCases(t *testing.T) {
	badAliases := []struct {
		full   string
		source string
		alias  string
	}{
		{"1.0.0+foo as ^2.0", "1.0.0+foo", "^2.0"},
		{"1.0.0+foo as  ~2.0", "1.0.0+foo", "~2.0"},
		{"1.0.0+foo  as >2.0", "1.0.0+foo", ">2.0"},
		{"1.0.0+foo as <2.0", "1.0.0+foo", "<2.0"},
		{"1.0.0+foo@dev as <2.0@dev", "1.0.0+foo@dev", "<2.0@dev"},
	}

	for _, tc := range badAliases {
		if _, err := normalizeVersionWithContext(tc.source, tc.full); err != nil {
			t.Errorf("normalizeVersionWithContext(%q, %q) unexpected source error: %v", tc.source, tc.full, err)
			continue
		}

		_, err := normalizeVersionWithContext(tc.alias, tc.full)
		expected := fmt.Sprintf(`invalid version string "%s" in "%s", the alias must be an exact version`, tc.alias, tc.full)
		if err == nil || err.Error() != expected {
			t.Errorf("normalizeVersionWithContext(%q, %q): expected error %q, got %v", tc.alias, tc.full, expected, err)
		}
	}

	badAliasees := []struct {
		full   string
		source string
		alias  string
	}{
		{"^2.0 as 1.0.0+foo", "^2.0", "1.0.0+foo"},
		{"~2.0 as  1.0.0+foo", "~2.0", "1.0.0+foo"},
		{">2.0  as 1.0.0+foo", ">2.0", "1.0.0+foo"},
		{"<2.0 as 1.0.0+foo", "<2.0", "1.0.0+foo"},
		{"<2.0@dev as 1.2.3@dev", "<2.0@dev", "1.2.3@dev"},
	}

	for _, tc := range badAliasees {
		_, err := normalizeVersionWithContext(tc.source, tc.full)
		expected := fmt.Sprintf(`invalid version string "%s" in "%s", the alias source must be an exact version, if it is a branch name you should prefix it with dev-`, tc.source, tc.full)
		if err == nil || err.Error() != expected {
			t.Errorf("normalizeVersionWithContext(%q, %q): expected error %q, got %v", tc.source, tc.full, expected, err)
		}

		if _, err := normalizeVersionWithContext(tc.alias, tc.full); err != nil {
			t.Errorf("normalizeVersionWithContext(%q, %q) unexpected alias error: %v", tc.alias, tc.full, err)
		}
	}
}

func TestSemverNormalizeBranchCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.x", "1.9999999.9999999.9999999-dev"},
		{"v1.*", "1.9999999.9999999.9999999-dev"},
		{"v1.0", "1.0.9999999.9999999-dev"},
		{"2.0", "2.0.9999999.9999999-dev"},
		{"v1.0.x", "1.0.9999999.9999999-dev"},
		{"v1.0.3.*", "1.0.3.9999999-dev"},
		{"v2.4.0", "2.4.0.9999999-dev"},
		{"2.4.4", "2.4.4.9999999-dev"},
		{"master", "dev-master"},
		{"trunk", "dev-trunk"},
		{"feature-a", "dev-feature-a"},
		{"FOOBAR", "dev-FOOBAR"},
		{"feature+issue-1", "dev-feature+issue-1"},
	}

	for _, tc := range tests {
		if actual := normalizeBranch(tc.input); actual != tc.expected {
			t.Errorf("normalizeBranch(%q): expected %q, got %q", tc.input, tc.expected, actual)
		}
	}
}

func TestSemverNumericAliasPrefixCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		ok       bool
	}{
		{"0.x-dev", "0.", true},
		{"1.0.x-dev", "1.0.", true},
		{"1.x-dev", "1.", true},
		{"1.2.x-dev", "1.2.", true},
		{"1.2-dev", "1.2.", true},
		{"1-dev", "1.", true},
		{"dev-develop", "", false},
		{"dev-master", "", false},
	}

	for _, tc := range tests {
		actual, ok := parseNumericAliasPrefix(tc.input)
		if actual != tc.expected || ok != tc.ok {
			t.Errorf("parseNumericAliasPrefix(%q): expected (%q, %v), got (%q, %v)", tc.input, tc.expected, tc.ok, actual, ok)
		}
	}
}

func TestSemverIsValidCases(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"0.x-dev", true},
		{"dev-develop", true},
		{"1.0.2", true},
		{"1.0.2.5", true},
		{"1.0.2.5.5", false},
		{"foo", false},
	}

	for _, tc := range tests {
		_, err := NewVersion(tc.input)
		if actual := err == nil; actual != tc.expected {
			t.Errorf("isValid(%q): expected %v, got %v (err: %v)", tc.input, tc.expected, actual, err)
		}
	}
}

func TestSemverComparatorCases(t *testing.T) {
	tests := []struct {
		version1 string
		operator string
		version2 string
		expected bool
	}{
		{"1.25.0", ">", "1.24.0", true},
		{"1.25.0", ">", "1.25.0", false},
		{"1.25.0", ">", "1.26.0", false},
		{"1.26.0", ">", "dev-foo", true},
		{"dev-foo", ">", "dev-master", false},
		{"dev-foo", ">", "dev-bar", false},
		{"1.25.0", ">=", "1.24.0", true},
		{"1.25.0", ">=", "1.25.0", true},
		{"1.25.0", ">=", "1.26.0", false},
		{"1.25.0", "<", "1.24.0", false},
		{"1.25.0", "<", "1.25.0", false},
		{"1.25.0", "<", "1.26.0", true},
		{"1.0.0", "<", "1.2-dev", true},
		{"dev-foo", "<", "1.26.0", true},
		{"dev-foo", "<", "dev-master", false},
		{"dev-foo", "<", "dev-bar", false},
		{"1.25.0", "<=", "1.24.0", false},
		{"1.25.0", "<=", "1.25.0", true},
		{"1.25.0", "<=", "1.26.0", true},
		{"1.25.0", "==", "1.24.0", false},
		{"1.25.0", "==", "1.25.0", true},
		{"1.25.0", "==", "1.26.0", false},
		{"dev-foo", "==", "1.26.0", false},
		{"dev-foo", "==", "dev-master", false},
		{"dev-foo", "==", "dev-bar", false},
		{"1.25.0", "!=", "1.24.0", true},
		{"1.25.0", "!=", "1.25.0", false},
		{"1.25.0", "!=", "1.26.0", true},
		{"1.25.0-beta2.1", "<", "1.25.0-b.3", true},
		{"1.25.0-b2.1", "<", "1.25.0beta.3", true},
		{"1.25.0-b-2.1", "<", "1.25.0-rc", true},
		{"1.25.0-beta2.1", "==", "1.25.0-b.2.1", true},
		{"1.25.0beta2.1", "==", "1.25.0-b2.1", true},
		{"1.25.0", "=", "1.24.0", false},
		{"1.25.0", "=", "1.25.0", true},
		{"1.25.0", "=", "1.26.0", false},
		{"1.25.0", "<>", "1.24.0", true},
		{"1.25.0", "<>", "1.25.0", false},
		{"1.25.0", "<>", "1.26.0", true},
	}

	for _, tc := range tests {
		actual, err := compareForTest(tc.version1, tc.operator, tc.version2)
		if err != nil {
			t.Errorf("compare(%q, %q, %q) unexpected error: %v", tc.version1, tc.operator, tc.version2, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("compare(%q, %q, %q): expected %v, got %v", tc.version1, tc.operator, tc.version2, tc.expected, actual)
		}
	}
}

func TestConstraintInvalidComparatorOperator(t *testing.T) {
	if _, err := compareForTest("1.1", "!==", "1.2"); err == nil {
		t.Fatal("expected invalid comparator operator to be rejected")
	}
}

func TestSemverSortCases(t *testing.T) {
	tests := []struct {
		versions []string
		sorted   []string
		rsorted  []string
	}{
		{
			[]string{"1.0", "0.1", "0.1", "3.2.1", "2.4.0-alpha", "2.4.0"},
			[]string{"0.1", "0.1", "1.0", "2.4.0-alpha", "2.4.0", "3.2.1"},
			[]string{"3.2.1", "2.4.0", "2.4.0-alpha", "1.0", "0.1", "0.1"},
		},
		{
			[]string{"dev-foo", "dev-master", "1.0", "50.2"},
			[]string{"dev-foo", "1.0", "50.2", "dev-master"},
			[]string{"dev-master", "50.2", "1.0", "dev-foo"},
		},
	}

	for _, tc := range tests {
		versions := mustVersions(t, tc.versions)
		sort.Sort(Collection(versions))
		if actual := versionOriginals(versions); !equalStrings(actual, tc.sorted) {
			t.Errorf("sort(%v): expected %v, got %v", tc.versions, tc.sorted, actual)
		}

		versions = mustVersions(t, tc.versions)
		sort.Sort(sort.Reverse(Collection(versions)))
		if actual := versionOriginals(versions); !equalStrings(actual, tc.rsorted) {
			t.Errorf("rsort(%v): expected %v, got %v", tc.versions, tc.rsorted, actual)
		}
	}
}

func TestSemverSatisfiedByCases(t *testing.T) {
	tests := []struct {
		constraint string
		versions   []string
		expected   []string
	}{
		{
			"~1.0",
			[]string{"1.0", "1.2", "1.9999.9999", "2.0", "2.1", "0.9999.9999"},
			[]string{"1.0", "1.2", "1.9999.9999"},
		},
		{
			">1.0 <3.0 || >=4.0",
			[]string{"1.0", "1.1", "2.9999.9999", "3.0", "3.1", "3.9999.9999", "4.0", "4.1"},
			[]string{"1.1", "2.9999.9999", "4.0", "4.1"},
		},
		{
			"^0.2.0",
			[]string{"0.1.1", "0.1.9999", "0.2.0", "0.2.1", "0.3.0"},
			[]string{"0.2.0", "0.2.1"},
		},
	}

	for _, tc := range tests {
		actual, err := satisfiedByForTest(tc.constraint, tc.versions)
		if err != nil {
			t.Errorf("satisfiedBy(%q, %v) unexpected error: %v", tc.constraint, tc.versions, err)
			continue
		}
		if !equalStrings(actual, tc.expected) {
			t.Errorf("satisfiedBy(%q, %v): expected %v, got %v", tc.constraint, tc.versions, tc.expected, actual)
		}
	}
}

func TestSemverStabilityCases(t *testing.T) {
	tests := []struct {
		expected string
		version  string
	}{
		{"stable", "1"},
		{"stable", "1.0"},
		{"stable", "3.2.1"},
		{"stable", "v3.2.1"},
		{"dev", "v2.0.x-dev"},
		{"dev", "v2.0.x-dev#abc123"},
		{"dev", "v2.0.x-dev#trunk/@123"},
		{"RC", "3.0-RC2"},
		{"dev", "dev-master"},
		{"dev", "3.1.2-dev"},
		{"dev", "dev-feature+issue-1"},
		{"stable", "3.1.2-p1"},
		{"stable", "3.1.2-pl2"},
		{"stable", "3.1.2-patch"},
		{"alpha", "3.1.2-alpha5"},
		{"beta", "3.1.2-beta"},
		{"beta", "2.0B1"},
		{"alpha", "1.2.0a1"},
		{"alpha", "1.2_a1"},
		{"RC", "2.0.0rc1"},
		{"alpha", "1.0.0-alpha11+cs-1.1.0"},
		{"dev", "1-2_dev"},
	}

	for _, tc := range tests {
		if actual := Stability(tc.version); actual != tc.expected {
			t.Errorf("Stability(%q): expected %q, got %q", tc.version, tc.expected, actual)
		}
	}
}

func TestSemverNormalizeStability(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"rc", "RC"},
		{"BeTa", "beta"},
	}

	for _, tc := range tests {
		if actual := normalizeStability(tc.input); actual != tc.expected {
			t.Errorf("normalizeStability(%q): expected %q, got %q", tc.input, tc.expected, actual)
		}
	}
}

func TestSemverRangeCases(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		expected   bool
	}{
		{"2.5.0", "v2.*", true},
		{"3.0.0", "v2.*", false},
		{"2.5.0", "2.*.*", true},
		{"3.0.0", "2.*.*", false},
		{"20.5.0", "20.*", true},
		{"21.0.0", "20.*", false},
		{"20.5.0", "20.*.*", true},
		{"21.0.0", "20.*.*", false},
		{"2.0.5", "2.0.*", true},
		{"2.1.0", "2.0.*", false},
		{"2.5.0", "2.x", true},
		{"3.0.0", "2.x", false},
		{"2.5.0", "2.x.x", true},
		{"3.0.0", "2.x.x", false},
		{"2.2.9", "2.2.x", true},
		{"2.3.0", "2.2.x", false},
		{"2.10.9", "2.10.X", true},
		{"2.11.0", "2.10.X", false},
		{"2.1.3.5", "2.1.3.*", true},
		{"2.1.4.0", "2.1.3.*", false},
		{"0.5.0", "0.*", true},
		{"1.0.0", "0.*", false},
		{"0.5.0", "0.*.*", true},
		{"1.0.0", "0.*.*", false},
		{"0.5.0", "0.x", true},
		{"1.0.0", "0.x", false},
		{"0.5.0", "0.x.x", true},
		{"1.0.0", "0.x.x", false},
		{"1.5.0", "~v1", true},
		{"2.0.0", "~v1", false},
		{"1.5.0", "~1.0", true},
		{"2.0.0", "~1.0", false},
		{"1.0.9", "~1.0.0", true},
		{"1.1.0", "~1.0.0", false},
		{"1.9.9", "~1.2", true},
		{"2.0.0", "~1.2", false},
		{"1.2.9", "~1.2.3", true},
		{"1.3.0", "~1.2.3", false},
		{"1.2.3.5", "~1.2.3.4", true},
		{"1.2.4.0", "~1.2.3.4", false},
		{"1.9.9", "~1.2-beta", true},
		{"2.0.0", "~1.2-beta", false},
		{"1.2.0-beta2", "~1.2-BETA2", true},
		{"1.2.0-beta1", "~1.2-BETA2", false},
		{"1.2.2-dev", "~1.2.2-dev", true},
		{"1.3.0-dev", "~1.2.2-dev", false},
		{"1.2.2", "~1.2.2-stable", true},
		{"1.2.2-RC1", "~1.2.2-stable", false},
		{"201903.10", "~201903.0", true},
		{"201904.0.0.0-dev", "~201903.0", false},
		{"201903.0-beta", "~201903.0-beta", true},
		{"201903.0-alpha", "~201903.0-beta", false},
		{"201903.0", "~201903.0-stable", true},
		{"201903.0-RC1", "~201903.0-stable", false},
		{"201903.205830.1", "~201903.205830.1-stable", true},
		{"201903.205831.0.0-dev", "~201903.205830.1-stable", false},
		{"2.9999999.9999999.9999999-dev", "~2.x-dev", true},
		{"3.0.0.0-dev", "~2.x-dev", false},
		{"2.0.9999999.9999999-dev", "~2.0.x-dev", true},
		{"2.1.0.0-dev", "~2.0.x-dev", false},
		{"2.0.3.9999999-dev", "~2.0.3.x-dev", true},
		{"2.0.4.0-dev", "~2.0.3.x-dev", false},
		{"0.9999999.9999999.9999999-dev", "~0.x-dev", true},
		{"1.0.0.0-dev", "~0.x-dev", false},
		{"1.5.0", "^v1", true},
		{"2.0.0", "^v1", false},
		{"0.5.0", "^0", true},
		{"1.0.0", "^0", false},
		{"0.0.5", "^0.0", true},
		{"0.1.0", "^0.0", false},
		{"1.9.9", "^1.2", true},
		{"2.0.0", "^1.2", false},
		{"1.2.3-beta2", "^1.2.3-beta.2", true},
		{"1.2.3-beta1", "^1.2.3-beta.2", false},
		{"1.2.3.5", "^1.2.3.4", true},
		{"2.0.0", "^1.2.3.4", false},
		{"1.9.9", "^1.2.3", true},
		{"2.0.0", "^1.2.3", false},
		{"0.2.9", "^0.2.3", true},
		{"0.3.0", "^0.2.3", false},
		{"0.2.9", "^0.2", true},
		{"0.3.0", "^0.2", false},
		{"0.2.9", "^0.2.0", true},
		{"0.3.0", "^0.2.0", false},
		{"0.0.3", "^0.0.3", true},
		{"0.0.4", "^0.0.3", false},
		{"0.0.3-alpha", "^0.0.3-alpha", true},
		{"0.0.4-dev", "^0.0.3-alpha", false},
		{"0.0.3-dev", "^0.0.3-dev", true},
		{"0.0.4-dev", "^0.0.3-dev", false},
		{"0.0.3", "^0.0.3-stable", true},
		{"0.0.3-RC1", "^0.0.3-stable", false},
		{"201903.10", "^201903.0", true},
		{"201904.0.0.0-dev", "^201903.0", false},
		{"201903.0-beta", "^201903.0-beta", true},
		{"201903.0-alpha", "^201903.0-beta", false},
		{"201903.205830.1", "^201903.205830.1-stable", true},
		{"201904.0.0.0-dev", "^201903.205830.1-stable", false},
		{"2.0.9999999.9999999-dev", "^2.0.x-dev", true},
		{"3.0.0.0-dev", "^2.0.x-dev", false},
		{"2.0.9999999.9999999-dev", "^2.0.*-dev", true},
		{"3.0.0.0-dev", "^2.0.*-dev", false},
		{"2.0.3.9999999-dev", "^2.0.3.x-dev", true},
		{"3.0.0.0-dev", "^2.0.3.x-dev", false},
		{"2.9999999.9999999.9999999-dev", "^2.x-dev", true},
		{"3.0.0.0-dev", "^2.x-dev", false},
		{"0.9999999.9999999.9999999-dev", "^0.x-dev", true},
		{"1.0.0.0-dev", "^0.x-dev", false},
		{"2.5.0", "v1 - v2", true},
		{"3.0.0", "v1 - v2", false},
		{"2.3.4.5", "1.2.3 - 2.3.4.5", true},
		{"2.3.4.6", "1.2.3 - 2.3.4.5", false},
		{"2.3.9", "1.2-beta - 2.3", true},
		{"2.4.0.0-dev", "1.2-beta - 2.3", false},
		{"2.3.0-dev", "1.2-beta - 2.3-dev", true},
		{"2.3.0", "1.2-beta - 2.3-dev", false},
		{"2.3.1", "1.2-RC - 2.3.1", true},
		{"2.3.1.1", "1.2-RC - 2.3.1", false},
		{"2.3.0-RC", "1.2.3-alpha - 2.3-RC", true},
		{"2.3.0", "1.2.3-alpha - 2.3-RC", false},
		{"2.0.9", "1 - 2.0", true},
		{"2.1.0.0-dev", "1 - 2.0", false},
		{"2.1.9", "1 - 2.1", true},
		{"2.2.0.0-dev", "1 - 2.1", false},
		{"2.1.0", "1.2 - 2.1.0", true},
		{"2.1.0.1", "1.2 - 2.1.0", false},
		{"2.1.3", "1.3 - 2.1.3", true},
		{"2.1.3.1", "1.3 - 2.1.3", false},
		{"3.0.3.9999999-dev", "2.0.3.x-dev - 3.0.3.x-dev", true},
		{"3.0.4.0-dev", "2.0.3.x-dev - 3.0.3.x-dev", false},
		{"3.0.9999999.9999999-dev", "2.0.x-dev - 3.0.x-dev", true},
		{"3.1.0.0-dev", "2.0.x-dev - 3.0.x-dev", false},
		{"3.9999999.9999999.9999999-dev", "2.x-dev - 3.x-dev", true},
		{"4.0.0.0-dev", "2.x-dev - 3.x-dev", false},
		{"1.9999999.9999999.9999999-dev", "0.x-dev - 1.x-dev", true},
		{"2.0.0.0-dev", "0.x-dev - 1.x-dev", false},
	}

	for _, tc := range tests {
		actual, err := checkSatisfiesForTest(tc.version, tc.constraint)
		if err != nil {
			t.Errorf("Satisfies(%q, %q) unexpected error: %v", tc.version, tc.constraint, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("Satisfies(%q, %q): expected %v, got %v", tc.version, tc.constraint, tc.expected, actual)
		}
	}
}

func TestVersionParserRangeSnapshots(t *testing.T) {
	noDev := branchSnapshot{Names: []string{}}
	tests := []struct {
		name       string
		constraint string
		start      string
		end        string
	}{
		{"wildcard v2.*", "v2.*", ">= 2.0.0.0-dev", "< 3.0.0.0-dev"},
		{"wildcard 2.*.*", "2.*.*", ">= 2.0.0.0-dev", "< 3.0.0.0-dev"},
		{"wildcard 20.*", "20.*", ">= 20.0.0.0-dev", "< 21.0.0.0-dev"},
		{"wildcard 20.*.*", "20.*.*", ">= 20.0.0.0-dev", "< 21.0.0.0-dev"},
		{"wildcard 2.0.*", "2.0.*", ">= 2.0.0.0-dev", "< 2.1.0.0-dev"},
		{"wildcard 2.x", "2.x", ">= 2.0.0.0-dev", "< 3.0.0.0-dev"},
		{"wildcard 2.x.x", "2.x.x", ">= 2.0.0.0-dev", "< 3.0.0.0-dev"},
		{"wildcard 2.2.x", "2.2.x", ">= 2.2.0.0-dev", "< 2.3.0.0-dev"},
		{"wildcard 2.10.X", "2.10.X", ">= 2.10.0.0-dev", "< 2.11.0.0-dev"},
		{"wildcard 2.1.3.*", "2.1.3.*", ">= 2.1.3.0-dev", "< 2.1.4.0-dev"},
		{"wildcard 0.*", "0.*", ">= 0.0.0.0-dev", "< 1.0.0.0-dev"},
		{"wildcard 0.*.*", "0.*.*", ">= 0.0.0.0-dev", "< 1.0.0.0-dev"},
		{"wildcard 0.x", "0.x", ">= 0.0.0.0-dev", "< 1.0.0.0-dev"},
		{"wildcard 0.x.x", "0.x.x", ">= 0.0.0.0-dev", "< 1.0.0.0-dev"},
		{"tilde ~v1", "~v1", ">= 1.0.0.0-dev", "< 2.0.0.0-dev"},
		{"tilde ~1.0", "~1.0", ">= 1.0.0.0-dev", "< 2.0.0.0-dev"},
		{"tilde ~1.0.0", "~1.0.0", ">= 1.0.0.0-dev", "< 1.1.0.0-dev"},
		{"tilde ~1.2", "~1.2", ">= 1.2.0.0-dev", "< 2.0.0.0-dev"},
		{"tilde ~1.2.3", "~1.2.3", ">= 1.2.3.0-dev", "< 1.3.0.0-dev"},
		{"tilde ~1.2.3.4", "~1.2.3.4", ">= 1.2.3.4-dev", "< 1.2.4.0-dev"},
		{"tilde ~1.2-beta", "~1.2-beta", ">= 1.2.0.0-beta", "< 2.0.0.0-dev"},
		{"tilde ~1.2-b2", "~1.2-b2", ">= 1.2.0.0-beta2", "< 2.0.0.0-dev"},
		{"tilde ~1.2-BETA2", "~1.2-BETA2", ">= 1.2.0.0-beta2", "< 2.0.0.0-dev"},
		{"tilde ~1.2.2-dev", "~1.2.2-dev", ">= 1.2.2.0-dev", "< 1.3.0.0-dev"},
		{"tilde ~1.2.2-stable", "~1.2.2-stable", ">= 1.2.2.0", "< 1.3.0.0-dev"},
		{"tilde ~201903.0", "~201903.0", ">= 201903.0-dev", "< 201904.0.0.0-dev"},
		{"tilde ~201903.0-beta", "~201903.0-beta", ">= 201903.0-beta", "< 201904.0.0.0-dev"},
		{"tilde ~201903.0-stable", "~201903.0-stable", ">= 201903.0", "< 201904.0.0.0-dev"},
		{"tilde ~201903.205830.1-stable", "~201903.205830.1-stable", ">= 201903.205830.1", "< 201903.205831.0.0-dev"},
		{"tilde ~2.x-dev", "~2.x-dev", ">= 2.9999999.9999999.9999999-dev", "< 3.0.0.0-dev"},
		{"tilde ~2.0.x-dev", "~2.0.x-dev", ">= 2.0.9999999.9999999-dev", "< 2.1.0.0-dev"},
		{"tilde ~2.0.3.x-dev", "~2.0.3.x-dev", ">= 2.0.3.9999999-dev", "< 2.0.4.0-dev"},
		{"tilde ~0.x-dev", "~0.x-dev", ">= 0.9999999.9999999.9999999-dev", "< 1.0.0.0-dev"},
		{"caret ^v1", "^v1", ">= 1.0.0.0-dev", "< 2.0.0.0-dev"},
		{"caret ^0", "^0", ">= 0.0.0.0-dev", "< 1.0.0.0-dev"},
		{"caret ^0.0", "^0.0", ">= 0.0.0.0-dev", "< 0.1.0.0-dev"},
		{"caret ^1.2", "^1.2", ">= 1.2.0.0-dev", "< 2.0.0.0-dev"},
		{"caret ^1.2.3-beta.2", "^1.2.3-beta.2", ">= 1.2.3.0-beta2", "< 2.0.0.0-dev"},
		{"caret ^1.2.3.4", "^1.2.3.4", ">= 1.2.3.4-dev", "< 2.0.0.0-dev"},
		{"caret ^1.2.3", "^1.2.3", ">= 1.2.3.0-dev", "< 2.0.0.0-dev"},
		{"caret ^0.2.3", "^0.2.3", ">= 0.2.3.0-dev", "< 0.3.0.0-dev"},
		{"caret ^0.2", "^0.2", ">= 0.2.0.0-dev", "< 0.3.0.0-dev"},
		{"caret ^0.2.0", "^0.2.0", ">= 0.2.0.0-dev", "< 0.3.0.0-dev"},
		{"caret ^0.0.3", "^0.0.3", ">= 0.0.3.0-dev", "< 0.0.4.0-dev"},
		{"caret ^0.0.3-alpha", "^0.0.3-alpha", ">= 0.0.3.0-alpha", "< 0.0.4.0-dev"},
		{"caret ^0.0.3-dev", "^0.0.3-dev", ">= 0.0.3.0-dev", "< 0.0.4.0-dev"},
		{"caret ^0.0.3-stable", "^0.0.3-stable", ">= 0.0.3.0", "< 0.0.4.0-dev"},
		{"caret ^201903.0", "^201903.0", ">= 201903.0-dev", "< 201904.0.0.0-dev"},
		{"caret ^201903.0-beta", "^201903.0-beta", ">= 201903.0-beta", "< 201904.0.0.0-dev"},
		{"caret ^201903.205830.1-stable", "^201903.205830.1-stable", ">= 201903.205830.1", "< 201904.0.0.0-dev"},
		{"caret ^2.x-dev", "^2.x-dev", ">= 2.9999999.9999999.9999999-dev", "< 3.0.0.0-dev"},
		{"caret ^2.0.*-dev", "^2.0.*-dev", ">= 2.0.9999999.9999999-dev", "< 3.0.0.0-dev"},
		{"caret ^2.0.x-dev", "^2.0.x-dev", ">= 2.0.9999999.9999999-dev", "< 3.0.0.0-dev"},
		{"caret ^2.0.3.x-dev", "^2.0.3.x-dev", ">= 2.0.3.9999999-dev", "< 3.0.0.0-dev"},
		{"caret ^0.x-dev", "^0.x-dev", ">= 0.9999999.9999999.9999999-dev", "< 1.0.0.0-dev"},
		{"hyphen v1 - v2", "v1 - v2", ">= 1.0.0.0-dev", "< 3.0.0.0-dev"},
		{"hyphen 1.2.3 - 2.3.4.5", "1.2.3 - 2.3.4.5", ">= 1.2.3.0-dev", "<= 2.3.4.5"},
		{"hyphen 1.2-beta - 2.3", "1.2-beta - 2.3", ">= 1.2.0.0-beta", "< 2.4.0.0-dev"},
		{"hyphen 1.2-beta - 2.3-dev", "1.2-beta - 2.3-dev", ">= 1.2.0.0-beta", "<= 2.3.0.0-dev"},
		{"hyphen 1.2-RC - 2.3.1", "1.2-RC - 2.3.1", ">= 1.2.0.0-RC", "<= 2.3.1.0"},
		{"hyphen 1.2.3-alpha - 2.3-RC", "1.2.3-alpha - 2.3-RC", ">= 1.2.3.0-alpha", "<= 2.3.0.0-RC"},
		{"hyphen 1 - 2.0", "1 - 2.0", ">= 1.0.0.0-dev", "< 2.1.0.0-dev"},
		{"hyphen 1 - 2.1", "1 - 2.1", ">= 1.0.0.0-dev", "< 2.2.0.0-dev"},
		{"hyphen 1.2 - 2.1.0", "1.2 - 2.1.0", ">= 1.2.0.0-dev", "<= 2.1.0.0"},
		{"hyphen 1.3 - 2.1.3", "1.3 - 2.1.3", ">= 1.3.0.0-dev", "<= 2.1.3.0"},
		{"hyphen 2.0.3.x-dev - 3.0.3.x-dev", "2.0.3.x-dev - 3.0.3.x-dev", ">= 2.0.3.9999999-dev", "<= 3.0.3.9999999-dev"},
		{"hyphen 2.0.x-dev - 3.0.x-dev", "2.0.x-dev - 3.0.x-dev", ">= 2.0.9999999.9999999-dev", "<= 3.0.9999999.9999999-dev"},
		{"hyphen 2.x-dev - 3.x-dev", "2.x-dev - 3.x-dev", ">= 2.9999999.9999999.9999999-dev", "<= 3.9999999.9999999.9999999-dev"},
		{"hyphen 0.x-dev - 1.x-dev", "0.x-dev - 1.x-dev", ">= 0.9999999.9999999.9999999-dev", "<= 1.9999999.9999999.9999999-dev"},
	}

	for _, tc := range tests {
		domains, err := constraintUnionDomains(tc.constraint)
		if err != nil {
			t.Errorf("%s: constraintUnionDomains(%q) unexpected error: %v", tc.name, tc.constraint, err)
			continue
		}
		expected := domainSnapshot{
			Numeric:  []intervalSnapshot{{Start: tc.start, End: tc.end}},
			Branches: noDev,
		}
		if actual := domainUnionSnapshot(domains); !reflect.DeepEqual(actual, expected) {
			t.Errorf("%s: expected %#v, got %#v", tc.name, expected, actual)
		}
	}
}

func TestVersionParserParseConstraintBehavior(t *testing.T) {
	satisfies := []struct {
		version    string
		constraint string
		expected   bool
	}{
		{"1.0.0", "1.0@dev", true},
		{"1.0.0-beta", ">=1.0@beta", true},
		{"1.0.0-alpha", ">=1.0@beta", false},
		{"dev-load-varnish-only-when-used", "dev-load-varnish-only-when-used as ^2.0@dev", true},
		{"dev-load-varnish-only-when-used", "dev-load-varnish-only-when-used@dev as ^2.0@dev", true},
		{"1.0.9999999.9999999-dev", "1.0.x-dev#abcd123", true},
		{"1.0.9999999.9999999-dev", "1.0.x-dev#trunk/@123", true},
		{"2.5", "~2.5.9|~2.6,>=2.6.2", false},
		{"2.5.9", "~2.5.9|~2.6,>=2.6.2", true},
		{"2.6.1", "~2.5.9|~2.6,>=2.6.2", false},
		{"2.6.2", "~2.5.9|~2.6,>=2.6.2", true},
		{"2.0.1", ">2.0,<=3.0", true},
		{"2.0.1", ">2.0  <=3.0", true},
		{"2.0.1", ">2.0, <=3.0", true},
		{"2.0.1", ">2.0 ,<=3.0", true},
		{"2.0.1", ">2.0 , <=3.0", true},
		{"2.0.1", ">2.0   , <=3.0", true},
		{"2.0.1", "> 2.0  ,  <=  3.0", true},
		{"2.0.1", "  > 2.0  ,  <=  3.0 ", true},
		{"3.0", ">2.0 <=3.0", true},
		{"2.0", "> 2.0   <=  3.0", false},
		{"2.0.4", ">2.0,<2.0.5 | >2.0.6", true},
		{"2.0.5", ">2.0,<2.0.5 | >2.0.6", false},
		{"2.0.7", ">2.0,<2.0.5 | >2.0.6", true},
		{"2.0.4", ">2.0,<2.0.5 || >2.0.6", true},
		{"2.0.5", ">2.0,<2.0.5 || >2.0.6", false},
		{"2.0.7", ">2.0,<2.0.5 || >2.0.6", true},
		{"2.0.4", "> 2.0 , <2.0.5 | >  2.0.6", true},
		{"2.0.5", "> 2.0 , <2.0.5 | >  2.0.6", false},
		{"2.0.7", "> 2.0 , <2.0.5 | >  2.0.6", true},
		{"1.1.0-alpha4", ">=1.1.0-alpha4,<1.2.x-dev", true},
		{"1.2.9999999.9999999-dev", ">=1.1.0-alpha4,<1.2.x-dev", false},
		{"1.2.0-beta1", ">=1.1.0-alpha4,<1.2-beta2", true},
		{"1.2.0-beta2", ">=1.1.0-alpha4,<1.2-beta2", false},
		{"3.0-dev", ">2.0@stable,<=3.0@dev", true},
		{"3.0", ">2.0@stable,<=3.0@dev", false},
		{"2.1", ">2.0@stable,@dev", true},
		{"2.0", ">2.0@stable,@dev", false},
		{"0", ">2.0@stable || 0@dev", true},
		{"2.1", ">2.0@stable || 0@dev", true},
		{"1.0", ">2.0@stable || 0@dev", false},
		{"1.0.1", "~0.1 || ~1.0 !=1.0.1", false},
		{"1.0.2", "~0.1 || ~1.0 !=1.0.1", true},
	}

	for _, tc := range satisfies {
		actual, err := checkSatisfiesForTest(tc.version, tc.constraint)
		if err != nil {
			t.Errorf("Satisfies(%q, %q) unexpected error: %v", tc.version, tc.constraint, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("Satisfies(%q, %q): expected %v, got %v", tc.version, tc.constraint, tc.expected, actual)
		}
	}

	invalid := []string{
		"1.0#abcd123",
		"1.0#trunk/@123",
	}
	for _, constraint := range invalid {
		if _, err := NewConstraint(constraint); err == nil {
			t.Errorf("expected constraint %q to be rejected", constraint)
		}
	}
}

func TestVersionParserMultiConstraintSnapshots(t *testing.T) {
	noDev := branchSnapshot{Names: []string{}}
	conjunctive := []string{
		">2.0,<=3.0",
		">2.0 <=3.0",
		">2.0  <=3.0",
		">2.0, <=3.0",
		">2.0 ,<=3.0",
		">2.0 , <=3.0",
		">2.0   , <=3.0",
		"> 2.0   <=  3.0",
		"> 2.0  ,  <=  3.0",
		"  > 2.0  ,  <=  3.0 ",
	}

	for _, constraint := range conjunctive {
		domains, err := constraintUnionDomains(constraint)
		if err != nil {
			t.Errorf("constraintUnionDomains(%q) unexpected error: %v", constraint, err)
			continue
		}
		expected := domainSnapshot{
			Numeric:  []intervalSnapshot{{Start: "> 2.0.0.0", End: "<= 3.0.0.0"}},
			Branches: noDev,
		}
		if actual := domainUnionSnapshot(domains); !reflect.DeepEqual(actual, expected) {
			t.Errorf("constraintUnionDomains(%q): expected %#v, got %#v", constraint, expected, actual)
		}
	}

	disjunctive := []string{
		">2.0,<2.0.5 | >2.0.6",
		">2.0,<2.0.5 || >2.0.6",
		"> 2.0 , <2.0.5 | >  2.0.6",
	}

	for _, constraint := range disjunctive {
		domains, err := constraintUnionDomains(constraint)
		if err != nil {
			t.Errorf("constraintUnionDomains(%q) unexpected error: %v", constraint, err)
			continue
		}
		expected := domainSnapshot{
			Numeric: []intervalSnapshot{
				{Start: "> 2.0.0.0", End: "< 2.0.5.0-dev"},
				{Start: "> 2.0.6.0", End: "< +Inf"},
			},
			Branches: noDev,
		}
		if actual := domainUnionSnapshot(domains); !reflect.DeepEqual(actual, expected) {
			t.Errorf("constraintUnionDomains(%q): expected %#v, got %#v", constraint, expected, actual)
		}
	}

	stability := []struct {
		constraint string
		expected   domainSnapshot
	}{
		{
			constraint: ">=1.1.0-alpha4,<1.2.x-dev",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.1.0.0-alpha4", End: "< 1.2.9999999.9999999-dev"}},
				Branches: noDev,
			},
		},
		{
			constraint: ">=1.1.0-alpha4,<1.2-beta2",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.1.0.0-alpha4", End: "< 1.2.0.0-beta2"}},
				Branches: noDev,
			},
		},
		{
			constraint: ">2.0@stable,<=3.0@dev",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: "> 2.0.0.0", End: "<= 3.0.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			constraint: ">2.0@stable,@dev",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: "> 2.0.0.0", End: "< +Inf"}},
				Branches: noDev,
			},
		},
		{
			constraint: ">2.0@stable || 0@dev",
			expected: domainSnapshot{
				Numeric: []intervalSnapshot{
					{Start: ">= 0.0.0.0", End: "<= 0.0.0.0"},
					{Start: "> 2.0.0.0", End: "< +Inf"},
				},
				Branches: noDev,
			},
		},
	}

	for _, tc := range stability {
		domains, err := constraintUnionDomains(tc.constraint)
		if err != nil {
			t.Errorf("constraintUnionDomains(%q) unexpected error: %v", tc.constraint, err)
			continue
		}
		if actual := domainUnionSnapshot(domains); !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("constraintUnionDomains(%q): expected %#v, got %#v", tc.constraint, tc.expected, actual)
		}
	}
}

func TestVersionParserSimpleConstraintSnapshots(t *testing.T) {
	noDev := branchSnapshot{Names: []string{}}
	allBranches := branchSnapshot{Names: []string{}, Exclude: true}
	allNumeric := []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}}
	exact := func(version string) []intervalSnapshot {
		return []intervalSnapshot{{Start: ">= " + version, End: "<= " + version}}
	}

	tests := []struct {
		name       string
		constraint string
		expected   domainSnapshot
	}{
		{"match any", "*", domainSnapshot{Numeric: allNumeric, Branches: allBranches}},
		{"match any/v", "v*", domainSnapshot{Numeric: allNumeric, Branches: noDev}},
		{"match any/2", "*.*", domainSnapshot{Numeric: allNumeric, Branches: noDev}},
		{"match any/2v", "v*.*", domainSnapshot{Numeric: allNumeric, Branches: noDev}},
		{"match any/3", "*.x.*", domainSnapshot{Numeric: allNumeric, Branches: noDev}},
		{"match any/4", "x.X.x.*", domainSnapshot{Numeric: allNumeric, Branches: noDev}},
		{
			name:       "not equal",
			constraint: "<>1.0.0",
			expected: domainSnapshot{
				Numeric: []intervalSnapshot{
					{Start: ">= 0.0.0.0-dev", End: "< 1.0.0.0"},
					{Start: "> 1.0.0.0", End: "< +Inf"},
				},
				Branches: allBranches,
			},
		},
		{
			name:       "not equal/2",
			constraint: "!=1.0.0",
			expected: domainSnapshot{
				Numeric: []intervalSnapshot{
					{Start: ">= 0.0.0.0-dev", End: "< 1.0.0.0"},
					{Start: "> 1.0.0.0", End: "< +Inf"},
				},
				Branches: allBranches,
			},
		},
		{"greater than", ">1.0.0", domainSnapshot{Numeric: []intervalSnapshot{{Start: "> 1.0.0.0", End: "< +Inf"}}, Branches: noDev}},
		{"lesser than", "<1.2.3.4", domainSnapshot{Numeric: []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< 1.2.3.4-dev"}}, Branches: noDev}},
		{"less/eq than", "<=1.2.3", domainSnapshot{Numeric: []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "<= 1.2.3.0"}}, Branches: noDev}},
		{"great/eq than", ">=1.2.3", domainSnapshot{Numeric: []intervalSnapshot{{Start: ">= 1.2.3.0-dev", End: "< +Inf"}}, Branches: noDev}},
		{"equals", "=1.2.3", domainSnapshot{Numeric: exact("1.2.3.0"), Branches: noDev}},
		{"double equals", "==1.2.3", domainSnapshot{Numeric: exact("1.2.3.0"), Branches: noDev}},
		{"no op means eq", "1.2.3", domainSnapshot{Numeric: exact("1.2.3.0"), Branches: noDev}},
		{"completes version", "=1.0", domainSnapshot{Numeric: exact("1.0.0.0"), Branches: noDev}},
		{"shorthand beta", "1.2.3b5", domainSnapshot{Numeric: exact("1.2.3.0-beta5"), Branches: noDev}},
		{"shorthand alpha", "1.2.3a1", domainSnapshot{Numeric: exact("1.2.3.0-alpha1"), Branches: noDev}},
		{"shorthand patch", "1.2.3p1234", domainSnapshot{Numeric: exact("1.2.3.0-patch1234"), Branches: noDev}},
		{"shorthand patch/2", "1.2.3pl1234", domainSnapshot{Numeric: exact("1.2.3.0-patch1234"), Branches: noDev}},
		{"accepts spaces", ">= 1.2.3", domainSnapshot{Numeric: []intervalSnapshot{{Start: ">= 1.2.3.0-dev", End: "< +Inf"}}, Branches: noDev}},
		{"accepts spaces/2", "< 1.2.3", domainSnapshot{Numeric: []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< 1.2.3.0-dev"}}, Branches: noDev}},
		{"accepts spaces/3", "> 1.2.3", domainSnapshot{Numeric: []intervalSnapshot{{Start: "> 1.2.3.0", End: "< +Inf"}}, Branches: noDev}},
		{"accepts master", ">=dev-master", domainSnapshot{Branches: noDev}},
		{"accepts master/2", "dev-master", domainSnapshot{Branches: branchSnapshot{Names: []string{"dev-master"}}}},
		{"accepts arbitrary", "dev-feature-a", domainSnapshot{Branches: branchSnapshot{Names: []string{"dev-feature-a"}}}},
		{"regression #550", "dev-some-fix", domainSnapshot{Branches: branchSnapshot{Names: []string{"dev-some-fix"}}}},
		{"regression #935", "dev-CAPS", domainSnapshot{Branches: branchSnapshot{Names: []string{"dev-CAPS"}}}},
		{"ignores aliases", "dev-master as 1.0.0", domainSnapshot{Branches: branchSnapshot{Names: []string{"dev-master"}}}},
		{"lesser than override", "<1.2.3.4-stable", domainSnapshot{Numeric: []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< 1.2.3.4"}}, Branches: noDev}},
		{"great/eq than override", ">=1.2.3.4-stable", domainSnapshot{Numeric: []intervalSnapshot{{Start: ">= 1.2.3.4", End: "< +Inf"}}, Branches: noDev}},
	}

	for _, tc := range tests {
		domains, err := constraintUnionDomains(tc.constraint)
		if err != nil {
			t.Errorf("%s: constraintUnionDomains(%q) unexpected error: %v", tc.name, tc.constraint, err)
			continue
		}
		actual := domainUnionSnapshot(domains)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("%s: expected %#v, got %#v", tc.name, tc.expected, actual)
		}
	}
}

func TestVersionParserConstraintSnapshots(t *testing.T) {
	noDev := branchSnapshot{Names: []string{}}
	exact := func(version string) []intervalSnapshot {
		return []intervalSnapshot{{Start: ">= " + version, End: "<= " + version}}
	}
	branchOnly := func(name string) domainSnapshot {
		return domainSnapshot{Branches: branchSnapshot{Names: []string{name}}}
	}

	tests := []struct {
		name       string
		constraint string
		expected   domainSnapshot
	}{
		{"numeric branch", "3.x-dev", domainSnapshot{Numeric: exact("3.9999999.9999999.9999999-dev"), Branches: noDev}},
		{"numeric branch without wildcard", "3-dev", domainSnapshot{Numeric: exact("3.0.0.0-dev"), Branches: noDev}},
		{"non-numeric branch", "dev-3.x", branchOnly("dev-3.x")},
		{"non-numeric branch suffix", "xsd2go-dev", branchOnly("dev-xsd2go")},
		{"non-numeric dotted branch suffix", "3.next-dev", branchOnly("dev-3.next")},
		{"non-numeric branch suffix/2", "foobar-dev", branchOnly("dev-foobar")},
		{"non-numeric dev branch", "dev-xsd2go", branchOnly("dev-xsd2go")},
		{"non-numeric dev branch/2", "dev-3.next", branchOnly("dev-3.next")},
		{"non-numeric dev branch/3", "dev-foobar", branchOnly("dev-foobar")},
		{"dev branch with constraint-like name", "dev-1.0.0-dev<1.0.5-dev", branchOnly("dev-1.0.0-dev<1.0.5-dev")},
		{"dev branch with constraint-like name/2", "dev-1.0.0-dev<1.0.5", branchOnly("dev-1.0.0-dev<1.0.5")},
		{"alias stripped", "foobar-dev as 2.1.0", branchOnly("dev-foobar")},
		{"alias stripped in disjunction", "foobar-dev as 2.1.0 || 3.5", domainSnapshot{Numeric: exact("3.5.0.0"), Branches: branchSnapshot{Names: []string{"dev-foobar"}}}},
		{"alias stripped in repeated disjunction", "foobar-dev as 2.1.0 || 3.5 as 1.5", domainSnapshot{Numeric: exact("3.5.0.0"), Branches: branchSnapshot{Names: []string{"dev-foobar"}}}},
		{"hyphen dev upper", "2.1.0 - 2.3-dev", domainSnapshot{Numeric: []intervalSnapshot{{Start: ">= 2.1.0.0-dev", End: "<= 2.3.0.0-dev"}}, Branches: noDev}},
		{"hyphen numeric dev upper wildcard", "1.0 - 2.0.x-dev", domainSnapshot{Numeric: []intervalSnapshot{{Start: ">= 1.0.0.0-dev", End: "<= 2.0.9999999.9999999-dev"}}, Branches: noDev}},
		{"caret with trailing dot", "^1.", domainSnapshot{Numeric: []intervalSnapshot{{Start: ">= 1.0.0.0-dev", End: "< 2.0.0.0-dev"}}, Branches: noDev}},
		{"tilde with trailing dot", "~1.", domainSnapshot{Numeric: []intervalSnapshot{{Start: ">= 1.0.0.0-dev", End: "< 2.0.0.0-dev"}}, Branches: noDev}},
		{"version with trailing dot", "1.2.", domainSnapshot{Numeric: exact("1.2.0.0"), Branches: noDev}},
		{"version with repeated dot dev", "1.2..dev", domainSnapshot{Numeric: exact("1.2.0.0-dev"), Branches: noDev}},
		{"version with dash dot dev", "1.2-.dev", domainSnapshot{Numeric: exact("1.2.0.0-dev"), Branches: noDev}},
		{"version with underscore dev", "1.2_-dev", domainSnapshot{Numeric: exact("1.2.0.0-dev"), Branches: noDev}},
	}

	for _, tc := range tests {
		domains, err := constraintUnionDomains(tc.constraint)
		if err != nil {
			t.Errorf("%s: constraintUnionDomains(%q) unexpected error: %v", tc.name, tc.constraint, err)
			continue
		}
		actual := domainUnionSnapshot(domains)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("%s: expected %#v, got %#v", tc.name, tc.expected, actual)
		}
	}
}

func TestConstraintVersionMatchesCases(t *testing.T) {
	tests := []struct {
		requireOp      string
		requireVersion string
		provideOp      string
		provideVersion string
		expected       bool
	}{
		{"==", "2", "==", "2", true},
		{"==", "2", "<", "3", true},
		{"==", "2", "<=", "2", true},
		{"==", "2", "<=", "3", true},
		{"==", "2", ">=", "1", true},
		{"==", "2", ">=", "2", true},
		{"==", "2", ">", "1", true},
		{"==", "2", "!=", "1", true},
		{"==", "2", "!=", "3", true},
		{"<", "2", "==", "1", true},
		{"<", "2", "<", "1", true},
		{"<", "2", "<", "2", true},
		{"<", "2", "<", "3", true},
		{"<", "2", "<=", "1", true},
		{"<", "2", "<=", "2", true},
		{"<", "2", "<=", "3", true},
		{"<", "2", ">=", "1", true},
		{"<", "2", ">", "1", true},
		{"<", "2", "!=", "1", true},
		{"<", "2", "!=", "2", true},
		{"<", "2", "!=", "3", true},
		{"<=", "2", "==", "1", true},
		{"<=", "2", "==", "2", true},
		{"<=", "2", "<", "1", true},
		{"<=", "2", "<", "2", true},
		{"<=", "2", "<", "3", true},
		{"<=", "2", "<=", "1", true},
		{"<=", "2", "<=", "2", true},
		{"<=", "2", "<=", "3", true},
		{"<=", "2", ">=", "1", true},
		{"<=", "2", ">=", "2", true},
		{"<=", "2", ">", "1", true},
		{"<=", "2", "!=", "1", true},
		{"<=", "2", "!=", "2", true},
		{"<=", "2", "!=", "3", true},
		{">=", "2", "==", "2", true},
		{">=", "2", "==", "3", true},
		{">=", "2", "<", "3", true},
		{">=", "2", "<=", "2", true},
		{">=", "2", "<=", "3", true},
		{">=", "2", ">=", "1", true},
		{">=", "2", ">=", "2", true},
		{">=", "2", ">=", "3", true},
		{">=", "2", ">", "1", true},
		{">=", "2", ">", "2", true},
		{">=", "2", ">", "3", true},
		{">=", "2", "!=", "1", true},
		{">=", "2", "!=", "2", true},
		{">=", "2", "!=", "3", true},
		{">", "2", "==", "3", true},
		{">", "2", "<", "3", true},
		{">", "2", "<=", "3", true},
		{">", "2", ">=", "1", true},
		{">", "2", ">=", "2", true},
		{">", "2", ">=", "3", true},
		{">", "2", ">", "1", true},
		{">", "2", ">", "2", true},
		{">", "2", ">", "3", true},
		{">", "2", "!=", "1", true},
		{">", "2", "!=", "2", true},
		{">", "2", "!=", "3", true},
		{"!=", "2", "!=", "1", true},
		{"!=", "2", "!=", "2", true},
		{"!=", "2", "!=", "3", true},
		{"!=", "2", "==", "1", true},
		{"!=", "2", "==", "3", true},
		{"!=", "2", "<", "1", true},
		{"!=", "2", "<", "2", true},
		{"!=", "2", "<", "3", true},
		{"!=", "2", "<=", "1", true},
		{"!=", "2", "<=", "2", true},
		{"!=", "2", "<=", "3", true},
		{"!=", "2", ">=", "1", true},
		{"!=", "2", ">=", "2", true},
		{"!=", "2", ">=", "3", true},
		{"!=", "2", ">", "1", true},
		{"!=", "2", ">", "2", true},
		{"!=", "2", ">", "3", true},
		{"==", "dev-foo-bar", "==", "dev-foo-bar", true},
		{"==", "dev-events+issue-17", "==", "dev-events+issue-17", true},
		{"==", "dev-foo-bar", "!=", "dev-foo-xyz", true},
		{"!=", "dev-foo-bar", "!=", "dev-foo-xyz", true},
		{"==", "0.12", "!=", "dev-foo", true},
		{"<", "0.12", "!=", "dev-foo", true},
		{"<=", "0.12", "!=", "dev-foo", true},
		{">=", "0.12", "!=", "dev-foo", true},
		{">", "0.12", "!=", "dev-foo", true},
		{"!=", "0.12", "==", "dev-foo", true},
		{"!=", "0.12", "!=", "dev-foo", true},
		{"==", "2", "==", "1", false},
		{"==", "2", "==", "3", false},
		{"==", "2", "<", "1", false},
		{"==", "2", "<", "2", false},
		{"==", "2", "<=", "1", false},
		{"==", "2", ">=", "3", false},
		{"==", "2", ">", "2", false},
		{"==", "2", ">", "3", false},
		{"==", "2", "!=", "2", false},
		{"<", "2", "==", "2", false},
		{"<", "2", "==", "3", false},
		{"<", "2", ">=", "2", false},
		{"<", "2", ">=", "3", false},
		{"<", "2", ">", "2", false},
		{"<", "2", ">", "3", false},
		{"<=", "2", "==", "3", false},
		{"<=", "2", ">=", "3", false},
		{"<=", "2", ">", "2", false},
		{"<=", "2", ">", "3", false},
		{">=", "2", "==", "1", false},
		{">=", "2", "<", "1", false},
		{">=", "2", "<", "2", false},
		{">=", "2", "<=", "1", false},
		{">", "2", "==", "1", false},
		{">", "2", "==", "2", false},
		{">", "2", "<", "1", false},
		{">", "2", "<", "2", false},
		{">", "2", "<=", "1", false},
		{">", "2", "<=", "2", false},
		{"!=", "2", "==", "2", false},
		{"==", "2.0-b2", "<", "2.0-beta2", false},
		{"==", "dev-foo-dist", "==", "dev-foo-zist", false},
		{"==", "dev-foo-bar", "==", "dev-foo-xyz", false},
		{"==", "dev-foo-bar", "<", "dev-foo-xyz", false},
		{"==", "dev-foo-bar", "<=", "dev-foo-xyz", false},
		{"==", "dev-foo-bar", ">=", "dev-foo-xyz", false},
		{"==", "dev-foo-bar", ">", "dev-foo-xyz", false},
		{"<", "dev-foo-bar", "==", "dev-foo-xyz", false},
		{"<", "dev-foo-bar", "<", "dev-foo-xyz", false},
		{"<", "dev-foo-bar", "<=", "dev-foo-xyz", false},
		{"<", "dev-foo-bar", ">=", "dev-foo-xyz", false},
		{"<", "dev-foo-bar", ">", "dev-foo-xyz", false},
		{"<", "dev-foo-bar", "!=", "dev-foo-xyz", false},
		{"<=", "dev-foo-bar", "==", "dev-foo-xyz", false},
		{"<=", "dev-foo-bar", "<", "dev-foo-xyz", false},
		{"<=", "dev-foo-bar", "<=", "dev-foo-xyz", false},
		{"<=", "dev-foo-bar", ">=", "dev-foo-xyz", false},
		{"<=", "dev-foo-bar", ">", "dev-foo-xyz", false},
		{"<=", "dev-foo-bar", "!=", "dev-foo-xyz", false},
		{">=", "dev-foo-bar", "==", "dev-foo-xyz", false},
		{">=", "dev-foo-bar", "<", "dev-foo-xyz", false},
		{">=", "dev-foo-bar", "<=", "dev-foo-xyz", false},
		{">=", "dev-foo-bar", ">=", "dev-foo-xyz", false},
		{">=", "dev-foo-bar", ">", "dev-foo-xyz", false},
		{">=", "dev-foo-bar", "!=", "dev-foo-xyz", false},
		{">", "dev-foo-bar", "==", "dev-foo-xyz", false},
		{">", "dev-foo-bar", "<", "dev-foo-xyz", false},
		{">", "dev-foo-bar", "<=", "dev-foo-xyz", false},
		{">", "dev-foo-bar", ">=", "dev-foo-xyz", false},
		{">", "dev-foo-bar", ">", "dev-foo-xyz", false},
		{">", "dev-foo-bar", "!=", "dev-foo-xyz", false},
		{"==", "dev-foo-bar", "<", "dev-foo-bar", false},
		{"==", "dev-foo-bar", "<=", "dev-foo-bar", false},
		{"==", "dev-foo-bar", ">=", "dev-foo-bar", false},
		{"==", "dev-foo-bar", ">", "dev-foo-bar", false},
		{"==", "dev-foo-bar", "!=", "dev-foo-bar", false},
		{"<", "dev-foo-bar", "==", "dev-foo-bar", false},
		{"<", "dev-foo-bar", "<", "dev-foo-bar", false},
		{"<", "dev-foo-bar", "<=", "dev-foo-bar", false},
		{"<", "dev-foo-bar", ">=", "dev-foo-bar", false},
		{"<", "dev-foo-bar", ">", "dev-foo-bar", false},
		{"<", "dev-foo-bar", "!=", "dev-foo-bar", false},
		{"<=", "dev-foo-bar", "==", "dev-foo-bar", false},
		{"<=", "dev-foo-bar", "<", "dev-foo-bar", false},
		{"<=", "dev-foo-bar", "<=", "dev-foo-bar", false},
		{"<=", "dev-foo-bar", ">=", "dev-foo-bar", false},
		{"<=", "dev-foo-bar", ">", "dev-foo-bar", false},
		{"<=", "dev-foo-bar", "!=", "dev-foo-bar", false},
		{">=", "dev-foo-bar", "==", "dev-foo-bar", false},
		{">=", "dev-foo-bar", "<", "dev-foo-bar", false},
		{">=", "dev-foo-bar", "<=", "dev-foo-bar", false},
		{">=", "dev-foo-bar", ">=", "dev-foo-bar", false},
		{">=", "dev-foo-bar", ">", "dev-foo-bar", false},
		{">=", "dev-foo-bar", "!=", "dev-foo-bar", false},
		{">", "dev-foo-bar", "==", "dev-foo-bar", false},
		{">", "dev-foo-bar", "<", "dev-foo-bar", false},
		{">", "dev-foo-bar", "<=", "dev-foo-bar", false},
		{">", "dev-foo-bar", ">=", "dev-foo-bar", false},
		{">", "dev-foo-bar", ">", "dev-foo-bar", false},
		{">", "dev-foo-bar", "!=", "dev-foo-bar", false},
		{"==", "0.12", "==", "dev-foo", false},
		{"==", "0.12", "<", "dev-foo", false},
		{"==", "0.12", "<=", "dev-foo", false},
		{"==", "0.12", ">=", "dev-foo", false},
		{"==", "0.12", ">", "dev-foo", false},
		{"<", "0.12", "==", "dev-foo", false},
		{"<", "0.12", "<", "dev-foo", false},
		{"<", "0.12", "<=", "dev-foo", false},
		{"<", "0.12", ">=", "dev-foo", false},
		{"<", "0.12", ">", "dev-foo", false},
		{"<=", "0.12", "==", "dev-foo", false},
		{"<=", "0.12", "<", "dev-foo", false},
		{"<=", "0.12", "<=", "dev-foo", false},
		{"<=", "0.12", ">=", "dev-foo", false},
		{"<=", "0.12", ">", "dev-foo", false},
		{">=", "0.12", "==", "dev-foo", false},
		{">=", "0.12", "<", "dev-foo", false},
		{">=", "0.12", "<=", "dev-foo", false},
		{">=", "0.12", ">=", "dev-foo", false},
		{">=", "0.12", ">", "dev-foo", false},
		{">", "0.12", "==", "dev-foo", false},
		{">", "0.12", "<", "dev-foo", false},
		{">", "0.12", "<=", "dev-foo", false},
		{">", "0.12", ">=", "dev-foo", false},
		{">", "0.12", ">", "dev-foo", false},
		{"!=", "0.12", "<", "dev-foo", false},
		{"!=", "0.12", "<=", "dev-foo", false},
		{"!=", "0.12", ">=", "dev-foo", false},
		{"!=", "0.12", ">", "dev-foo", false},
	}

	for _, tc := range tests {
		left := tc.requireOp + " " + tc.requireVersion
		right := tc.provideOp + " " + tc.provideVersion
		actual, err := ConstraintIntersects(left, right)
		if err != nil {
			t.Errorf("ConstraintIntersects(%q, %q) unexpected error: %v", left, right, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("ConstraintIntersects(%q, %q): expected %v, got %v", left, right, tc.expected, actual)
		}

		actual, err = ConstraintIntersects(right, left)
		if err != nil {
			t.Errorf("ConstraintIntersects(%q, %q) unexpected error: %v", right, left, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("ConstraintIntersects(%q, %q): expected %v, got %v", right, left, tc.expected, actual)
		}
	}
}

func TestConstraintBoundsCases(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		version  string
		lower    boundSnapshot
		upper    boundSnapshot
	}{
		{"equal to 1.0.0.0", "==", "1.0.0.0", boundSnapshot{"1.0.0.0", true}, boundSnapshot{"1.0.0.0", true}},
		{"equal to 1.0.0.0-rc3", "==", "1.0.0.0-rc3", boundSnapshot{"1.0.0.0-RC3", true}, boundSnapshot{"1.0.0.0-RC3", true}},
		{"greater/equal dev feature branch", ">=", "dev-feature-branch", zeroBoundSnapshot(), positiveInfinityBoundSnapshot()},
		{"lower than 0.0.4.0", "<", "0.0.4.0", zeroBoundSnapshot(), boundSnapshot{"0.0.4.0", false}},
		{"lower than 1.0.0.0", "<", "1.0.0.0", zeroBoundSnapshot(), boundSnapshot{"1.0.0.0", false}},
		{"lower than 2.0.0.0", "<", "2.0.0.0", zeroBoundSnapshot(), boundSnapshot{"2.0.0.0", false}},
		{"lower than 3.0.3.0", "<", "3.0.3.0", zeroBoundSnapshot(), boundSnapshot{"3.0.3.0", false}},
		{"lower than 3.0.3.0-rc3", "<", "3.0.3.0-rc3", zeroBoundSnapshot(), boundSnapshot{"3.0.3.0-RC3", false}},
		{"lower than dev feature branch", "<", "dev-feature-branch", zeroBoundSnapshot(), positiveInfinityBoundSnapshot()},
		{"greater than 0.0.4.0", ">", "0.0.4.0", boundSnapshot{"0.0.4.0", false}, positiveInfinityBoundSnapshot()},
		{"greater than 1.0.0.0", ">", "1.0.0.0", boundSnapshot{"1.0.0.0", false}, positiveInfinityBoundSnapshot()},
		{"greater than 2.0.0.0", ">", "2.0.0.0", boundSnapshot{"2.0.0.0", false}, positiveInfinityBoundSnapshot()},
		{"greater than 3.0.3.0", ">", "3.0.3.0", boundSnapshot{"3.0.3.0", false}, positiveInfinityBoundSnapshot()},
		{"greater than 3.0.3.0-rc3", ">", "3.0.3.0-rc3", boundSnapshot{"3.0.3.0-RC3", false}, positiveInfinityBoundSnapshot()},
		{"greater than dev feature branch", ">", "dev-feature-branch", zeroBoundSnapshot(), positiveInfinityBoundSnapshot()},
		{"lower/equal 0.0.4.0", "<=", "0.0.4.0", zeroBoundSnapshot(), boundSnapshot{"0.0.4.0", true}},
		{"lower/equal 1.0.0.0", "<=", "1.0.0.0", zeroBoundSnapshot(), boundSnapshot{"1.0.0.0", true}},
		{"lower/equal 2.0.0.0", "<=", "2.0.0.0", zeroBoundSnapshot(), boundSnapshot{"2.0.0.0", true}},
		{"lower/equal 3.0.3.0", "<=", "3.0.3.0", zeroBoundSnapshot(), boundSnapshot{"3.0.3.0", true}},
		{"lower/equal 3.0.3.0-rc3", "<=", "3.0.3.0-rc3", zeroBoundSnapshot(), boundSnapshot{"3.0.3.0-RC3", true}},
		{"lower/equal dev feature branch", "<=", "dev-feature-branch", zeroBoundSnapshot(), positiveInfinityBoundSnapshot()},
		{"greater/equal 0.0.4.0", ">=", "0.0.4.0", boundSnapshot{"0.0.4.0", true}, positiveInfinityBoundSnapshot()},
		{"greater/equal 1.0.0.0", ">=", "1.0.0.0", boundSnapshot{"1.0.0.0", true}, positiveInfinityBoundSnapshot()},
		{"greater/equal 2.0.0.0", ">=", "2.0.0.0", boundSnapshot{"2.0.0.0", true}, positiveInfinityBoundSnapshot()},
		{"greater/equal 3.0.3.0", ">=", "3.0.3.0", boundSnapshot{"3.0.3.0", true}, positiveInfinityBoundSnapshot()},
		{"greater/equal 3.0.3.0-rc3", ">=", "3.0.3.0-rc3", boundSnapshot{"3.0.3.0-RC3", true}, positiveInfinityBoundSnapshot()},
		{"not equal to 1.0.0.0", "<>", "1.0.0.0", zeroBoundSnapshot(), positiveInfinityBoundSnapshot()},
	}

	for _, tc := range tests {
		lower, upper, err := constraintBoundsSnapshot(tc.operator, tc.version)
		if err != nil {
			t.Errorf("%s: constraintBoundsSnapshot unexpected error: %v", tc.name, err)
			continue
		}
		if lower != tc.lower || upper != tc.upper {
			t.Errorf("%s: expected lower %#v upper %#v, got lower %#v upper %#v", tc.name, tc.lower, tc.upper, lower, upper)
		}
	}
}

func TestConstraintMatrixCases(t *testing.T) {
	versions := []string{"1.0", "2.0", "dev-master", "dev-foo", "3.0-b2", "3.0-beta2"}
	operators := []string{"==", "!=", ">", "<", ">=", "<="}

	for _, requireVersion := range versions {
		for _, requireOperator := range operators {
			requireConstraint := requireOperator + " " + requireVersion
			for _, provideVersion := range versions {
				for _, provideOperator := range operators {
					provideConstraint := provideOperator + " " + provideVersion
					actual, err := ConstraintIntersects(requireConstraint, provideConstraint)
					if err != nil {
						t.Errorf("ConstraintIntersects(%q, %q) unexpected error: %v", requireConstraint, provideConstraint, err)
						continue
					}
					reverse, err := ConstraintIntersects(provideConstraint, requireConstraint)
					if err != nil {
						t.Errorf("ConstraintIntersects(%q, %q) unexpected error: %v", provideConstraint, requireConstraint, err)
						continue
					}
					if actual != reverse {
						t.Errorf("ConstraintIntersects symmetry mismatch for %q and %q: %v vs %v", requireConstraint, provideConstraint, actual, reverse)
					}

					if provideOperator == "==" {
						satisfied, err := checkSatisfiesForTest(provideVersion, requireConstraint)
						if err != nil {
							t.Errorf("Satisfies(%q, %q) unexpected error: %v", provideVersion, requireConstraint, err)
							continue
						}
						if actual != satisfied {
							t.Errorf("ConstraintIntersects(%q, %q): expected exact-version result %v, got %v", requireConstraint, provideConstraint, satisfied, actual)
						}
					}
				}
			}
		}
	}
}

func TestConstraintInvalidOperatorCases(t *testing.T) {
	tests := []string{
		"invalid 1.2.3",
		"! 1.2.3",
		"equals 1.2.3",
	}

	for _, constraint := range tests {
		if _, err := NewConstraint(constraint); err == nil {
			t.Errorf("expected invalid operator constraint %q to be rejected", constraint)
		}
	}
}

func TestConstraintComparableBranches(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		expected   bool
	}{
		{"dev-foo", ">0.12", false},
		{"dev-foo", "<0.12", false},
		{"dev-foo", ">=0.12", false},
		{"dev-foo", "<=0.12", false},
		{"dev-foo", "0.12", false},
		{"dev-foo", "!=0.12", true},
	}

	for _, tc := range tests {
		actual, err := checkSatisfiesForTest(tc.version, tc.constraint)
		if err != nil {
			t.Errorf("Satisfies(%q, %q) unexpected error: %v", tc.version, tc.constraint, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("Satisfies(%q, %q): expected %v, got %v", tc.version, tc.constraint, tc.expected, actual)
		}
	}

	intersects, err := ConstraintIntersects(">0.12", "dev-foo")
	if err != nil {
		t.Fatalf("ConstraintIntersects unexpected error: %v", err)
	}
	if intersects {
		t.Errorf("numeric range should not intersect exact dev branch")
	}
}

func TestMultiConstraintBoundsCases(t *testing.T) {
	tests := []struct {
		name        string
		parts       []string
		conjunctive bool
		raw         bool
		lower       boundSnapshot
		upper       boundSnapshot
	}{
		{"all equal", []string{"1.0.0.0", "1.0.0.0"}, true, true, boundSnapshot{"1.0.0.0", true}, boundSnapshot{"1.0.0.0", true}},
		{"greater should take precedence over greater/equal when conjunctive", []string{">1.0.0.0", ">=1.0.0.0", ">1.0.0.0"}, true, true, boundSnapshot{"1.0.0.0", false}, positiveInfinityBoundSnapshot()},
		{"greater/equal should take precedence over greater when disjunctive", []string{">1.0.0.0", ">=1.0.0.0", ">1.0.0.0"}, false, true, boundSnapshot{"1.0.0.0", true}, positiveInfinityBoundSnapshot()},
		{"bounds limited when conjunctive", []string{">=7.0.0.0", "<8.0.0.0"}, true, true, boundSnapshot{"7.0.0.0", true}, boundSnapshot{"8.0.0.0", false}},
		{"bounds unlimited when disjunctive", []string{">=7.0.0.0", "<8.0.0.0"}, false, true, zeroBoundSnapshot(), positiveInfinityBoundSnapshot()},
		{"integration ^7.0", []string{"^7.0"}, false, false, boundSnapshot{"7.0.0.0-dev", true}, boundSnapshot{"8.0.0.0-dev", false}},
		{"integration ^7.2", []string{"^7.2"}, false, false, boundSnapshot{"7.2.0.0-dev", true}, boundSnapshot{"8.0.0.0-dev", false}},
		{"integration 7.4 wildcard", []string{"7.4.*"}, false, false, boundSnapshot{"7.4.0.0-dev", true}, boundSnapshot{"7.5.0.0-dev", false}},
		{"integration 7.2 or 7.4 wildcard", []string{"7.2.* || 7.4.*"}, false, false, boundSnapshot{"7.2.0.0-dev", true}, boundSnapshot{"7.5.0.0-dev", false}},
		{"multiple multi constraints merging", []string{"^7.0", "^7.2", "7.4.*", "7.2.* || 7.4.*"}, true, false, boundSnapshot{"7.4.0.0-dev", true}, boundSnapshot{"7.5.0.0-dev", false}},
		{"multiple multi constraints merging with gaps", []string{"^7.1.15 || ^7.2.3", "^7.2.2"}, true, false, boundSnapshot{"7.2.2.0-dev", true}, boundSnapshot{"8.0.0.0-dev", false}},
	}

	for _, tc := range tests {
		var lower boundSnapshot
		var upper boundSnapshot
		var err error
		if tc.raw {
			lower, upper, err = combineRawConstraintBoundsSnapshot(tc.parts, tc.conjunctive)
		} else {
			var domains []constraintDomain
			domains, err = composeConstraintDomains(tc.parts, tc.conjunctive)
			if err == nil {
				lower, upper = domainUnionBoundsSnapshot(domains)
			}
		}
		if err != nil {
			t.Errorf("%s: compose bounds unexpected error: %v", tc.name, err)
			continue
		}
		if lower != tc.lower || upper != tc.upper {
			t.Errorf("%s: expected lower %#v upper %#v, got lower %#v upper %#v", tc.name, tc.lower, tc.upper, lower, upper)
		}
	}
}

func TestSubsetsCases(t *testing.T) {
	tests := []struct {
		left     string
		right    string
		expected bool
	}{
		{"*", "*", true},
		{"*", "!= 1 || == 1", true},
		{"1.0.0", "*", true},
		{"1.0.*", "*", true},
		{"^1.0 || ^2.0", "*", true},
		{"^3.0", "^3.2 || *", true},
		{"^1.0 || ^2.0", "^1.0 || ^2.0", true},
		{"^1.0 || ^2.0", "^1.0 || ^2.0 || ^4.0", true},
		{"^1.0 || ^2.1", "^1.0 || ^2.1 || ^4.0", true},
		{"^1.2", "^1.0 || ^2.0", true},
		{"1.2.3", "^1.0 || ^2.0", true},
		{"2.0.0-dev", "^1.0 || ^2.0", true},
		{">= 2.1.0", ">= 2.0.0", true},
		{"^2.0", "<3.0.0", true},
		{"^3.0", "> 2.1.3", true},
		{"3.0.0", "<= 3.0.0", true},
		{"!= 3.0.0", "*", true},
		{"!= 3.0.0", "!= 3.0", true},
		{"!= 3.0, != 2.0", "!= 2.0, != 3.0", true},
		{">3", "^2 || ^3 || >=4", true},
		{">3", ">=3", true},
		{"<3", "<=3", true},
		{"= dev-foo", "= dev-foo", true},
		{"!= dev-foo", "!= dev-foo", true},
		{"< dev-foo", "= dev-foo", true},
		{"1.5.*", "^1.4", true},
		{"1.5.*", "1.3 - 1.6 || 1.8 - 1.9", true},
		{"1.3.2", "1.3.0 || 1.3.1 || 1.3.2", true},
		{"1.3.1", "1.3.0 || 1.3.1 || 1.3.2", true},
		{"1.3.1 || 1.3.1", "1.3.1", true},
		{"^1.0 || ^3.2", "^1.0 || ^3.0", true},
		{"^1.3 || ^3.2", ">1.2", true},
		{"^1.6", "<1.3 || >1.5", true},
		{">1.6", "<1.3 || >1.5", true},
		{">1.6", ">1.5, >1.4, !=1.1", true},
		{">1.6", ">1.5 || >1.7", true},
		{"^1.1", "> 1.0.0", true},
		{"^1.1, !=1.5.0", "> 1.0.0", true},
		{"^1.1, !=0.5.0", "> 1.0.0", true},
		{"^2.0 || dev-foo", "> 1.0 || dev-foo || dev-bar", true},
		{"^1.0, ^1.2", ">=1.2", true},
		{"^1.0, ^1.2", "^1.2", true},
		{"^1.0, ^1.2 || ^1.3", "^1.2", true},
		{"*", ">= 1 || < 1", false},
		{"*", "1.0.0", false},
		{"*", "1.0.*", false},
		{"*", "^1.0 || ^2.0", false},
		{"^1.0 || ^2.0", "^1.0, ^2.0", false},
		{"^1.0 || ^2.0", "^1.2", false},
		{"^1.0 || ^2.0", "^1.0", false},
		{"^1.0 || ^2.0", "1.2.3", false},
		{"^1.0 || ^3.0", "1.2.3", false},
		{"3.0.0", "^1.0 || ^2.0", false},
		{"3.0.0", "< 3.0.0", false},
		{"3.0.0", ">= 3.0.1", false},
		{"!= 3.0.0", "> 3.0.0 || < 3.0.0-stable", false},
		{"!= 3.0.0-dev", "^2.0 || <2 || >3.0-dev", false},
		{"!= 3.0.0", "= 3.0.0", false},
		{"!= 3.0.0", "!= 3.0.1", false},
		{"!= 3.0.0", "dev-foo || dev-bar", false},
		{"!= 3.0.0", "<dev-foo || >dev-bar", false},
		{">= 1.0.0", "= 1.2.3", false},
		{"< 2.0.0", "= 1.2.3", false},
		{">3", "^2 || ^3 || >4", false},
		{">=3", ">3", false},
		{"<=3", "<3", false},
		{"^2.1", "^2.0, !=2.1.3", false},
		{"<2.0", ">=1.1", false},
		{"!= dev-foo", "!= dev-bar", false},
		{"!= dev-foo", "= dev-bar", false},
		{"1.3.3", "1.3.0 || 1.3.1 || 1.3.2", false},
		{"1.3.1 || 1.3.2", "1.3.1", false},
		{">1.6", ">1.5, >1.4, !=1.7", false},
		{">1.6", ">1.5, >1.7", false},
		{"^1.0 || ^3.2", "^1.2 || ^3.0", false},
		{"^1.0 || ^3.2", "^3.0", false},
		{"^1.3 || ^3.2", ">1.4", false},
		{"^2.0 || dev-foo", "> 1.0 || dev-bar", false},
	}

	for _, tc := range tests {
		actual, err := ConstraintSubsetOf(tc.left, tc.right)
		if err != nil {
			t.Errorf("ConstraintSubsetOf(%q, %q) unexpected error: %v", tc.left, tc.right, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("ConstraintSubsetOf(%q, %q): expected %v, got %v", tc.left, tc.right, tc.expected, actual)
		}
	}
}

func TestIntervalsCompactCases(t *testing.T) {
	tests := []struct {
		name        string
		expected    string
		toCompact   []string
		conjunctive bool
	}{
		{"simple disjunctive multi", "1.0 - 1.2 || ^1.5", []string{"1.0 - 1.2 || ^1.5", "1.8 - 1.9 || ^1.12"}, false},
		{"simple conjunctive multi", "1.8 - 1.9 || ^1.12", []string{"1.0 - 1.2 || ^1.5", "1.8 - 1.9 || ^1.12"}, true},
		{"dev constraints propagate, disjunctive", "1.8 - 1.9 || ^1.12 || dev-master || dev-foo", []string{"1.8 - 1.9 || ^1.12", "dev-master", "dev-foo"}, false},
		{"dev constraints + numeric constraint, conjunctive results in match-none", "", []string{"1.8 - 1.9 || ^1.12", "dev-master", "dev-foo"}, true},
		{"conflicting numeric constraint, conjunctive results in match-none", "", []string{"1.0", "2.0"}, true},
		{"simple disjunctive results in same output", "1.0 || 2.0", []string{"1.0", "2.0"}, false},
		{"simple conjunctive results in same output", "!= 1.2, != 1.6", []string{"!= 1.2", "!= 1.6"}, true},
		{"simple conjunctive results in same output/2", "!= 1.0, != 2.0", []string{"!= 1.0", "!= 2.0"}, true},
		{"switches to conjunctive if more than != x is present", ">1.5, != 2.0", []string{"!= 2.0", "> 1.5"}, true},
		{"complex conjunctive with dev", "!= 1.0, != 2.0", []string{"!= 1.0", "!= 2.0"}, true},
		{"simple disjunctive with negation", "!= 1.0", []string{"!= 1.0", "!= 1.0"}, false},
		{"disjunctive with complex negation", "*", []string{"!= 1.0", "!= 1.0", "!= dev-foo", "1.0.5.*"}, false},
		{"conjunctive with complex negation", "1.0.5.*", []string{"!= 1.0", "!= 1.0", "!= dev-foo", "1.0.5.*"}, true},
		{"conjunctive with complex negation/2", ">= 1.0-dev, != 1.2-stable, <2", []string{"!= 1.2", "!= dev-foo", "!= dev-bar", "1.*"}, true},
		{"conjunctive with complex negation/3", "!= 1.2, != dev-foo, != dev-bar", []string{"!= 1.2", "!= dev-foo", "!= dev-bar"}, true},
		{"disjunctive with complex negation/3", "*", []string{"!= 1.2", "!= dev-foo", "!= dev-bar"}, false},
		{"conjunctive with complex negation/4", "== dev-foo", []string{"!= 1.2", "== dev-foo", "!= dev-bar"}, true},
		{"disjunctive with complex negation and dev ==", "*", []string{"!= 1.0", "!= 1.0", "!= dev-foo", "1.0.5.*", "== dev-bla"}, false},
		{"conjunctive with complex negation and dev ==", "dev-bla", []string{"!= 1.0", "!= 1.0", "!= dev-foo", "== dev-bla"}, true},
		{"complex conjunctive which can not match anything", "", []string{"!= 1.0", "!= 1.0", "!= dev-foo", "1.0.5.*", "== dev-bla"}, true},
		{"conjunctive with more than one dev negation", "!= dev-master, != dev-foo", []string{"!= dev-master", "!= dev-foo"}, true},
		{"conjunctive with mix of devs", "== dev-foo", []string{"!= dev-master", "== dev-foo"}, true},
		{"disjunctive with mix of devs", "!= dev-master", []string{"!= dev-master", "== dev-foo"}, false},
		{"conjunctive with more than one dev negation, and numeric constraint", "> 5", []string{"!= dev-master", "!= dev-foo", "> 5"}, true},
		{"conjunctive with more than one of the same dev negation", "!= dev-foo", []string{"!= dev-foo", "!= dev-foo"}, true},
		{"switches to conjunctive when excluding versions and complex", "!= 3-stable, <5 || >=6, <9", []string{"!= 3, <5", ">=6, <9"}, false},
		{"conjunctive with multiple numeric negations and a disjunctive exact match for dev versions", "== dev-foo || == dev-bar", []string{"!= 1.0", "!= 2.0", "==dev-foo || ==dev-bar"}, true},
	}

	for _, tc := range tests {
		actual, err := composeConstraintDomains(tc.toCompact, tc.conjunctive)
		if err != nil {
			t.Errorf("%s: composeConstraintDomains unexpected error: %v", tc.name, err)
			continue
		}

		if tc.expected == "" {
			if !domainUnionEmpty(actual) {
				t.Errorf("%s: expected composed constraints to match nothing", tc.name)
			}
			continue
		}

		expected, err := constraintUnionDomains(tc.expected)
		if err != nil {
			t.Errorf("%s: expected constraint %q unexpected error: %v", tc.name, tc.expected, err)
			continue
		}
		if !domainUnionsEquivalent(actual, expected) {
			t.Errorf("%s: composed constraints are not equivalent to %q", tc.name, tc.expected)
		}
	}
}

func TestIntervalsGetCases(t *testing.T) {
	noDev := branchSnapshot{Names: []string{}}
	tests := []struct {
		name       string
		constraint string
		expected   domainSnapshot
	}{
		{
			name:       "simple case",
			constraint: "^1.0",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.0.0.0-dev", End: "< 2.0.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:       "simple case/2",
			constraint: "> 1.0",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: "> 1.0.0.0", End: "< +Inf"}},
				Branches: noDev,
			},
		},
		{
			name:       "intervals should be sorted",
			constraint: "1.3.4 || 1.2.3 || >2.3,<2.5 || <1,>=0.9",
			expected: domainSnapshot{
				Numeric: []intervalSnapshot{
					{Start: ">= 0.9.0.0-dev", End: "< 1.0.0.0-dev"},
					{Start: ">= 1.2.3.0", End: "<= 1.2.3.0"},
					{Start: ">= 1.3.4.0", End: "<= 1.3.4.0"},
					{Start: "> 2.3.0.0", End: "< 2.5.0.0-dev"},
				},
				Branches: noDev,
			},
		},
		{
			name:       "intervals should be sorted and consecutive ones merged",
			constraint: "^4.0 || ^1.0 || ^3.0",
			expected: domainSnapshot{
				Numeric: []intervalSnapshot{
					{Start: ">= 1.0.0.0-dev", End: "< 2.0.0.0-dev"},
					{Start: ">= 3.0.0.0-dev", End: "< 5.0.0.0-dev"},
				},
				Branches: noDev,
			},
		},
		{
			name:       "consecutive intervals should be merged even if one has no end",
			constraint: "^4.0 || >= 5",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 4.0.0.0-dev", End: "< +Inf"}},
				Branches: noDev,
			},
		},
		{
			name:       "consecutive intervals should be merged even if one has no start",
			constraint: ">= 5,< 6 || < 5",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< 6.0.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:       "consecutive intervals representing everything should become any numeric",
			constraint: ">= 5 || < 5",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: noDev,
			},
		},
		{
			name:       "intervals should be sorted and overlapping ones merged",
			constraint: "^4.0 || ^1.1 || ^3.0 || ^1.2",
			expected: domainSnapshot{
				Numeric: []intervalSnapshot{
					{Start: ">= 1.1.0.0-dev", End: "< 2.0.0.0-dev"},
					{Start: ">= 3.0.0.0-dev", End: "< 5.0.0.0-dev"},
				},
				Branches: noDev,
			},
		},
		{
			name:       "intervals should be sorted and overlapping ones merged/2",
			constraint: "1.2 - 1.4 || 1.0 - 1.3",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.0.0.0-dev", End: "< 1.5.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:       "overlapping intervals should be merged even if the last has no end",
			constraint: "^4.0 || >= 4.5",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 4.0.0.0-dev", End: "< +Inf"}},
				Branches: noDev,
			},
		},
		{
			name:       "overlapping intervals should be merged even if the first has no start",
			constraint: ">= 5,< 6 || < 5.3",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< 6.0.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:       "overlapping intervals representing everything should become any numeric",
			constraint: ">= 5 || <= 5",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: noDev,
			},
		},
		{
			name:       "equal intervals should be merged",
			constraint: "^1.0 || ^1.0",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.0.0.0-dev", End: "< 2.0.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:       "weird input order should still be a good result",
			constraint: "< 2.0 || < 1.2",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< 2.0.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:       "weird input order should still be a good result, matches everything numeric",
			constraint: "< 2.0 || >= 1",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: noDev,
			},
		},
		{
			name:       "weird input order should still be a good result, conjunctive",
			constraint: "< 2.0, >= 1",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.0.0.0-dev", End: "< 2.0.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:       "conjunctive constraints result in no interval if conflicting",
			constraint: "^1.0, ^2.0",
			expected:   domainSnapshot{Branches: noDev},
		},
		{
			name:       "conjunctive constraints result in no interval if conflicting/2",
			constraint: "^1.0, ^3.0",
			expected:   domainSnapshot{Branches: noDev},
		},
		{
			name:       "conjunctive constraints result in no interval if conflicting/3",
			constraint: "== 1.0, != 1.0",
			expected:   domainSnapshot{Branches: noDev},
		},
		{
			name:       "conjunctive constraints result in no interval if conflicting/4",
			constraint: "> 1.0, dev-master",
			expected:   domainSnapshot{Branches: noDev},
		},
		{
			name:       "conjunctive constraints result in no branches interval if numeric is provided",
			constraint: "!= dev-master, != dev-foo, > 5",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: "> 5.0.0.0", End: "< +Inf"}},
				Branches: noDev,
			},
		},
		{
			name:       "conjunctive constraints result in no branches interval if numeric is provided, even if one matches dev",
			constraint: "!= 6, > 5",
			expected: domainSnapshot{
				Numeric: []intervalSnapshot{
					{Start: "> 5.0.0.0", End: "< 6.0.0.0"},
					{Start: "> 6.0.0.0", End: "< +Inf"},
				},
				Branches: noDev,
			},
		},
		{
			name:       "disjunctive constraints keeps branch intervals if numeric is provided",
			constraint: "!= dev-master, != dev-foo || > 5",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{"dev-foo", "dev-master"}, Exclude: true},
			},
		},
		{
			name:       "conjunctive constraints should be intersected",
			constraint: "^1.0, ^1.2",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.2.0.0-dev", End: "< 2.0.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:       "conjunctive constraints should be intersected/2",
			constraint: "^1.0, ^1.2, 1.4 - 1.8, 1.5 - 1.6, 1.5 - 2",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.5.0.0-dev", End: "< 1.7.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:       "conjunctive constraints simple",
			constraint: "1.5 - 2",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.5.0.0-dev", End: "< 3.0.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:       "conjunctive constraints with dev exclusions",
			constraint: "!= 1.4.5, ^1.0, != 1.2.3, != 2.3, != dev-foo, != dev-master",
			expected: domainSnapshot{
				Numeric: []intervalSnapshot{
					{Start: ">= 1.0.0.0-dev", End: "< 1.2.3.0"},
					{Start: "> 1.2.3.0", End: "< 1.4.5.0"},
					{Start: "> 1.4.5.0", End: "< 2.0.0.0-dev"},
				},
				Branches: noDev,
			},
		},
		{
			name:       "conjunctive constraints with dev exact versions suppresses the number scope matches",
			constraint: "!= 1.4.5, ^1.0, != 1.2.3, != 2.3, == dev-foo, == dev-foo",
			expected:   domainSnapshot{Branches: noDev},
		},
		{
			name:       "conjunctive constraints with dev exact versions suppresses number scope but keeps dev when allowed",
			constraint: "!= 1.2.3, != 2.3, == dev-foo, == dev-foo",
			expected:   domainSnapshot{Branches: branchSnapshot{Names: []string{"dev-foo"}}},
		},
		{
			name:       "disjunctive constraints with exclusions in dev constraints makes number scope match any",
			constraint: "^1.0 || != dev-foo",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{"dev-foo"}, Exclude: true},
			},
		},
		{
			name:       "disjunctive constraints with exclusions, if matches all numeric and dev, then any",
			constraint: "!= 1.4.5 || ^1.0 || != dev-foo || != dev-master || == dev-master",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{}, Exclude: true},
			},
		},
		{
			name:       "disjunctive constraints with exclusions, if dev constraints match all, then any",
			constraint: "^1.0 || != dev-master || == dev-master",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{}, Exclude: true},
			},
		},
		{
			name:       "disjunctive constraints with exclusions, dev scope excluded only",
			constraint: "^1.0 || != dev-foo || == dev-master",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{"dev-foo"}, Exclude: true},
			},
		},
		{
			name:       "disjunctive constraints with exact dev matches returns numeric and unique dev constraints",
			constraint: "^1.0 || == dev-foo || == dev-master || == dev-master",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.0.0.0-dev", End: "< 2.0.0.0-dev"}},
				Branches: branchSnapshot{Names: []string{"dev-foo", "dev-master"}},
			},
		},
		{
			name:       "conjunctive constraints with exact versions",
			constraint: "dev-master, ^1.0",
			expected:   domainSnapshot{Branches: noDev},
		},
		{
			name:       "conjunctive constraints with exact versions, dev only, diff version",
			constraint: "dev-master, dev-foo",
			expected:   domainSnapshot{Branches: noDev},
		},
		{
			name:       "conjunctive constraints with exact versions, dev only, same version",
			constraint: "dev-master, dev-master",
			expected:   domainSnapshot{Branches: branchSnapshot{Names: []string{"dev-master"}}},
		},
		{
			name:       "conjunctive constraints with same dev exclusion",
			constraint: "!= dev-master, != dev-master",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{"dev-master"}, Exclude: true},
			},
		},
		{
			name:       "conjunctive constraints with different dev exclusions",
			constraint: "!= dev-master, != dev-foo",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{"dev-foo", "dev-master"}, Exclude: true},
			},
		},
		{
			name:       "disjunctive constraints with exact versions",
			constraint: "dev-master || ^1.0 || dev-foo || dev-master",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.0.0.0-dev", End: "< 2.0.0.0-dev"}},
				Branches: branchSnapshot{Names: []string{"dev-foo", "dev-master"}},
			},
		},
		{
			name:       "conjunctive constraints with star should skip it",
			constraint: "^1.0, *",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.0.0.0-dev", End: "< 2.0.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:       "disjunctive constraints with star should result in any",
			constraint: "^1.0 || *",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{}, Exclude: true},
			},
		},
		{
			name:       "conjunctive constraints with only star should result in any",
			constraint: "*, *",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{}, Exclude: true},
			},
		},
		{
			name:       "disjunctive constraints with star and dev exclusion should not return exclusion",
			constraint: "!= dev-foo || *",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{}, Exclude: true},
			},
		},
		{
			name:       "conjunctive constraints with various dev constraints/2",
			constraint: "> 5, *",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: "> 5.0.0.0", End: "< +Inf"}},
				Branches: noDev,
			},
		},
		{
			name:       "conjunctive constraints with various dev constraints/3",
			constraint: "!= dev-foo, > 5",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: "> 5.0.0.0", End: "< +Inf"}},
				Branches: noDev,
			},
		},
		{
			name:       "conjunctive constraints with various dev constraints/4",
			constraint: "!= dev-foo, != dev-foo",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{"dev-foo"}, Exclude: true},
			},
		},
		{
			name:       "conjunctive constraints with various dev constraints/5",
			constraint: "!= dev-foo, != dev-bar",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{"dev-bar", "dev-foo"}, Exclude: true},
			},
		},
		{
			name:       "conjunctive constraints with various dev constraints/6",
			constraint: "!= dev-foo, == dev-bar",
			expected:   domainSnapshot{Branches: branchSnapshot{Names: []string{"dev-bar"}}},
		},
		{
			name:       "conjunctive constraints with various dev constraints/7",
			constraint: "dev-foo, > 5",
			expected:   domainSnapshot{Branches: noDev},
		},
		{
			name:       "complex conjunctive which can not match anything",
			constraint: "!= 1.0, != 1.0, != dev-foo, 1.0.5.*, == dev-bla",
			expected:   domainSnapshot{Branches: noDev},
		},
		{
			name:       "disjunctive constraints with various dev constraints",
			constraint: "!= dev-foo, != dev-bar || != dev-foo",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{"dev-foo"}, Exclude: true},
			},
		},
		{
			name:       "disjunctive constraints with various dev constraints/2",
			constraint: "!= dev-foo, != dev-bar || != dev-foo, != dev-bar",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{"dev-bar", "dev-foo"}, Exclude: true},
			},
		},
		{
			name:       "disjunctive constraints with various dev constraints/4",
			constraint: "== dev-foo || == dev-bar",
			expected:   domainSnapshot{Branches: branchSnapshot{Names: []string{"dev-bar", "dev-foo"}}},
		},
		{
			name:       "disjunctive constraints with various dev constraints/5",
			constraint: "== dev-foo || != dev-foo",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{}, Exclude: true},
			},
		},
		{
			name:       "disjunctive constraints with various dev constraints/6",
			constraint: "== dev-foo || != dev-bar",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{"dev-bar"}, Exclude: true},
			},
		},
		{
			name:       "disjunctive constraints with various dev constraints/7",
			constraint: "== dev-foo || != dev-bar || != dev-bar",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{"dev-bar"}, Exclude: true},
			},
		},
		{
			name:       "disjunctive constraints with various dev constraints/8",
			constraint: "== dev-foo || != dev-bar || != dev-foo",
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
				Branches: branchSnapshot{Names: []string{}, Exclude: true},
			},
		},
	}

	for _, tc := range tests {
		domains, err := constraintUnionDomains(tc.constraint)
		if err != nil {
			t.Errorf("%s: constraintUnionDomains(%q) unexpected error: %v", tc.name, tc.constraint, err)
			continue
		}
		actual := domainUnionSnapshot(domains)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("%s: expected %#v, got %#v", tc.name, tc.expected, actual)
		}
	}
}

func TestIntervalsGetComposedCases(t *testing.T) {
	noDev := branchSnapshot{Names: []string{}}
	tests := []struct {
		name        string
		parts       []string
		conjunctive bool
		expected    domainSnapshot
	}{
		{
			name: "conjunctive constraints should be intersected, not flattened by version parser",
			parts: []string{
				">=1.0.0.0-dev <2.0.0.0-dev",
				">=1.2.0.0-dev <2.0.0.0-dev",
				">=1.4.0.0-dev <1.9.0.0-dev",
				">=1.5.0.0-dev <1.7.0.0-dev",
				">=1.5.0.0-dev <3.0.0.0-dev",
			},
			conjunctive: true,
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.5.0.0-dev", End: "< 1.7.0.0-dev"}},
				Branches: noDev,
			},
		},
		{
			name:        "conjunctive constraints with disjunctive subcomponents should be intersected",
			parts:       []string{"1.0 - 1.2 || ^1.5", "1.8 - 1.9 || ^1.12"},
			conjunctive: true,
			expected: domainSnapshot{
				Numeric: []intervalSnapshot{
					{Start: ">= 1.8.0.0-dev", End: "< 1.10.0.0-dev"},
					{Start: ">= 1.12.0.0-dev", End: "< 2.0.0.0-dev"},
				},
				Branches: noDev,
			},
		},
		{
			name:        "conjunctive constraints with equal constraints",
			parts:       []string{"1.3.1.0-dev || 1.3.2.0-dev || 1.3.3.0-dev", "1.3.2.0-dev"},
			conjunctive: true,
			expected: domainSnapshot{
				Numeric:  []intervalSnapshot{{Start: ">= 1.3.2.0-dev", End: "<= 1.3.2.0-dev"}},
				Branches: noDev,
			},
		},
	}

	for _, tc := range tests {
		domains, err := composeConstraintDomains(tc.parts, tc.conjunctive)
		if err != nil {
			t.Errorf("%s: composeConstraintDomains unexpected error: %v", tc.name, err)
			continue
		}
		actual := domainUnionSnapshot(domains)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("%s: expected %#v, got %#v", tc.name, tc.expected, actual)
		}
	}
}

func TestMatcherAndSpecialConstraintBehavior(t *testing.T) {
	satisfies := []struct {
		version    string
		constraint string
		expected   bool
	}{
		{"2", ">=1", true},
		{"1.0", ">=2.11", false},
		{"11.0", ">=2.1", true},
		{"1.1", "*", true},
		{"1.1", "^1.0, ^2.0", false},
		{"2.0", "^1.0, ^2.0", false},
		{"dev-foo", "^1.0, ^2.0", false},
	}

	for _, tc := range satisfies {
		actual, err := checkSatisfiesForTest(tc.version, tc.constraint)
		if err != nil {
			t.Errorf("Satisfies(%q, %q) unexpected error: %v", tc.version, tc.constraint, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("Satisfies(%q, %q): expected %v, got %v", tc.version, tc.constraint, tc.expected, actual)
		}
	}

	if c := MustConstraints(NewConstraint("*")); c.String() != "*" {
		t.Errorf("match-all string: expected %q, got %q", "*", c.String())
	}
}

func TestMultiConstraintBehavior(t *testing.T) {
	intersections := []struct {
		left     string
		right    string
		expected bool
	}{
		{">1.0 <1.2", "1.1", true},
		{">1.0 <1.2", ">=1.1 <2.0", true},
		{">1.0 || <1.2", ">1.0 || <1.2", true},
		{">1.0 <1.2", "<1.0 || >2.0", false},
		{">1.0 <1.2", "1.2", false},
		{">dev-foo || >dev-bar", "1.1", false},
		{"!=dev-foo, !=dev-bar", "!=1.1", true},
		{"^7.0", ">=7.0.0-dev <8.0.0-dev", true},
		{"^7.2", ">=7.2.0-dev <8.0.0-dev", true},
		{"7.4.*", ">=7.4.0-dev <7.5.0-dev", true},
		{"7.2.* || 7.4.*", ">=7.2.0-dev <7.5.0-dev", true},
		{"^7.1.15 || ^7.2.3", "^7.2.2", true},
	}

	for _, tc := range intersections {
		actual, err := ConstraintIntersects(tc.left, tc.right)
		if err != nil {
			t.Errorf("ConstraintIntersects(%q, %q) unexpected error: %v", tc.left, tc.right, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("ConstraintIntersects(%q, %q): expected %v, got %v", tc.left, tc.right, tc.expected, actual)
		}
	}

	subsets := []struct {
		left     string
		right    string
		expected bool
	}{
		{"^2.5 || ^3.0", ">=2.5.0-dev <4.0.0-dev", true},
		{"^2.5 || ^3.0 || ^4.0", ">=2.5.0-dev <5.0.0-dev", true},
		{"~2.5.9 || ~2.6, >=2.6.2", "~2.5.9 || ~2.6, >=2.6.2", true},
		{"^0.2 || ^1.0", ">=0.2.0-dev <0.3.0-dev || >=1.0.0-dev <2.0.0-dev", true},
		{"^0.1 || ^1.0 || ^2.0", ">=0.1.0-dev <0.2.0-dev || >=1.0.0-dev <3.0.0-dev", true},
		{"^1.0 || 2.1 || ^3.0", "^1.0 || 2.1 || ^3.0", true},
	}

	for _, tc := range subsets {
		actual, err := ConstraintSubsetOf(tc.left, tc.right)
		if err != nil {
			t.Errorf("ConstraintSubsetOf(%q, %q) unexpected error: %v", tc.left, tc.right, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("ConstraintSubsetOf(%q, %q): expected %v, got %v", tc.left, tc.right, tc.expected, actual)
		}
	}
}

func TestMultiConstraintMatchAllCreate(t *testing.T) {
	all := domainSnapshot{
		Numeric:  []intervalSnapshot{{Start: ">= 0.0.0.0-dev", End: "< +Inf"}},
		Branches: branchSnapshot{Names: []string{}, Exclude: true},
	}

	empty, err := composeConstraintDomains(nil, true)
	if err != nil {
		t.Fatalf("composeConstraintDomains(nil, true) unexpected error: %v", err)
	}
	if actual := domainUnionSnapshot(empty); !reflect.DeepEqual(actual, all) {
		t.Fatalf("empty conjunctive multi: expected %#v, got %#v", all, actual)
	}

	conjunctive, err := composeConstraintDomains([]string{">=2.5.0.0-dev", "<=3.0.0.0-dev", "*"}, true)
	if err != nil {
		t.Fatalf("composeConstraintDomains conjunctive unexpected error: %v", err)
	}
	expectedConjunctive, err := constraintUnionDomains(">=2.5.0.0-dev <=3.0.0.0-dev")
	if err != nil {
		t.Fatalf("expected conjunctive constraint unexpected error: %v", err)
	}
	if !domainUnionsEquivalent(conjunctive, expectedConjunctive) {
		t.Errorf("match-all inside conjunctive multi should be skipped")
	}

	disjunctive, err := composeConstraintDomains([]string{">=2.5.0.0-dev", "*"}, false)
	if err != nil {
		t.Fatalf("composeConstraintDomains disjunctive unexpected error: %v", err)
	}
	if actual := domainUnionSnapshot(disjunctive); !reflect.DeepEqual(actual, all) {
		t.Errorf("match-all inside disjunctive multi: expected %#v, got %#v", all, actual)
	}
}

func TestMultiConstraintOptimizationCases(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		expected   string
	}{
		{"collapses contiguous", "^2.5 || ^3.0", ">=2.5.0-dev <4.0.0-dev"},
		{"collapses multiple contiguous", "^2.5 || ^3.0 || ^4.0", ">=2.5.0-dev <5.0.0-dev"},
		{"does not collapse when one side is more complex", "~2.5.9 || ~2.6, >=2.6.2", "~2.5.9 || ~2.6, >=2.6.2"},
		{"collapses only the simple contiguous tail", "^1.0 || ^2.0 !=2.0.1 || ^3.0 || ^4.0", "^1.0 || ^2.0 !=2.0.1 || >=3.0.0-dev <5.0.0-dev"},
		{"does not collapse contiguous ranges with exclusions", "^1.0 != 1.0.1 || ^2.0 !=2.0.1 || ^3.0 || ^4.0 != 4.0.1", "^1.0 !=1.0.1 || ^2.0 !=2.0.1 || ^3.0 || ^4.0 !=4.0.1"},
		{"does not collapse if another constraint also applies", "~0.1 || ~1.0 !=1.0.1", "~0.1 || ~1.0 !=1.0.1"},
		{"does not collapse non-contiguous caret ranges", "^0.2 || ^1.0", "^0.2 || ^1.0"},
		{"collapses following contiguous ranges after a gap", "^0.1 || ^1.0 || ^2.0", "^0.1 || >=1.0.0-dev <3.0.0-dev"},
		{"does not collapse exact constraint outside range", "^1.0 || 2.1 || ^3.0", "^1.0 || 2.1 || ^3.0"},
	}

	for _, tc := range tests {
		actual, err := constraintUnionDomains(tc.constraint)
		if err != nil {
			t.Errorf("%s: constraintUnionDomains(%q) unexpected error: %v", tc.name, tc.constraint, err)
			continue
		}
		expected, err := constraintUnionDomains(tc.expected)
		if err != nil {
			t.Errorf("%s: expected constraint %q unexpected error: %v", tc.name, tc.expected, err)
			continue
		}
		if !domainUnionsEquivalent(actual, expected) {
			t.Errorf("%s: expected %q to be equivalent to %q", tc.name, tc.constraint, tc.expected)
		}
	}
}

func TestSemverInvalidConstraint(t *testing.T) {
	tests := []string{
		"",
		"1.0.0-meh",
		">2.0,,<=3.0",
		">2.0 ,, <=3.0",
		">2.0 ||| <=3.0",
		",^1@dev || ^4@dev",
		",^1@dev",
		"|| ^1@dev",
		"^1@dev ||",
		"^1@dev ,",
		"^2.0.*",
		"^2.0.x",
		"^2.0.x-beta",
		"^2.*",
		"^2.x",
		"^2.x-beta",
		"^2.1.2.*",
		"^2.1.2.x",
		"^2.1.2.x-beta",
		"~2.0.*",
		"~2.0.x",
		"~2.0.x-beta",
		"~2.*",
		"~2.x",
		"~2.x-beta",
		"~2.1.2.*",
		"~2.1.2.x",
		"~2.1.2.x-beta",
		"1.x - 2.*",
		"2.x.x.x-dev - 3.x.x.x-dev",
		"^1.*-beta-dev",
		"^1. *-dev",
		"~1.*-beta-dev",
		"1.0.0-dev<1.0.5-dev",
		"*-dev",
		"~>1.2",
		"^",
		"^8 || ^",
		"~",
		"~1 ~",
	}

	for _, constraint := range tests {
		t.Run(constraint, func(t *testing.T) {
			if _, err := NewConstraint(constraint); err == nil {
				t.Fatalf("expected constraint %q to be rejected", constraint)
			}
		})
	}
}

func checkSatisfiesForTest(versionString, constraintString string) (bool, error) {
	constraint, err := NewConstraint(constraintString)
	if err != nil {
		return false, err
	}

	v, err := NewVersion(versionString)
	if err != nil {
		return false, err
	}

	return constraint.Check(v), nil
}

func satisfiedByForTest(constraintString string, versions []string) ([]string, error) {
	matches := make([]string, 0, len(versions))
	for _, versionString := range versions {
		ok, err := checkSatisfiesForTest(versionString, constraintString)
		if err != nil {
			return nil, err
		}
		if ok {
			matches = append(matches, versionString)
		}
	}
	return matches, nil
}

func compareForTest(version1, operator, version2 string) (bool, error) {
	v1, err := NewVersion(version1)
	if err != nil {
		return false, err
	}

	v2, err := NewVersion(version2)
	if err != nil {
		return false, err
	}

	switch operator {
	case "=", "==":
		return v1.Equal(v2), nil
	case "!=", "<>":
		return !v1.Equal(v2), nil
	case ">":
		return v1.GreaterThan(v2), nil
	case ">=":
		return v1.GreaterThanOrEqual(v2), nil
	case "<":
		return v1.LessThan(v2), nil
	case "<=":
		return v1.LessThanOrEqual(v2), nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

func mustVersions(t *testing.T, versions []string) []*Version {
	t.Helper()

	result := make([]*Version, len(versions))
	for i, versionString := range versions {
		v, err := NewVersion(versionString)
		if err != nil {
			t.Fatalf("NewVersion(%q) unexpected error: %v", versionString, err)
		}
		result[i] = v
	}
	return result
}

func versionOriginals(versions []*Version) []string {
	result := make([]string, len(versions))
	for i, v := range versions {
		result[i] = v.Original()
	}
	return result
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
