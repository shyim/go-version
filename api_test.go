package version

import "testing"

func TestSatisfies(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		expected   bool
	}{
		{"6.4.28", ">=6.4.0,<6.4.29", true},
		{"6.4.29", ">=6.4.0,<6.4.29", false},
		{"1.2.3", "", true},
		{"1.2.3", "*", true},
	}

	for _, tc := range tests {
		actual, err := Satisfies(tc.version, tc.constraint)
		if err != nil {
			t.Errorf("Satisfies(%q, %q) unexpected error: %v", tc.version, tc.constraint, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("Satisfies(%q, %q): expected %v, got %v", tc.version, tc.constraint, tc.expected, actual)
		}
	}
}

func TestSatisfiesErrors(t *testing.T) {
	if _, err := Satisfies("not-a-version", "*"); err == nil {
		t.Fatal("expected invalid version error")
	}
	if _, err := Satisfies("1.2.3", ">>1.0"); err == nil {
		t.Fatal("expected malformed constraint error")
	}
}

func TestNormalizeVersion(t *testing.T) {
	normalized, err := normalizeVersion("v1.2.3")
	if err != nil {
		t.Fatalf("unexpected normalize error: %v", err)
	}
	if normalized != "1.2.3.0" {
		t.Fatalf("expected 1.2.3.0, got %s", normalized)
	}
}

func TestStability(t *testing.T) {
	tests := []struct {
		version  string
		expected string
	}{
		{"dev-main", StabilityDev},
		{"1.x-dev", StabilityDev},
		{"1.2.0-alpha1", StabilityAlpha},
		{"1.2.0-beta2", StabilityBeta},
		{"1.2.0-RC1", StabilityRC},
		{"1.2.0", StabilityStable},
	}

	for _, tc := range tests {
		actual := Stability(tc.version)
		if actual != tc.expected {
			t.Errorf("Stability(%q): expected %q, got %q", tc.version, tc.expected, actual)
		}
	}
}
