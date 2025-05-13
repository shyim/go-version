package version

import (
	"testing"
)

func TestWildcard(t *testing.T) {
	c := MustConstraints(NewConstraint("*"))
	if !c.Check(Must(NewVersion("1.0.0"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("1.0.0-alpha"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("1.0.0-beta"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("1.0.0-rc1"))) {
		t.Errorf("Expected true, got false")
	}
}

func TestWildcardInVersionMajor(t *testing.T) {
	c := MustConstraints(NewConstraint("2.*"))
	if !c.Check(Must(NewVersion("2.0.0"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("2.0.5"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("2.1.0"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("2.99.99"))) {
		t.Errorf("Expected true, got false")
	}
	if c.Check(Must(NewVersion("1.0.0"))) {
		t.Errorf("Expected false, got true")
	}
	if c.Check(Must(NewVersion("3.0.0"))) {
		t.Errorf("Expected false, got true")
	}
	// Test with pre-release versions
	if !c.Check(Must(NewVersion("2.0.0-alpha"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("2.1.0-beta"))) {
		t.Errorf("Expected true, got false")
	}
}

func TestWildcardInVersionMinor(t *testing.T) {
	c := MustConstraints(NewConstraint("2.0.*"))
	if !c.Check(Must(NewVersion("2.0.0"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("2.0.5"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("2.0.99"))) {
		t.Errorf("Expected true, got false")
	}
	if c.Check(Must(NewVersion("2.1.0"))) {
		t.Errorf("Expected false, got true")
	}
	if c.Check(Must(NewVersion("1.0.0"))) {
		t.Errorf("Expected false, got true")
	}
	if c.Check(Must(NewVersion("3.0.0"))) {
		t.Errorf("Expected false, got true")
	}
	// Test with pre-release versions
	if !c.Check(Must(NewVersion("2.0.0-alpha"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("2.0.5-beta"))) {
		t.Errorf("Expected true, got false")
	}
	if c.Check(Must(NewVersion("2.1.0-alpha"))) {
		t.Errorf("Expected false, got true")
	}
}

func TestConstraints(t *testing.T) {
	MustConstraints(NewConstraint(">=1.0.0"))
	MustConstraints(NewConstraint(">=1.0.0 || <2.0.0"))
	MustConstraints(NewConstraint(">=1.0.0,<2.0.0"))
}

func TestConstraintParsingWhitespaceAnd(t *testing.T) {
	c, err := NewConstraint(">=1.0 <2.0")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if c.String() != ">=1.0,<2.0" {
		t.Errorf("Expected '>=1.0,<2.0', got %s", c.String())
	}
	if !c.Check(Must(NewVersion("1.0.0"))) {
		t.Errorf("Expected true, got false")
	}
	if c.Check(Must(NewVersion("2.0.0"))) {
		t.Errorf("Expected false, got true")
	}
}

func TestConstraintParsingWhitespaceAndOr(t *testing.T) {
	c, err := NewConstraint("~6.4 >=6.4.20.0 || ~6.5")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if c.String() != "~6.4,>=6.4.20.0||~6.5" {
		t.Errorf("Expected '~6.4,>=6.4.20.0||~6.5', got %s", c.String())
	}
	if !c.Check(Must(NewVersion("6.4.20"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("6.4.20.0"))) {
		t.Errorf("Expected true, got false")
	}
	if !c.Check(Must(NewVersion("6.5.0"))) {
		t.Errorf("Expected true, got false")
	}
	if c.Check(Must(NewVersion("6.4.0.0"))) {
		t.Errorf("Expected false, got true")
	}
}

func TestConstraintWithoutWhitespace(t *testing.T) {
	c, err := NewConstraint("<6.6.1.0||>=6.3.5.0")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if c.String() != "<6.6.1.0||>=6.3.5.0" {
		t.Errorf("Expected '<6.6.1.0||>=6.3.5.0', got %s", c.String())
	}
	if !c.Check(Must(NewVersion("6.4.0.0"))) {
		t.Errorf("Expected true, got false")
	}
}

func TestConstraintVersionNumber(t *testing.T) {
	c, err := NewConstraint("1.0.0")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if c.String() != "1.0.0" {
		t.Errorf("Expected '1.0.0', got %s", c.String())
	}
	if !c.Check(Must(NewVersion("1.0.0"))) {
		t.Errorf("Expected true, got false")
	}
	if c.Check(Must(NewVersion("1.0.1"))) {
		t.Errorf("Expected false, got true")
	}
}

func TestConstraintMatchingWith4Digits(t *testing.T) {
	c, err := NewConstraint("~6.5.0.0")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	v := Must(NewVersion("6.5.0.0-rc1"))

	t.Logf("Constraint: %s, Version: %s", c.String(), v.String())
	t.Logf("Version segments: %v", v.Segments64())

	if !c.Check(v) {
		t.Errorf("Expected true, got false")
	}
}

func TestPrereleaseConstraints(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		expected   bool
	}{
		{">=1.0.0", "1.0.0-alpha", true},
		{">=1.0.0-alpha", "1.0.0", true},
		{"=1.0.0-alpha", "1.0.0-alpha", true},
		{"=1.0.0-alpha", "1.0.0-beta", false},
		{">=1.0.0-alpha", "1.0.0-beta", true},
		{"<=1.0.0", "1.0.0-alpha", true},
		{"^1.0.0", "1.1.0-alpha", true},
		{"~1.0.0", "1.0.1-beta", true},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("Failed to parse constraint %s: %v", tc.constraint, err)
			continue
		}

		v := Must(NewVersion(tc.version))
		actual := c.Check(v)
		if actual != tc.expected {
			t.Errorf("Constraint %s with version %s: expected %v, got %v", tc.constraint, tc.version, tc.expected, actual)
		}
	}
}

func TestCaretOperatorEdgeCases(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		expected   bool
	}{
		{"^0.1.0", "0.1.0", true},
		{"^0.1.0", "0.2.0", true},
		{"^1.0.0", "2.0.0", false},
		{"^1.0.0", "1.9.9", true},
		{"^1.0.0-alpha", "1.0.0-beta", true},
		{"^0.0.1", "0.0.2", true},
		{"^0.0.1", "0.0.1-alpha", true},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("Failed to parse constraint %s: %v", tc.constraint, err)
			continue
		}

		v := Must(NewVersion(tc.version))
		actual := c.Check(v)
		if actual != tc.expected {
			t.Errorf("Constraint %s with version %s: expected %v, got %v", tc.constraint, tc.version, tc.expected, actual)
		}
	}
}

func TestTildeOperatorEdgeCases(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		expected   bool
	}{
		{"~1.2", "1.2.0", true},
		{"~1.2", "1.3.0", false},
		{"~1.2.3", "1.2.4", true},
		{"~1.2.3", "1.3.0", false},
		{"~1.2.3-alpha", "1.2.3-beta", true},
		{"~0.1.2", "0.1.3", true},
		{"~0.1.2", "0.2.0", false},
		{"~0.1", "0.2.0", false},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("Failed to parse constraint %s: %v", tc.constraint, err)
			continue
		}

		v := Must(NewVersion(tc.version))
		actual := c.Check(v)
		if actual != tc.expected {
			t.Errorf("Constraint %s with version %s: expected %v, got %v", tc.constraint, tc.version, tc.expected, actual)
		}
	}
}
func TestMalformedConstraints(t *testing.T) {
	malformed := []string{
		">>1.0.0",
		"!1.0.0",
		"1.0.0-",
		"~>1.a.0",
		">=1.0.0-",
		"^1.0.0-",
		"~1.0.0-",
	}

	for _, c := range malformed {
		_, err := NewConstraint(c)
		if err == nil {
			t.Errorf("Expected error for malformed constraint: %s", c)
		}
	}
}

func TestMustConstraintsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected MustConstraints to panic with invalid constraint")
		}
	}()

	MustConstraints(NewConstraint(">>1.0.0"))
}

func TestConstraintString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{">=1.0.0", ">=1.0.0"},
		{"=1.0.0", "=1.0.0"},
		{"^1.0.0", "^1.0.0"},
		{"~1.0.0", "~1.0.0"},
		{">=1.0.0,<2.0.0", ">=1.0.0,<2.0.0"},
		{">=1.0.0 || <2.0.0", ">=1.0.0||<2.0.0"},
		{"v6.6.10.0", "v6.6.10.0"},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.input)
		if err != nil {
			t.Errorf("Failed to parse constraint %s: %v", tc.input, err)
			continue
		}

		if c.String() != tc.expected {
			t.Errorf("Expected String() to return %s, got %s", tc.expected, c.String())
		}
	}
}

func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		expected   bool
	}{
		{">1.0.0", "1.0.1", true},
		{">1.0.0", "1.0.0", false},
		{">1.0.0", "0.9.9", false},
		{"<1.0.0", "0.9.9", true},
		{"<1.0.0", "1.0.0", false},
		{"<1.0.0", "1.0.1", false},
		{"!=1.0.0", "1.0.1", true},
		{"<>1.0.0", "1.0.1", true},
		{"!=1.0.0", "1.0.0", false},
		{"<>1.0.0", "1.0.0", false},
		{">1.0.0-alpha", "1.0.0-beta", true},
		{">1.0.0-beta", "1.0.0-alpha", false},
		{"<1.0.0-alpha", "1.0.0-beta", false},
		{"<1.0.0-beta", "1.0.0-alpha", true},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("Failed to parse constraint %s: %v", tc.constraint, err)
			continue
		}

		v := Must(NewVersion(tc.version))
		actual := c.Check(v)
		if actual != tc.expected {
			t.Errorf("Constraint %s with version %s: expected %v, got %v", tc.constraint, tc.version, tc.expected, actual)
		}
	}
}

func TestComplexConstraints(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		expected   bool
	}{
		{">=1.0.0,<2.0.0", "1.5.0", true},
		{">=1.0.0,<2.0.0", "2.0.0", false},
		{">=1.0.0,<2.0.0", "0.9.9", false},
		{">=1.0.0 || >=3.0.0", "2.0.0", true},
		{">=1.0.0 || >=3.0.0", "3.0.0", true},
		{">=1.0.0 || >=3.0.0", "0.9.9", false},
		{"~1.2.3 || ^2.0.0", "1.2.4", true},
		{"~1.2.3 || ^2.0.0", "2.1.0", true},
		{"~1.2.3 || ^2.0.0", "1.3.0", false},
		{">=1.0.0-alpha,<2.0.0", "1.0.0-beta", true},
		{">=1.0.0-alpha,<2.0.0", "2.0.0-alpha", false},
		{">=1.0.0-alpha || >=2.0.0-alpha", "1.0.0-beta", true},
		{">=1.0.0-alpha || >=2.0.0-alpha", "2.0.0-beta", true},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("Failed to parse constraint %s: %v", tc.constraint, err)
			continue
		}

		v := Must(NewVersion(tc.version))
		actual := c.Check(v)
		if actual != tc.expected {
			t.Errorf("Constraint %s with version %s: expected %v, got %v", tc.constraint, tc.version, tc.expected, actual)
		}
	}
}

func TestConstraintPrereleaseFunction(t *testing.T) {
	tests := []struct {
		constraint string
		expected   bool
	}{
		{"1.0.0", false},
		{"1.0.0-alpha", true},
		{">=1.0.0", false},
		{">=1.0.0-alpha", true},
		{"~1.0.0-alpha", true},
		{"^1.0.0-alpha", true},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("Failed to parse constraint %s: %v", tc.constraint, err)
			continue
		}

		actual := c[0][0].Prerelease()
		if actual != tc.expected {
			t.Errorf("Constraint %s: expected Prerelease() to be %v, got %v", tc.constraint, tc.expected, actual)
		}
	}
}

func TestEqualConstraint(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		expected   bool
	}{
		{"==1.0.0", "1.0.0", true},
		{"==1.0.0", "1.0.1", false},
		{"==1.0.0", "0.9.9", false},
		{"==1.0.0-alpha", "1.0.0-alpha", true},
		{"==1.0.0-alpha", "1.0.0-beta", false},

		// Alias
		{"=1.0.0", "1.0.0", true},
		{"=1.0.0", "1.0.1", false},
		{"=1.0.0", "0.9.9", false},
		{"=1.0.0-alpha", "1.0.0-alpha", true},
		{"=1.0.0-alpha", "1.0.0-beta", false},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("Failed to parse constraint %s: %v", tc.constraint, err)
			continue
		}

		v := Must(NewVersion(tc.version))
		actual := c.Check(v)
		if actual != tc.expected {
			t.Errorf("Constraint %s with version %s: expected %v, got %v", tc.constraint, tc.version, tc.expected, actual)
		}
	}
}

func TestWildcardWithOperators(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		expected   bool
	}{
		// Greater than with wildcard
		{">= 2.*", "2.0.0", true},
		{">= 2.*", "2.1.0", true},
		{">= 2.*", "3.0.0", true},
		{">= 2.*", "1.9.9", false},

		// Less than with wildcard
		{"< 2.*", "1.9.9", true},
		{"< 2.*", "2.0.0", false},
		{"< 2.*", "2.1.0", false},

		// Multiple constraints with wildcard
		{">=2.* <4.*", "2.0.0", true},
		{">=2.* <4.*", "3.9.9", true},
		{">=2.* <4.*", "4.0.0", false},
		{">=2.* <4.*", "1.9.9", false},

		// Minor version wildcards with operators
		{">= 2.0.*", "2.0.0", true},
		{">= 2.0.*", "2.0.9", true},
		{">= 2.0.*", "2.1.0", true},
		{">= 2.0.*", "1.9.9", false},

		{"< 2.1.*", "2.0.9", true},
		{"< 2.1.*", "2.1.0", false},
		{"< 2.1.*", "2.2.0", false},

		// Complex constraints with wildcards
		{">=2.0.* <2.2.*", "2.0.0", true},
		{">=2.0.* <2.2.*", "2.1.9", true},
		{">=2.0.* <2.2.*", "2.2.0", false},
		{">=2.0.* <2.2.*", "1.9.9", false},

		// Pre-release versions with wildcards
		{"2.*", "2.0.0-alpha", true},
		{"2.*", "2.1.0-beta", true},
		{"2.0.*", "2.0.0-alpha", true},
		{"2.0.*", "2.0.5-beta", true},
		{"2.0.*", "2.1.0-alpha", false},
	}

	for _, tc := range tests {
		c := MustConstraints(NewConstraint(tc.constraint))
		actual := c.Check(Must(NewVersion(tc.version)))
		if actual != tc.expected {
			t.Errorf("Constraint '%s' with version '%s': expected %v, got %v", tc.constraint, tc.version, tc.expected, actual)
		}
	}
}

func TestMalformedWildcardConstraints(t *testing.T) {
	malformed := []string{
		"2.*.*",
		"*.0.0",
		"2.*.0",
		"*.*",
		"2.*.5",
		">=*",
		">=*.0",
	}

	for _, c := range malformed {
		_, err := NewConstraint(c)
		if err == nil {
			t.Errorf("Expected error for malformed wildcard constraint: %s", c)
		}
	}
}

func TestHyphenatedVersionRange(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		expected   bool
	}{
		{"1.0.0 - 2.0.0", "1.0.0-alpha", true},
		{"1.0.0 - 2.0.0", "2.0.0-alpha", true},
		{"1.0.0 - 2.0.0", "0.9.9", false},
		{"1.0.0 - 2.0.0", "3.0.0", false},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("Failed to parse constraint %s: %v", tc.constraint, err)
			continue
		}

		v := Must(NewVersion(tc.version))
		actual := c.Check(v)
		if actual != tc.expected {
			t.Errorf("Constraint %s with version %s: expected %v, got %v", tc.constraint, tc.version, tc.expected, actual)
		}
	}
}
