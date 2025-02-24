package version

import (
	"sort"
	"testing"
)

func TestBranchParsing(t *testing.T) {
	v := Must(NewVersion("6.5.x-dev"))

	if v.String() != "6.5.9999999.9999999-dev" {
		t.Errorf("Expected 6.5.9999999.9999999-dev, got %s", v.String())
	}

	c := MustConstraints(NewConstraint("~6.5.0"))

	if !c.Check(v) {
		t.Errorf("Expected true, got false")
	}
}

func TestMatchingRCWithTilde(t *testing.T) {
	vs := []*Version{
		Must(NewVersion("6.4.4.0")),
		Must(NewVersion("6.5.0.0-rc1")),
	}

	constraint, _ := NewConstraint("~6.5.0")

	match := ""

	for _, v := range vs {
		if constraint.Check(v) {
			match = v.String()
			break
		}
	}

	if match != "6.5.0.0-rc1" {
		t.Errorf("Expected 6.5.0.0-rc1 but got %s", match)
	}
}

func TestMatchingRCWithCaret(t *testing.T) {
	vs := []*Version{
		Must(NewVersion("6.4.4.0")),
		Must(NewVersion("6.5.0.0-rc1")),
	}

	constraint, _ := NewConstraint("^6.5")

	match := ""

	for _, v := range vs {
		if constraint.Check(v) {
			match = v.String()
			break
		}
	}

	if match != "6.5.0.0-rc1" {
		t.Errorf("Expected 6.5.0.0-rc1 but got %s", match)
	}
}

func TestMatchingRCWithCaretThreeNumbers(t *testing.T) {
	vs := []*Version{
		Must(NewVersion("6.4.4.0")),
		Must(NewVersion("6.5.0.0-rc1")),
	}

	constraint, _ := NewConstraint("^6.5.0")

	match := ""

	for _, v := range vs {
		if constraint.Check(v) {
			match = v.String()
			break
		}
	}

	if match != "6.5.0.0-rc1" {
		t.Errorf("Expected 6.5.0.0-rc1 but got %s", match)
	}
}

func TestMatchingRCWithGreaterThanEqual(t *testing.T) {
	vs := []*Version{
		Must(NewVersion("6.4.4.0")),
		Must(NewVersion("6.5.0.0-rc1")),
	}

	constraint, _ := NewConstraint(">=6.5")

	match := ""

	for _, v := range vs {
		if constraint.Check(v) {
			match = v.String()
			break
		}
	}

	if match != "6.5.0.0-rc1" {
		t.Errorf("Expected 6.5.0.0-rc1 but got %s", match)
	}
}

func TestCaretConstraint(t *testing.T) {
	constraint, _ := NewConstraint("^6.4.0")

	if constraint.Check(Must(NewVersion("6.3.0.0"))) {
		t.Errorf("Expected false, got true")
	}
	if !constraint.Check(Must(NewVersion("6.4.0.0"))) {
		t.Errorf("Expected true, got false")
	}
	if !constraint.Check(Must(NewVersion("6.4.0.1"))) {
		t.Errorf("Expected true, got false")
	}
	if !constraint.Check(Must(NewVersion("6.4.1.0"))) {
		t.Errorf("Expected true, got false")
	}
	if !constraint.Check(Must(NewVersion("6.4.5.0"))) {
		t.Errorf("Expected true, got false")
	}
	if !constraint.Check(Must(NewVersion("6.5.5.5"))) {
		t.Errorf("Expected true, got false")
	}
	if !constraint.Check(Must(NewVersion("6.9.9.9"))) {
		t.Errorf("Expected true, got false")
	}
	if constraint.Check(Must(NewVersion("7.0.0"))) {
		t.Errorf("Expected false, got true")
	}
}

func TestVersionWithoutOperator(t *testing.T) {
	constraint, err := NewConstraint("6.4.0.0")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if constraint.Check(Must(NewVersion("6.3.0.0"))) {
		t.Errorf("Expected false, got true")
	}
	if !constraint.Check(Must(NewVersion("6.4.0.0"))) {
		t.Errorf("Expected true, got false")
	}
	if constraint.Check(Must(NewVersion("6.5.0.0"))) {
		t.Errorf("Expected false, got true")
	}
}

func TestSortingVersions(t *testing.T) {
	vs := []*Version{
		Must(NewVersion("6.5.0.0-rc2")),
		Must(NewVersion("6.3.1.0")),
		Must(NewVersion("6.5.0.0-rc1")),
		Must(NewVersion("6.2.0")),
		Must(NewVersion("6.4.8.0")),
		Must(NewVersion("6.5.0.0")),
	}

	sort.Sort(Collection(vs))

	if vs[0].String() != "6.2.0.0" {
		t.Errorf("Expected 6.2.0.0, got %s", vs[0].String())
	}
	if vs[1].String() != "6.3.1.0" {
		t.Errorf("Expected 6.3.1.0, got %s", vs[1].String())
	}
	if vs[2].String() != "6.4.8.0" {
		t.Errorf("Expected 6.4.8.0, got %s", vs[2].String())
	}
	if vs[3].String() != "6.5.0.0-rc1" {
		t.Errorf("Expected 6.5.0.0-rc1, got %s", vs[3].String())
	}
	if vs[4].String() != "6.5.0.0-rc2" {
		t.Errorf("Expected 6.5.0.0-rc2, got %s", vs[4].String())
	}
	if vs[5].String() != "6.5.0.0" {
		t.Errorf("Expected 6.5.0.0, got %s", vs[5].String())
	}
}

func TestVersionIncrease(t *testing.T) {
	version := Must(NewVersion("1.2.3.0"))
	version.IncreasePatch()
	if version.String() != "1.2.4.0" {
		t.Errorf("Expected 1.2.4.0, got %s", version.String())
	}

	version.IncreaseMinor()
	if version.String() != "1.3.0.0" {
		t.Errorf("Expected 1.3.0.0, got %s", version.String())
	}

	version.IncreaseMajor()
	if version.String() != "2.0.0.0" {
		t.Errorf("Expected 2.0.0.0, got %s", version.String())
	}
}

func TestVersionString(t *testing.T) {
	cases := [][]string{
		{"1.2.3", "1.2.3.0"},
		{"1.2-beta", "1.2.0.0-beta"},
	}

	for _, tc := range cases {
		v, err := NewVersion(tc[0])
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v.String()
		expected := tc[1]
		if actual != expected {
			t.Fatalf("expected: %s\nactual: %s", expected, actual)
		}
		if actual := v.Original(); actual != tc[0] {
			t.Fatalf("expected original: %q\nactual: %q", tc[0], actual)
		}
	}
}

func TestEqual(t *testing.T) {
	cases := []struct {
		v1       string
		v2       string
		expected bool
	}{
		{"1.2.3", "1.4.5", false},
		{"1.2-beta", "1.2-beta", true},
		{"1.2", "1.1.4", false},
		{"1.2", "1.2-beta", false},
		{"1.2+foo", "1.2+beta", true},
		{"v1.2", "v1.2-beta", false},
		{"v1.2+foo", "v1.2+beta", true},
		{"v1.2.3.4", "v1.2.3.4", true},
		{"v1.2.0.0", "v1.2", true},
		{"v1.2", "v1.2.0.0", true},
		{"v1.2.0.0", "v1.2.0.1", false},
		{"v1.2.3.0", "v1.2.3.4", false},
		{"1.7-rc2", "1.7-rc1", false},
		{"1.7-rc2", "1.7", false},
	}

	for _, tc := range cases {
		v1, err := NewVersion(tc.v1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		v2, err := NewVersion(tc.v2)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v1.Equal(v2)
		expected := tc.expected
		if actual != expected {
			t.Fatalf(
				"%s <=> %s\nexpected: %t\nactual: %t",
				tc.v1, tc.v2,
				expected, actual)
		}
	}
}

func TestGreaterThan(t *testing.T) {
	cases := []struct {
		v1       string
		v2       string
		expected bool
	}{
		{"1.2.3", "1.4.5", false},
		{"1.2-beta", "1.2-beta", false},
		{"1.2", "1.1.4", true},
		{"1.2", "1.2-beta", true},
		{"1.2+foo", "1.2+beta", false},
		{"v1.2", "v1.2-beta", true},
		{"v1.2+foo", "v1.2+beta", false},
		{"v1.2.3.4", "v1.2.3.4", false},
		{"v1.2.0.0", "v1.2", false},
		{"v1.2.0.1", "v1.2", true},
		{"v1.2", "v1.2.0.0", false},
		{"v1.2", "v1.2.0.1", false},
		{"v1.2.0.0", "v1.2.0.1", false},
		{"v1.2.3.0", "v1.2.3.4", false},
		{"1.7-rc2", "1.7-rc1", true},
		{"1.7-rc2", "1.7", false},
	}

	for _, tc := range cases {
		v1, err := NewVersion(tc.v1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		v2, err := NewVersion(tc.v2)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v1.GreaterThan(v2)
		expected := tc.expected
		if actual != expected {
			t.Fatalf(
				"%s > %s\nexpected: %t\nactual: %t",
				tc.v1, tc.v2,
				expected, actual)
		}
	}
}

func TestLessThan(t *testing.T) {
	cases := []struct {
		v1       string
		v2       string
		expected bool
	}{
		{"1.2.3", "1.4.5", true},
		{"1.2-beta", "1.2-beta", false},
		{"1.2", "1.1.4", false},
		{"1.2", "1.2-beta", false},
		{"1.2+foo", "1.2+beta", false},
		{"v1.2", "v1.2-beta", false},
		{"v1.2+foo", "v1.2+beta", false},
		{"v1.2.3.4", "v1.2.3.4", false},
		{"v1.2.0.0", "v1.2", false},
		{"v1.2", "v1.2.0.0", false},
		{"v1.2", "v1.2.0.1", true},
		{"v1.2.0.0", "v1.2.0.1", true},
		{"v1.2.3.0", "v1.2.3.4", true},
		{"1.7-rc2", "1.7-rc1", false},
		{"1.7-rc2", "1.7", true},
	}

	for _, tc := range cases {
		v1, err := NewVersion(tc.v1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		v2, err := NewVersion(tc.v2)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v1.LessThan(v2)
		expected := tc.expected
		if actual != expected {
			t.Fatalf(
				"%s < %s\nexpected: %t\nactual: %t",
				tc.v1, tc.v2,
				expected, actual)
		}
	}
}

func TestGreaterThanOrEqual(t *testing.T) {
	cases := []struct {
		v1       string
		v2       string
		expected bool
	}{
		{"1.2.3", "1.4.5", false},
		{"1.2-beta", "1.2-beta", true},
		{"1.2", "1.1.4", true},
		{"1.2", "1.2-beta", true},
		{"1.2+foo", "1.2+beta", true},
		{"v1.2", "v1.2-beta", true},
		{"v1.2+foo", "v1.2+beta", true},
		{"v1.2.3.4", "v1.2.3.4", true},
		{"v1.2.0.0", "v1.2", true},
		{"v1.2.0.1", "v1.2", true},
		{"v1.2", "v1.2.0.0", true},
		{"v1.2", "v1.2.0.1", false},
		{"v1.2.0.0", "v1.2.0.1", false},
		{"v1.2.3.0", "v1.2.3.4", false},
		{"1.7-rc2", "1.7-rc1", true},
		{"1.7-rc2", "1.7", false},
	}

	for _, tc := range cases {
		v1, err := NewVersion(tc.v1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		v2, err := NewVersion(tc.v2)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v1.GreaterThanOrEqual(v2)
		expected := tc.expected
		if actual != expected {
			t.Fatalf(
				"%s >= %s\nexpected: %t\nactual: %t",
				tc.v1, tc.v2,
				expected, actual)
		}
	}
}

func TestLessThanOrEqual(t *testing.T) {
	cases := []struct {
		v1       string
		v2       string
		expected bool
	}{
		{"1.2.3", "1.4.5", true},
		{"1.2-beta", "1.2-beta", true},
		{"1.2", "1.1.4", false},
		{"1.2", "1.2-beta", false},
		{"1.2+foo", "1.2+beta", true},
		{"v1.2", "v1.2-beta", false},
		{"v1.2+foo", "v1.2+beta", true},
		{"v1.2.3.4", "v1.2.3.4", true},
		{"v1.2.0.0", "v1.2", true},
		{"v1.2", "v1.2.0.0", true},
		{"v1.2.0.1", "v1.2", false},
		{"v1.2", "v1.2.0.1", true},
		{"v1.2.0.0", "v1.2.0.1", true},
		{"v1.2.3.0", "v1.2.3.4", true},
		{"1.7-rc2", "1.7-rc1", false},
		{"1.7-rc2", "1.7", true},
	}

	for _, tc := range cases {
		v1, err := NewVersion(tc.v1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		v2, err := NewVersion(tc.v2)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v1.LessThanOrEqual(v2)
		expected := tc.expected
		if actual != expected {
			t.Fatalf(
				"%s <= %s\nexpected: %t\nactual: %t",
				tc.v1, tc.v2,
				expected, actual)
		}
	}
}

func TestConstraintPrerelease(t *testing.T) {
	cases := []struct {
		constraint string
		prerelease bool
	}{
		{"= 1.0", false},
		{"= 1.0-beta", true},
		{"~> 2.1.0", false},
		{"~> 2.1.0-dev", true},
		{"> 2.0", false},
		{">= 2.1.0-alpha", true},
	}

	for _, tc := range cases {
		c, err := parseSingle(tc.constraint)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := c.Prerelease()
		expected := tc.prerelease
		if actual != expected {
			t.Fatalf("Constraint: %s\nExpected: %#v",
				tc.constraint, expected)
		}
	}
}
