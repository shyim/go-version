package version

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Constraint represents a single constraint for a version, such as
// ">= 1.0".
type Constraint struct {
	f        constraintFunc
	check    *Version
	original string
}

// Constraints is a 2D slice of constraints. We make a custom type so
// that we can add methods to it.
type Constraints [][]*Constraint

type constraintFunc func(v, c *Version) bool

var (
	constraintRegexp    *regexp.Regexp
	constraintOperators map[string]constraintFunc
)

func init() {
	constraintOperators = map[string]constraintFunc{
		"":   constraintEqual,
		"=":  constraintEqual,
		"==": constraintEqual,
		"!=": constraintNotEqual,
		"<>": constraintNotEqual,
		">":  constraintGreaterThan,
		"<":  constraintLessThan,
		">=": constraintGreaterThanEqual,
		"<=": constraintLessThanEqual,
		"~>": constraintPessimistic,
		"^":  constraintCaret,
		"~":  constraintTilde,
		"*":  constraintWildcard,
	}

	ops := []string{
		"=",
		"==",
		"!=",
		"<>",
		">",
		"<",
		">=",
		"<=",
		"~>",
		"\\^",
		"~",
		"",
	}

	constraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`^\s*(%s)\s*([0-9*]+(?:\.[0-9*]+)*(?:-[0-9A-Za-z-.]+)?(?:\+[0-9A-Za-z-.]+)?)\s*$`,
		strings.Join(ops, "|")))
}

// NewConstraint will parse one or more constraints from the given
// constraint string. The string must be a comma or pipe separated
// list of constraints.
func NewConstraint(cs string) (Constraints, error) {
	cs = strings.ReplaceAll(cs, "||", "|")
	ors := strings.Split(cs, "|")
	or := make([][]*Constraint, len(ors))
	for k, v := range ors {
		// Check for hyphenated range
		if strings.Contains(v, " - ") {
			parts := strings.Split(v, " - ")
			if len(parts) != 2 {
				return nil, fmt.Errorf("malformed constraint: %s", v)
			}

			// Create >= for the lower bound
			lowerBound, err := parseSingle(">=" + strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, err
			}

			// Create <= for the upper bound
			upperBound, err := parseSingle("<=" + strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, err
			}

			or[k] = []*Constraint{lowerBound, upperBound}
			continue
		}

		// Normalize spaces between constraints to comma
		v = strings.TrimSpace(v)
		// Replace spaces between constraints with commas, but preserve spaces in operator-version pairs
		parts := strings.Fields(v)
		var normalized []string
		for i := 0; i < len(parts); i++ {
			if isOperator(parts[i]) && i+1 < len(parts) {
				normalized = append(normalized, parts[i]+parts[i+1])
				i++
			} else {
				normalized = append(normalized, parts[i])
			}
		}
		v = strings.Join(normalized, ",")

		vs := strings.Split(v, ",")
		result := make([]*Constraint, len(vs))
		for i, single := range vs {
			c, err := parseSingle(single)
			if err != nil {
				return nil, err
			}

			result[i] = c
		}
		or[k] = result
	}

	return Constraints(or), nil
}

// isOperator checks if a string is a version constraint operator
func isOperator(s string) bool {
	_, ok := constraintOperators[s]
	return ok || s == ">" || s == "<" || s == ">=" || s == "<=" || s == "!=" || s == "==" || s == "~>" || s == "~"
}

// MustConstraints is a helper that wraps a call to a function
// returning (Constraints, error) and panics if error is non-nil.
func MustConstraints(c Constraints, err error) Constraints {
	if err != nil {
		panic(err)
	}

	return c
}

// Check tests if a version satisfies all the constraints.
func (cs Constraints) Check(v *Version) bool {
	for _, o := range cs {
		ok := true
		for _, c := range o {
			if !c.Check(v) {
				ok = false
				break
			}
		}

		if ok {
			return true
		}
	}

	return false
}

// Prerelease returns true if the version underlying this constraint
// contains a prerelease field.
func (c *Constraint) Prerelease() bool {
	return len(c.check.Prerelease()) > 0
}

// Returns the string format of the constraints
func (cs Constraints) String() string {
	orStr := make([]string, len(cs))
	for i, o := range cs {
		csStr := make([]string, len(o))
		for j, c := range o {
			csStr[j] = c.String()
		}

		orStr[i] = strings.Join(csStr, ",")
	}

	return strings.Join(orStr, "||")
}

// Check tests if a constraint is validated by the given version.
func (c *Constraint) Check(v *Version) bool {
	return c.f(v, c.check)
}

func (c *Constraint) String() string {
	return c.original
}

func parseSingle(v string) (*Constraint, error) {
	if strings.TrimSpace(v) == "*" {
		return &Constraint{
			f:        constraintWildcard,
			check:    nil,
			original: v,
		}, nil
	}

	matches := constraintRegexp.FindStringSubmatch(v)
	if matches == nil {
		return nil, fmt.Errorf("malformed constraint: %s", v)
	}

	operator := matches[1]
	version := matches[2]

	// Handle wildcards in version numbers
	if strings.Contains(version, "*") {
		return parseWildcardConstraint(operator, version, v)
	}

	check, err := NewVersion(matches[2])
	if err != nil {
		return nil, err
	}

	return &Constraint{
		f:        constraintOperators[matches[1]],
		check:    check,
		original: v,
	}, nil
}

// parseWildcardConstraint handles parsing of version constraints containing wildcards.
// It validates the wildcard pattern and creates appropriate constraint functions.
func parseWildcardConstraint(operator, version, original string) (*Constraint, error) {
	parts := strings.Split(version, ".")

	// Check for malformed wildcard patterns
	starCount := 0
	for i, part := range parts {
		if part == "*" {
			starCount++
			// Wildcard can only appear at the end
			if i < len(parts)-1 {
				return nil, fmt.Errorf("malformed constraint: %s", original)
			}
		}
	}
	if starCount > 1 {
		return nil, fmt.Errorf("malformed constraint: %s", original)
	}

	if len(parts) >= 2 && parts[1] == "*" {
		// Convert 2.* to check for major version match
		majorVersion := parts[0]
		if strings.Contains(majorVersion, "*") {
			return nil, fmt.Errorf("malformed constraint: %s", original)
		}
		check, err := NewVersion(majorVersion + ".0.0")
		if err != nil {
			return nil, err
		}

		return &Constraint{
			f: func(v, c *Version) bool {
				switch operator {
				case ">=":
					return v.segments[0] >= c.segments[0]
				case ">":
					return v.segments[0] > c.segments[0]
				case "<=":
					return v.segments[0] <= c.segments[0]
				case "<":
					return v.segments[0] < c.segments[0]
				case "", "=", "==":
					return v.segments[0] == c.segments[0]
				case "!=", "<>":
					return v.segments[0] != c.segments[0]
				default:
					return v.segments[0] == c.segments[0]
				}
			},
			check:    check,
			original: original,
		}, nil
	} else if len(parts) >= 3 && parts[2] == "*" {
		// Convert 2.0.* to check for major.minor version match
		majorMinor := parts[0] + "." + parts[1]
		if strings.Contains(majorMinor, "*") {
			return nil, fmt.Errorf("malformed constraint: %s", original)
		}
		check, err := NewVersion(majorMinor + ".0")
		if err != nil {
			return nil, err
		}

		return &Constraint{
			f: func(v, c *Version) bool {
				// First check major version
				if v.segments[0] != c.segments[0] {
					switch operator {
					case ">=":
						return v.segments[0] > c.segments[0]
					case ">":
						return v.segments[0] > c.segments[0]
					case "<=":
						return v.segments[0] < c.segments[0]
					case "<":
						return v.segments[0] < c.segments[0]
					case "", "=", "==":
						return false
					case "!=", "<>":
						return true
					default:
						return false
					}
				}

				// If major version matches, check minor version
				switch operator {
				case ">=":
					return v.segments[1] >= c.segments[1]
				case ">":
					return v.segments[1] > c.segments[1]
				case "<=":
					return v.segments[1] <= c.segments[1]
				case "<":
					return v.segments[1] < c.segments[1]
				case "", "=", "==":
					return v.segments[1] == c.segments[1]
				case "!=", "<>":
					return v.segments[1] != c.segments[1]
				default:
					return v.segments[1] == c.segments[1]
				}
			},
			check:    check,
			original: original,
		}, nil
	}

	return nil, fmt.Errorf("malformed constraint: %s", original)
}

func prereleaseCheck(v, c *Version) bool {
	switch vPre, cPre := v.Prerelease() != "", c.Prerelease() != ""; {
	case cPre && vPre:
		// A constraint with a pre-release can only match a pre-release version
		// with the same base segments.
		return reflect.DeepEqual(c.Segments64(), v.Segments64())

	case !cPre && vPre:
		// A constraint without a pre-release can match a version with a
		// pre-release for tilde and caret operators
		return true

	case cPre && !vPre:
		// A constraint with a pre-release cannot match a version without a
		// pre-release
		return false

	case !cPre && !vPre:
		// Neither has prerelease
		return true
	}
	return true
}

//-------------------------------------------------------------------
// Constraint functions
//-------------------------------------------------------------------

func constraintEqual(v, c *Version) bool {
	return v.Equal(c)
}

func constraintNotEqual(v, c *Version) bool {
	return !v.Equal(c)
}

func constraintGreaterThan(v, c *Version) bool {
	return (bothNotPreRelease(v, c) || prereleaseCheck(v, c)) && v.Compare(c) == 1
}

func constraintLessThan(v, c *Version) bool {
	return (bothNotPreRelease(v, c) || prereleaseCheck(v, c)) && v.Compare(c) == -1
}

func constraintGreaterThanEqual(v, c *Version) bool {
	// For pre-release versions, we need to check if they match the constraint
	if v.IsPrerelease() && !c.IsPrerelease() {
		// Compare without pre-release info first
		vNoPrerelease := &Version{
			segments: v.segments,
			si:       v.si,
		}
		cNoPrerelease := &Version{
			segments: c.segments,
			si:       c.si,
		}
		return vNoPrerelease.Compare(cNoPrerelease) >= 0
	}

	return (bothNotPreRelease(v, c) || prereleaseCheck(v, c)) && v.Compare(c) >= 0
}

func constraintLessThanEqual(v, c *Version) bool {
	return (bothNotPreRelease(v, c) || prereleaseCheck(v, c)) && v.Compare(c) <= 0
}

func constraintPessimistic(v, c *Version) bool {
	// Using a pessimistic constraint with a pre-release, restricts versions to pre-releases
	if !prereleaseCheck(v, c) || (c.Prerelease() != "" && v.Prerelease() == "") {
		return false
	}

	// If the version being checked is naturally less than the constraint, then there
	// is no way for the version to be valid against the constraint
	if v.LessThan(c) {
		return false
	}
	// We'll use this more than once, so grab the length now so it's a little cleaner
	// to write the later checks
	cs := len(c.segments)

	// If the version being checked has less specificity than the constraint, then there
	// is no way for the version to be valid against the constraint
	if cs > len(v.segments) {
		return false
	}

	// Check the segments in the constraint against those in the version. If the version
	// being checked, at any point, does not have the same values in each index of the
	// constraints segments, then it cannot be valid against the constraint.
	for i := 0; i < c.si-1; i++ {
		if v.segments[i] != c.segments[i] {
			return false
		}
	}

	// Check the last part of the segment in the constraint. If the version segment at
	// this index is less than the constraints segment at this index, then it cannot
	// be valid against the constraint
	return c.segments[cs-1] <= v.segments[cs-1]
}

func constraintCaret(v, c *Version) bool {
	// For pre-release versions, we need to check if they match the constraint
	if v.IsPrerelease() && !c.IsPrerelease() {
		// Compare without pre-release info first
		vNoPrerelease := &Version{
			segments: v.segments,
			si:       v.si,
		}
		cNoPrerelease := &Version{
			segments: c.segments,
			si:       c.si,
		}
		if vNoPrerelease.LessThan(cNoPrerelease) {
			return false
		}
		if vNoPrerelease.segments[0] != cNoPrerelease.segments[0] {
			return false
		}
		return true
	}

	// If the version being checked is naturally less than the constraint, then there
	// is no way for the version to be valid against the constraint
	if v.LessThan(c) {
		return false
	}

	// Check the major version
	if v.segments[0] != c.segments[0] {
		return false
	}

	return true
}

func constraintTilde(v, c *Version) bool {
	// For tilde operator with prerelease versions, we need to compare without the prerelease tag first
	vNoPrerelease := &Version{
		segments: v.segments,
		si:       v.si,
	}
	cNoPrerelease := &Version{
		segments: c.segments,
		si:       c.si,
	}

	// If the version without prerelease is less than the constraint without prerelease,
	// then there is no way for the version to be valid against the constraint
	if vNoPrerelease.LessThan(cNoPrerelease) {
		return false
	}

	// Check the major version
	if v.segments[0] != c.segments[0] {
		return false
	}

	// Check the minor version if specified in the constraint
	if c.si > 1 && v.segments[1] != c.segments[1] {
		return false
	}

	// For tilde operator, we allow any prerelease version
	// as long as the major and minor versions match
	if v.IsPrerelease() && !c.IsPrerelease() {
		return true
	}

	// For exact version matches, we need to check prereleases
	if c.si == len(v.segments) && reflect.DeepEqual(v.segments[:c.si], c.segments[:c.si]) {
		return prereleaseCheck(v, c)
	}

	return true
}

func constraintWildcard(v, c *Version) bool {
	return true
}

func bothNotPreRelease(v, c *Version) bool {
	return !v.IsPrerelease() || !c.IsPrerelease()
}
