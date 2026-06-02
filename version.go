package version

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// The compiled regular expression used to test the validity of a version.
var (
	versionRegexp *regexp.Regexp
)

// VersionRegexpRaw The raw regular expression string used for testing the validity
// of a version.
const (
	VersionRegexpRaw string = `[\^~v]?([0-9]+(\.[0-9]+)*?)` +
		`(-([0-9]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)|([-@]?([A-Za-z\-~]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)))?` +
		`(\+([0-9A-Za-z\-~]+(\.[0-9A-Za-z\-~]+)*))?` +
		`?`
)

// Version represents a single version.
type Version struct {
	pre      string
	segments []int64
	si       int
	original string
	branch   string
}

func init() {
	versionRegexp = regexp.MustCompile("^" + VersionRegexpRaw + "$")
}

// NewVersion parses the given version and returns a new
// Version.
func NewVersion(v string) (*Version, error) {
	return newVersionFromRegExp(v, versionRegexp)
}

func newVersionFromRegExp(v string, pattern *regexp.Regexp) (*Version, error) {
	normalized, err := normalizeVersion(v)

	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(strings.ToLower(normalized), "dev-") {
		return &Version{
			pre:      "dev",
			segments: []int64{0, 0, 0},
			original: v,
			branch:   normalized,
		}, nil
	}

	matches := pattern.FindStringSubmatch(normalized)
	if matches == nil {
		return nil, fmt.Errorf("malformed version: %s", v)
	}
	segmentsStr := strings.Split(matches[1], ".")
	segments := make([]int64, len(segmentsStr))
	si := 0
	for i, str := range segmentsStr {
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, fmt.Errorf(
				"error parsing version: %s", err)
		}

		segments[i] = val
		si++
	}

	// Even though we could support more than three segments, if we
	// got less than three, pad it with 0s. This is to cover the basic
	// default usecase of semver, which is MAJOR.MINOR.PATCH at the minimum
	for i := len(segments); i < 3; i++ {
		segments = append(segments, 0)
	}

	pre := matches[7]
	if pre == "" {
		pre = matches[4]
	}

	return &Version{
		pre:      pre,
		segments: segments,
		si:       si,
		original: v,
	}, nil
}

// Must is a helper that wraps a call to a function returning (*Version, error)
// and panics if error is non-nil.
func Must(v *Version, err error) *Version {
	if err != nil {
		panic(err)
	}

	return v
}

// Compare compares this version to another version. This
// returns -1, 0, or 1 if this version is smaller, equal,
// or larger than the other version, respectively.
//
// If you want boolean results, use the LessThan, Equal,
// GreaterThan, GreaterThanOrEqual or LessThanOrEqual methods.
func (v *Version) Compare(other *Version) int {
	// A quick, efficient equality check
	if v.NormalizedString() == other.NormalizedString() {
		return 0
	}

	if v.branch != "" || other.branch != "" {
		if v.branch == "" {
			return 1
		}
		if other.branch == "" {
			return -1
		}
		if v.branch < other.branch {
			return -1
		}
		return 1
	}

	segmentsSelf := v.Segments64()
	segmentsOther := other.Segments64()

	// If the segments are the same, we must compare on prerelease info
	if reflect.DeepEqual(segmentsSelf, segmentsOther) {
		preSelf := v.Prerelease()
		preOther := other.Prerelease()
		if preSelf == "" && preOther == "" {
			return 0
		}
		if preSelf == "" {
			if parsePrereleasePart(preOther).rank == prereleaseRankPatch {
				return -1
			}
			return 1
		}
		if preOther == "" {
			if parsePrereleasePart(preSelf).rank == prereleaseRankPatch {
				return 1
			}
			return -1
		}

		return comparePrereleases(preSelf, preOther)
	}

	// Get the highest specificity (hS), or if they're equal, just use segmentSelf length
	lenSelf := len(segmentsSelf)
	lenOther := len(segmentsOther)
	hS := lenSelf
	if lenSelf < lenOther {
		hS = lenOther
	}
	// Compare the segments
	// Because a constraint could have less specificity than the version it's
	// checking, we need to account for a lopsided or jagged comparison
	for i := 0; i < hS; i++ {
		if i > lenSelf-1 {
			// This means Self had the lower specificity
			// Check to see if the remaining segments in Other are all zeros
			if !allZero(segmentsOther[i:]) {
				// if not, it means that Other has to be greater than Self
				return -1
			}
			break
		} else if i > lenOther-1 {
			// this means Others had the lower specificity
			// Check to see if the remaining segments in Self are all zeros -
			if !allZero(segmentsSelf[i:]) {
				// if not, it means that Self has to be greater than Other
				return 1
			}
			break
		}
		lhs := segmentsSelf[i]
		rhs := segmentsOther[i]
		if lhs == rhs {
			continue
		} else if lhs < rhs {
			return -1
		}
		// Otherwise, rhs was > lhs, they're not equal
		return 1
	}

	// if we got this far, they're equal
	return 0
}

func allZero(segs []int64) bool {
	for _, s := range segs {
		if s != 0 {
			return false
		}
	}
	return true
}

type prereleasePart struct {
	rank        int
	suffix      string
	hasSuffix   bool
	suffixValue int64
	suffixNum   bool
}

const (
	prereleaseRankDev = iota
	prereleaseRankAlpha
	prereleaseRankBeta
	prereleaseRankRC
	prereleaseRankStable
	prereleaseRankPatch
	prereleaseRankOther
)

func parsePrereleasePart(part string) prereleasePart {
	lower := strings.ToLower(part)
	for _, prefix := range []struct {
		name string
		rank int
	}{
		{"dev", prereleaseRankDev},
		{"alpha", prereleaseRankAlpha},
		{"beta", prereleaseRankBeta},
		{"rc", prereleaseRankRC},
		{"stable", prereleaseRankStable},
		{"patch", prereleaseRankPatch},
	} {
		if strings.HasPrefix(lower, prefix.name) {
			suffix := part[len(prefix.name):]
			parsed := prereleasePart{
				rank:      prefix.rank,
				suffix:    suffix,
				hasSuffix: suffix != "",
			}
			if suffix != "" {
				if value, err := strconv.ParseInt(suffix, 10, 64); err == nil {
					parsed.suffixValue = value
					parsed.suffixNum = true
				}
			}
			return parsed
		}
	}

	parsed := prereleasePart{
		rank:      prereleaseRankOther,
		suffix:    part,
		hasSuffix: part != "",
	}
	if value, err := strconv.ParseInt(part, 10, 64); err == nil {
		parsed.rank = prereleaseRankStable
		parsed.suffixValue = value
		parsed.suffixNum = true
	}
	return parsed
}

func comparePart(preSelf string, preOther string) int {
	if preSelf == preOther {
		return 0
	}

	var selfInt int64
	selfNumeric := true
	selfInt, err := strconv.ParseInt(preSelf, 10, 64)
	if err != nil {
		selfNumeric = false
	}

	var otherInt int64
	otherNumeric := true
	otherInt, err = strconv.ParseInt(preOther, 10, 64)
	if err != nil {
		otherNumeric = false
	}

	// if a part is empty, we use the other to decide
	if preSelf == "" {
		if otherNumeric {
			return -1
		}
		return 1
	}

	if preOther == "" {
		if selfNumeric {
			return 1
		}
		return -1
	}

	selfPart := parsePrereleasePart(preSelf)
	otherPart := parsePrereleasePart(preOther)
	if selfPart.rank != otherPart.rank {
		if selfPart.rank < otherPart.rank {
			return -1
		}
		return 1
	}

	if selfPart.rank != prereleaseRankOther {
		if !selfPart.hasSuffix && !otherPart.hasSuffix {
			return 0
		}
		if !selfPart.hasSuffix {
			return -1
		}
		if !otherPart.hasSuffix {
			return 1
		}
		if selfPart.suffixNum && otherPart.suffixNum {
			if selfPart.suffixValue == otherPart.suffixValue {
				return 0
			}
			if selfPart.suffixValue < otherPart.suffixValue {
				return -1
			}
			return 1
		}
		if selfPart.suffix > otherPart.suffix {
			return 1
		}
		return -1
	}

	if selfNumeric && !otherNumeric {
		return -1
	} else if !selfNumeric && otherNumeric {
		return 1
	} else if !selfNumeric && !otherNumeric && preSelf > preOther {
		return 1
	} else if selfInt > otherInt {
		return 1
	}

	return -1
}

func comparePrereleases(v string, other string) int {
	// the same pre-release!
	if v == other {
		return 0
	}

	// split both pre-releases for analyse their parts
	selfPreReleaseMeta := strings.Split(v, ".")
	otherPreReleaseMeta := strings.Split(other, ".")

	selfPreReleaseLen := len(selfPreReleaseMeta)
	otherPreReleaseLen := len(otherPreReleaseMeta)

	biggestLen := otherPreReleaseLen
	if selfPreReleaseLen > otherPreReleaseLen {
		biggestLen = selfPreReleaseLen
	}

	// loop for parts to find the first difference
	for i := 0; i < biggestLen; i = i + 1 {
		partSelfPre := ""
		if i < selfPreReleaseLen {
			partSelfPre = selfPreReleaseMeta[i]
		}

		partOtherPre := ""
		if i < otherPreReleaseLen {
			partOtherPre = otherPreReleaseMeta[i]
		}

		compare := comparePart(partSelfPre, partOtherPre)
		// if parts are equals, continue the loop
		if compare != 0 {
			return compare
		}
	}

	return 0
}

// bothBranches reports whether both versions are dev-branch versions, which
// Composer treats as unordered: they only compare meaningfully for equality.
func (v *Version) bothBranches(o *Version) bool {
	return v.branch != "" && o.branch != ""
}

// Equal tests if two versions are equal.
func (v *Version) Equal(o *Version) bool {
	if v.bothBranches(o) {
		return v.branch == o.branch
	}

	return v.Compare(o) == 0
}

// GreaterThan tests if this version is greater than another version.
func (v *Version) GreaterThan(o *Version) bool {
	if v.bothBranches(o) {
		return false
	}

	return v.Compare(o) > 0
}

// GreaterThanOrEqual tests if this version is greater than or equal to another version.
func (v *Version) GreaterThanOrEqual(o *Version) bool {
	if v.bothBranches(o) {
		return v.branch == o.branch
	}

	return v.Compare(o) >= 0
}

// LessThan tests if this version is less than another version.
func (v *Version) LessThan(o *Version) bool {
	if v.bothBranches(o) {
		return false
	}

	return v.Compare(o) < 0
}

// LessThanOrEqual tests if this version is less than or equal to another version.
func (v *Version) LessThanOrEqual(o *Version) bool {
	if v.bothBranches(o) {
		return v.branch == o.branch
	}

	return v.Compare(o) <= 0
}

// Prerelease returns any prerelease data that is part of the version,
// or blank if there is no prerelease data.
//
// Prerelease information is anything that comes after the "-" in the
// version (but before any metadata). For example, with "1.2.3-beta",
// the prerelease information is "beta".
func (v *Version) Prerelease() string {
	return v.pre
}

// IsPrerelease returns true if the version has prerelease information.
func (v *Version) IsPrerelease() bool {
	return v.pre != ""
}

// Segments return the numeric segments of the version as a slice of ins.
//
// This excludes any metadata or pre-release information. For example,
// for a version "1.2.3-beta", segments will return a slice of
// 1, 2, 3.
func (v *Version) Segments() []int {
	segmentSlice := make([]int, len(v.segments))
	for i, v := range v.segments {
		segmentSlice[i] = int(v)
	}
	return segmentSlice
}

// Major returns the major version number.
func (v *Version) Major() int {
	return int(v.segments[0])
}

// Minor returns the minor version number.
func (v *Version) Minor() int {
	return int(v.segments[1])
}

// Patch returns the patch version number.
func (v *Version) Patch() int {
	return int(v.segments[2])
}

// Segments64 returns the numeric segments of the version as a slice of int64s.
//
// This excludes any metadata or pre-release information. For example,
// for a version "1.2.3-beta", segments will return a slice of
// 1, 2, 3.
func (v *Version) Segments64() []int64 {
	result := make([]int64, len(v.segments))
	copy(result, v.segments)
	return result
}

// IncreaseMajor increases the major version number by 1 and resets minor, patch and build to 0.
func (v *Version) IncreaseMajor() {
	v.segments[0]++
	v.segments[1] = 0
	v.segments[2] = 0
	v.segments[3] = 0
}

// IncreaseMinor increases the minor version number by 1 and resets patch and build to 0.
func (v *Version) IncreaseMinor() {
	v.segments[1]++
	v.segments[2] = 0
	v.segments[3] = 0
}

// IncreasePatch increases the patch version number by 1 and resets build to 0.
func (v *Version) IncreasePatch() {
	v.segments[2]++
	v.segments[3] = 0
}

// Original returns the original parsed version as-is, including any
// potential space, `v` prefix, etc.
func (v *Version) String() string {
	return v.original
}

// NormalizedString returns the canonicalized version string including any
// pre-release information. Build metadata is not retained, matching Composer,
// which discards everything after "+" during normalization.
//
// This value is rebuilt according to the parsed segments and other
// information. Therefore, ambiguities in the version string such as
// prefixed zeroes (1.04.0 => 1.4.0), `v` prefix (v1.0.0 => 1.0.0), and
// missing parts (1.0 => 1.0.0) will be made into a canonicalized form
// as shown in the parenthesized examples.
func (v *Version) NormalizedString() string {
	if v.branch != "" {
		return v.branch
	}

	var buf bytes.Buffer
	segments := v.segments
	if v.si > 0 && v.si <= len(v.segments) {
		segments = v.segments[:v.si]
	}
	fmtParts := make([]string, len(segments))
	for i, s := range segments {
		// We can ignore err here since we've pre-parsed the values in segments
		str := strconv.FormatInt(s, 10)
		fmtParts[i] = str
	}
	fmt.Fprint(&buf, strings.Join(fmtParts, "."))
	if v.pre != "" {
		fmt.Fprintf(&buf, "-%s", v.pre)
	}

	result := buf.String()
	// Re-run the canonical normalizer so the output is a fixed point: a bare
	// single-segment numeric form (e.g. parsed from "000000") would otherwise
	// render as "0" yet re-parse to the padded "0.0.0.0", breaking idempotency.
	// The normalizer is idempotent, so well-formed multi-segment results pass
	// through unchanged.
	if normalized, err := normalizeVersion(result); err == nil {
		return normalized
	}
	return result
}

// Original returns the original parsed version as-is, including any
// potential space, `v` prefix, etc.
func (v *Version) Original() string {
	return v.original
}
