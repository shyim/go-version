package version

import (
	"reflect"
	"testing"
)

func TestConstraintIntersects(t *testing.T) {
	tests := []struct {
		left     string
		right    string
		expected bool
	}{
		{"^6.4", ">=6.4.0,<6.4.29", true},
		{"^6.4", ">=7.0.0,<7.1.0", false},
		{"^1.0", "^1.0", true},
		{"^1.0", "^2.0", false},
		{"^0.2.0", ">=0.2.1,<0.3.0", true},
		{"^0.2.0", ">=0.3.0", false},
		{"1.0 - 2.0", "2.0.*", true},
		{"1.0 - 2.0", "2.1.*", false},
		{"1.2.*", ">=1.2.5,<1.3.0", true},
		{"1.2.*", ">=1.3.0", false},
		{">1.0", "<2.0", true},
		{">=1.0,<2.0", ">=2.0", false},
		{"1.2.3", ">=1.0,<2.0", true},
		{"1.2.3", ">=2.0", false},
		{"dev-main", "dev-main", true},
		{"dev-main", "dev-feature", false},
		{"*", "dev-main", true},
		{"*", "^1.0", true},
		{"", "^1.0", true},
		{"!=1.2.3", "1.2.3", false},
		{"!=1.2.3", "1.2.4", true},
	}

	for _, tc := range tests {
		actual, err := ConstraintIntersects(tc.left, tc.right)
		if err != nil {
			t.Errorf("ConstraintIntersects(%q, %q) unexpected error: %v", tc.left, tc.right, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("ConstraintIntersects(%q, %q): expected %v, got %v", tc.left, tc.right, tc.expected, actual)
		}
	}
}

// TestConstraintIntersectsStability locks in Composer's behavior around
// @stability flags. Composer strips the @stability flag from a constraint and
// discards it entirely when it is "stable", so ">=1.0@stable" behaves exactly
// like ">=1.0". An explicit "-stable" version modifier is a different thing and
// is preserved, which raises the lower bound above same-version prereleases.
func TestConstraintIntersectsStability(t *testing.T) {
	tests := []struct {
		left     string
		right    string
		expected bool
	}{
		// @stable flag is ignored: identical to plain ">=1.0".
		{">=1.0@stable", "1.0-beta", true},
		{">=1.0", "1.0-beta", true},
		// Explicit -stable modifier is kept: excludes the same-version prerelease.
		{">=1.0-stable", "1.0-beta", false},
		// Other flags and ranges keep matching Composer.
		{">=1.0@stable", "1.1-beta", true},
		{">=1.0@stable", "2.0.0", true},
		{"^1.0@beta", "1.0-beta", true},
		{"<2.0@stable", "1.0-beta", true},
	}

	for _, tc := range tests {
		actual, err := ConstraintIntersects(tc.left, tc.right)
		if err != nil {
			t.Errorf("ConstraintIntersects(%q, %q) unexpected error: %v", tc.left, tc.right, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("ConstraintIntersects(%q, %q): expected %v, got %v", tc.left, tc.right, tc.expected, actual)
		}
	}
}

func TestConstraintSubsetOfStability(t *testing.T) {
	tests := []struct {
		left     string
		right    string
		expected bool
	}{
		// ">=1.0@stable" == ">=1.0", so each is a subset of the other.
		{">=1.0@stable", ">=1.0", true},
		{">=1.0", ">=1.0@stable", true},
		// ">=1.0-stable" excludes 1.0 prereleases, so it is a strict subset of
		// ">=1.0" but ">=1.0" is not a subset of it.
		{">=1.0-stable", ">=1.0", true},
		{">=1.0", ">=1.0-stable", false},
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

func TestConstraintIntersectsErrors(t *testing.T) {
	if _, err := ConstraintIntersects("^1.0", ">>1.0"); err == nil {
		t.Fatal("expected malformed constraint error")
	}
}

func TestConstraintSubsetOf(t *testing.T) {
	tests := []struct {
		left     string
		right    string
		expected bool
	}{
		{"^1.2", "^1.0 || ^2.0", true},
		{"^1.0 || ^2.0", "^1.0", false},
		{"*", ">=1 || <1", false},
		{"!= dev-foo", "!= dev-foo", true},
		{"!= dev-foo", "!= dev-bar", false},
		{"< dev-foo", "= dev-foo", true},
		{"^1.1, !=1.5.0", ">1.0.0", true},
		{">1.6", ">1.5, >1.4, !=1.7", false},
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

func TestComposeConstraintDomains(t *testing.T) {
	tests := []struct {
		name        string
		parts       []string
		conjunctive bool
		expected    string
		empty       bool
	}{
		{
			name:        "disjunctive union",
			parts:       []string{"1.0", "2.0"},
			conjunctive: false,
			expected:    "1.0 || 2.0",
		},
		{
			name:        "conjunctive intersection",
			parts:       []string{"^1.0", "^1.2"},
			conjunctive: true,
			expected:    "^1.2",
		},
		{
			name:        "conjunctive conflict",
			parts:       []string{"1.0", "2.0"},
			conjunctive: true,
			empty:       true,
		},
	}

	for _, tc := range tests {
		actual, err := composeConstraintDomains(tc.parts, tc.conjunctive)
		if err != nil {
			t.Errorf("%s: composeConstraintDomains unexpected error: %v", tc.name, err)
			continue
		}
		if tc.empty {
			if !domainUnionEmpty(actual) {
				t.Errorf("%s: expected empty composed domain", tc.name)
			}
			continue
		}
		expected, err := constraintUnionDomains(tc.expected)
		if err != nil {
			t.Errorf("%s: constraintUnionDomains unexpected error: %v", tc.name, err)
			continue
		}
		if !domainUnionsEquivalent(actual, expected) {
			t.Errorf("%s: expected composed domain equivalent to %q", tc.name, tc.expected)
		}
	}
}

func TestDomainUnionSnapshot(t *testing.T) {
	domains, err := constraintUnionDomains("!= 6, > 5")
	if err != nil {
		t.Fatalf("constraintUnionDomains unexpected error: %v", err)
	}

	actual := domainUnionSnapshot(domains)
	expected := domainSnapshot{
		Numeric: []intervalSnapshot{
			{Start: "> 5.0.0.0", End: "< 6.0.0.0"},
			{Start: "> 6.0.0.0", End: "< +Inf"},
		},
		Branches: branchSnapshot{Names: []string{}},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %#v, got %#v", expected, actual)
	}
}
