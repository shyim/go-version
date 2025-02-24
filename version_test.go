package version

import (
	"reflect"
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

func TestVersionParsing(t *testing.T) {
	tests := []struct {
		version    string
		expected   string
		shouldFail bool
	}{
		{"1.2.3", "1.2.3.0", false},
		{"v1.2.3", "1.2.3.0", false},
		{"1.2", "1.2.0.0", false},
		{"1", "1.0.0.0", false},
		{"1.2.3-beta", "1.2.3.0-beta", false},
		{"1.2.3+build", "1.2.3.0", false},
		{"1.2.3-beta+build", "1.2.3.0-beta", false},
		{"1.2.3.4", "1.2.3.4", false},
		{"1.2.3.4-beta", "1.2.3.4-beta", false},
		{"1.2.3.4+build", "1.2.3.4", false},
		{"v1.2.3.4-beta+build", "1.2.3.4-beta", false},
		{"1.2.3-beta.2", "1.2.3.0-beta2", false},
		{"1.2.3+build.123", "1.2.3.0", false},
		{"", "", true},
		{"invalid", "", true},
		{"1.invalid", "", true},
		{"1.2.invalid", "", true},
		{"1.2.3-", "", true},
		{"1.2.3+", "", true},
	}

	for _, tc := range tests {
		v, err := NewVersion(tc.version)
		if tc.shouldFail {
			if err == nil {
				t.Errorf("Expected error for version %s, got none", tc.version)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for version %s: %v", tc.version, err)
			continue
		}

		if v.String() != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, v.String())
		}
	}
}

func TestVersionMust(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected Must to panic with invalid version")
		}
	}()

	Must(NewVersion("invalid"))
}

func TestVersionSegments(t *testing.T) {
	tests := []struct {
		version          string
		expectedInt      []int
		expectedInt64    []int64
		expectedOriginal string
	}{
		{
			"1.2.3",
			[]int{1, 2, 3, 0},
			[]int64{1, 2, 3, 0},
			"1.2.3",
		},
		{
			"v1.2.3.4",
			[]int{1, 2, 3, 4},
			[]int64{1, 2, 3, 4},
			"v1.2.3.4",
		},
		{
			"1.2",
			[]int{1, 2, 0, 0},
			[]int64{1, 2, 0, 0},
			"1.2",
		},
		{
			"1",
			[]int{1, 0, 0, 0},
			[]int64{1, 0, 0, 0},
			"1",
		},
	}

	for _, tc := range tests {
		v := Must(NewVersion(tc.version))
		segments := v.Segments()
		segments64 := v.Segments64()
		original := v.Original()

		if len(segments) != len(tc.expectedInt) {
			t.Errorf("Expected %d segments, got %d for version %s", len(tc.expectedInt), len(segments), tc.version)
		}

		for i := 0; i < len(segments); i++ {
			if segments[i] != tc.expectedInt[i] {
				t.Errorf("Expected segment %d to be %d, got %d for version %s", i, tc.expectedInt[i], segments[i], tc.version)
			}
		}

		if len(segments64) != len(tc.expectedInt64) {
			t.Errorf("Expected %d segments64, got %d for version %s", len(tc.expectedInt64), len(segments64), tc.version)
		}

		for i := 0; i < len(segments64); i++ {
			if segments64[i] != tc.expectedInt64[i] {
				t.Errorf("Expected segment64 %d to be %d, got %d for version %s", i, tc.expectedInt64[i], segments64[i], tc.version)
			}
		}

		if original != tc.expectedOriginal {
			t.Errorf("Expected original %s, got %s", tc.expectedOriginal, original)
		}
	}
}

func TestVersionMetadata(t *testing.T) {
	tests := []struct {
		version          string
		expectedMetadata string
		expectedPre      string
		isPrerelease     bool
	}{
		{"1.2.3", "", "", false},
		{"1.2.3+build", "", "", false},
		{"1.2.3-beta", "", "beta", true},
		{"1.2.3-beta+build", "", "beta", true},
		{"1.2.3+build.1", "", "", false},
		{"1.2.3-beta1+build.1", "", "beta1", true},
		{"1.2.3-alpha1+build.123.4", "", "alpha1", true},
	}

	for _, tc := range tests {
		v := Must(NewVersion(tc.version))
		metadata := v.Metadata()
		prerelease := v.Prerelease()
		isPrerelease := v.IsPrerelease()

		if metadata != tc.expectedMetadata {
			t.Errorf("Expected metadata %s, got %s for version %s", tc.expectedMetadata, metadata, tc.version)
		}

		if prerelease != tc.expectedPre {
			t.Errorf("Expected prerelease %s, got %s for version %s", tc.expectedPre, prerelease, tc.version)
		}

		if isPrerelease != tc.isPrerelease {
			t.Errorf("Expected IsPrerelease() to be %v, got %v for version %s", tc.isPrerelease, isPrerelease, tc.version)
		}
	}
}

func TestVersionIncrementFunctions(t *testing.T) {
	tests := []struct {
		version       string
		afterMajor    string
		afterMinor    string
		afterPatch    string
		startSegments []int64
		majorSegments []int64
		minorSegments []int64
		patchSegments []int64
	}{
		{
			"1.2.3",
			"2.0.0.0",
			"1.3.0.0",
			"1.2.4.0",
			[]int64{1, 2, 3, 0},
			[]int64{2, 0, 0, 0},
			[]int64{1, 3, 0, 0},
			[]int64{1, 2, 4, 0},
		},
		{
			"1.2.3.4",
			"2.0.0.0",
			"1.3.0.0",
			"1.2.4.0",
			[]int64{1, 2, 3, 4},
			[]int64{2, 0, 0, 0},
			[]int64{1, 3, 0, 0},
			[]int64{1, 2, 4, 0},
		},
		{
			"0.1.2",
			"1.0.0.0",
			"0.2.0.0",
			"0.1.3.0",
			[]int64{0, 1, 2, 0},
			[]int64{1, 0, 0, 0},
			[]int64{0, 2, 0, 0},
			[]int64{0, 1, 3, 0},
		},
	}

	for _, tc := range tests {
		// Test IncreaseMajor
		v := Must(NewVersion(tc.version))
		if !reflect.DeepEqual(v.segments, tc.startSegments) {
			t.Errorf("Initial segments don't match for %s. Expected %v, got %v", tc.version, tc.startSegments, v.segments)
		}

		v.IncreaseMajor()
		if !reflect.DeepEqual(v.segments, tc.majorSegments) {
			t.Errorf("After IncreaseMajor segments don't match for %s. Expected %v, got %v", tc.version, tc.majorSegments, v.segments)
		}
		if v.String() != tc.afterMajor {
			t.Errorf("Expected %s after IncreaseMajor, got %s", tc.afterMajor, v.String())
		}

		// Test IncreaseMinor
		v = Must(NewVersion(tc.version))
		v.IncreaseMinor()
		if !reflect.DeepEqual(v.segments, tc.minorSegments) {
			t.Errorf("After IncreaseMinor segments don't match for %s. Expected %v, got %v", tc.version, tc.minorSegments, v.segments)
		}
		if v.String() != tc.afterMinor {
			t.Errorf("Expected %s after IncreaseMinor, got %s", tc.afterMinor, v.String())
		}

		// Test IncreasePatch
		v = Must(NewVersion(tc.version))
		v.IncreasePatch()
		if !reflect.DeepEqual(v.segments, tc.patchSegments) {
			t.Errorf("After IncreasePatch segments don't match for %s. Expected %v, got %v", tc.version, tc.patchSegments, v.segments)
		}
		if v.String() != tc.afterPatch {
			t.Errorf("Expected %s after IncreasePatch, got %s", tc.afterPatch, v.String())
		}
	}
}

func TestVersionComparePrerelease(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.2.3-alpha", "1.2.3-alpha", 0},
		{"1.2.3-alpha1", "1.2.3-alpha2", -1},
		{"1.2.3-alpha2", "1.2.3-alpha1", 1},
		{"1.2.3-alpha", "1.2.3-beta", -1},
		{"1.2.3-beta", "1.2.3-alpha", 1},
		{"1.2.3-alpha", "1.2.3", -1},
		{"1.2.3", "1.2.3-alpha", 1},
		{"1.2.3-rc1", "1.2.3-rc2", -1},
		{"1.2.3-rc2", "1.2.3-rc1", 1},
		{"1.2.3-rc1", "1.2.3-rc10", -1},
		{"1.2.3-rc10", "1.2.3-rc1", 1},
	}

	for _, tc := range tests {
		v1 := Must(NewVersion(tc.v1))
		v2 := Must(NewVersion(tc.v2))
		result := v1.Compare(v2)
		if result != tc.expected {
			t.Errorf("Comparing %s with %s: expected %d, got %d", tc.v1, tc.v2, tc.expected, result)
		}
	}
}

func TestVersionCompareSegments(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.3", "1.2.4", -1},
		{"1.2.4", "1.2.3", 1},
		{"1.2.3", "1.3.0", -1},
		{"1.3.0", "1.2.3", 1},
		{"1.2.3", "2.0.0", -1},
		{"2.0.0", "1.2.3", 1},
		{"1.2.3.0", "1.2.3", 0},
		{"1.2.3.1", "1.2.3", 1},
		{"1.2.3", "1.2.3.1", -1},
		{"1.2.3.4", "1.2.3.4", 0},
		{"1.2.3.4", "1.2.3.5", -1},
		{"1.2.3.5", "1.2.3.4", 1},
	}

	for _, tc := range tests {
		v1 := Must(NewVersion(tc.v1))
		v2 := Must(NewVersion(tc.v2))
		result := v1.Compare(v2)
		if result != tc.expected {
			t.Errorf("Comparing %s with %s: expected %d, got %d", tc.v1, tc.v2, tc.expected, result)
		}
	}
}
