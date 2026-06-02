package version

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Constraint represents a single constraint for a version, such as
// ">= 1.0".
type Constraint struct {
	f            constraintFunc
	check        *Version
	original     string
	stability    string
	origSegments int    // Number of segments in the original constraint string
	operator     string // The operator used (e.g., "~", "^", ">=", etc.)
	stableBound  bool
}

// Constraints is a 2D slice of constraints. We make a custom type so
// that we can add methods to it.
type Constraints [][]*Constraint

type constraintFunc func(v, c *Version, origSegments int) bool

var (
	constraintOperators map[string]constraintFunc
	stabilityLevels     map[string]int
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
		"^":  constraintCaret,
		"~":  constraintTilde,
		"*":  constraintWildcard,
	}

	stabilityLevels = map[string]int{
		"dev":    0,
		"alpha":  1,
		"beta":   2,
		"rc":     3,
		"stable": 4,
	}

}

// NewConstraint will parse one or more constraints from the given
// constraint string. The string must be a comma or pipe separated
// list of constraints.
func NewConstraint(cs string) (Constraints, error) {
	cs = strings.ReplaceAll(cs, "||", "|")
	ors := strings.Split(cs, "|")
	or := make([][]*Constraint, len(ors))
	for k, v := range ors {
		v = stripConstraintAlias(v)
		// Check for hyphenated range
		if strings.Contains(v, " - ") && !strings.Contains(v, ",") {
			hyphenConstraints, err := parseHyphenRange(v)
			if err != nil {
				return nil, err
			}
			or[k] = hyphenConstraints
			continue
		}

		v = strings.TrimSpace(v)
		vs, err := splitAndConstraints(v)
		if err != nil {
			return nil, err
		}
		result := make([]*Constraint, 0, len(vs))
		for _, single := range vs {
			if strings.Contains(single, " - ") {
				hyphenConstraints, err := parseHyphenRange(single)
				if err != nil {
					return nil, err
				}
				result = append(result, hyphenConstraints...)
				continue
			}

			c, err := parseSingle(single)
			if err != nil {
				return nil, err
			}

			result = append(result, c)
		}
		or[k] = result
	}

	return Constraints(or), nil
}

func parseHyphenRange(v string) ([]*Constraint, error) {
	parts := strings.Split(v, " - ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("malformed constraint: %s", v)
	}
	if invalidHyphenWildcard(strings.TrimSpace(parts[0])) || invalidHyphenWildcard(strings.TrimSpace(parts[1])) {
		return nil, fmt.Errorf("malformed constraint: %s", v)
	}

	lowerBound, err := parseSingle(">=" + strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, err
	}

	upperBound, err := parseHyphenUpperBound(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, err
	}

	return []*Constraint{lowerBound, upperBound}, nil
}

func splitAndConstraints(v string) ([]string, error) {
	commaParts := strings.Split(v, ",")
	var result []string
	for _, commaPart := range commaParts {
		commaPart = strings.TrimSpace(commaPart)
		if commaPart == "" {
			return nil, fmt.Errorf("malformed constraint: %s", v)
		}
		if strings.Contains(commaPart, " - ") {
			result = append(result, commaPart)
			continue
		}

		fields := strings.Fields(commaPart)
		for i := 0; i < len(fields); i++ {
			if isOperator(fields[i]) && i+1 < len(fields) {
				result = append(result, fields[i]+fields[i+1])
				i++
			} else {
				result = append(result, fields[i])
			}
		}
	}
	return result, nil
}

func stripConstraintAlias(constraint string) string {
	lower := strings.ToLower(constraint)
	if index := strings.Index(lower, " as "); index >= 0 {
		return strings.TrimSpace(constraint[:index])
	}
	return constraint
}

func parseHyphenUpperBound(v string) (*Constraint, error) {
	parts, ok := simpleVersionParts(v)
	if !ok || len(parts) >= 3 {
		return parseSingle("<=" + v)
	}

	upper, err := incrementVersionParts(parts)
	if err != nil {
		return nil, err
	}

	return parseSingle("<" + upper)
}

func invalidHyphenWildcard(v string) bool {
	return isWildcardConstraintVersion(v) && !isNumericDevBranch(v)
}

func simpleVersionParts(v string) ([]string, bool) {
	v = strings.TrimSpace(v)
	if len(v) > 1 && (v[0] == 'v' || v[0] == 'V') {
		v = v[1:]
	}
	if v == "" || strings.ContainsAny(v, "-+@*") {
		return nil, false
	}

	parts := strings.Split(v, ".")
	for _, part := range parts {
		if part == "" {
			return nil, false
		}
		if _, err := strconv.ParseInt(part, 10, 64); err != nil {
			return nil, false
		}
	}

	return parts, true
}

func incrementVersionParts(parts []string) (string, error) {
	segments := make([]int64, len(parts))
	for i, part := range parts {
		segment, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return "", err
		}
		segments[i] = segment
	}

	if len(segments) == 1 {
		return fmt.Sprintf("%d.0.0", segments[0]+1), nil
	}

	return fmt.Sprintf("%d.%d.0", segments[0], segments[1]+1), nil
}

// isOperator checks if a string is a version constraint operator
func isOperator(s string) bool {
	_, ok := constraintOperators[s]
	return ok || s == ">" || s == "<" || s == ">=" || s == "<=" || s == "!=" || s == "==" || s == "~"
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
	if c.stability != "" {
		if stabilityLevels[strings.ToLower(c.stability)] > stabilityLevels[getVersionStability(v)] {
			return false
		}
	}
	if c.check != nil && (v.branch != "" || c.check.branch != "") {
		// Branch versions (dev-*) have no numeric ordering, so Composer only
		// gives them meaning under the equality operators; an ordering operator
		// applied to a branch matches nothing.
		switch c.operator {
		case "", "=", "==":
			return v.Equal(c.check)
		case "!=", "<>":
			return !v.Equal(c.check)
		default:
			return false
		}
	}
	if c.stableBound && c.excludesSameVersionPrerelease() && v.IsPrerelease() && reflect.DeepEqual(v.Segments64(), c.check.Segments64()) {
		return false
	}
	return c.f(v, c.check, c.origSegments)
}

func (c *Constraint) excludesSameVersionPrerelease() bool {
	switch c.operator {
	case "", "=", "==", ">", ">=", "^", "~":
		return true
	default:
		return false
	}
}

func (c *Constraint) String() string {
	return c.original
}

func parseSingle(v string) (*Constraint, error) {
	if strings.TrimSpace(v) == "*" {
		return &Constraint{
			f:            constraintWildcard,
			check:        nil,
			original:     v,
			origSegments: 1,
			operator:     "*",
		}, nil
	}
	if stability, ok := standaloneStabilityConstraint(v); ok {
		return &Constraint{
			f:            constraintWildcard,
			check:        nil,
			original:     v,
			stability:    stability,
			origSegments: 1,
			operator:     "*",
		}, nil
	}

	operator, version, stability, ok := splitConstraintParts(v)
	if !ok {
		return nil, fmt.Errorf("malformed constraint: %s", v)
	}

	if stability != "" {
		if _, ok := stabilityLevels[stability]; !ok {
			return nil, fmt.Errorf("unknown stability: %s", stability)
		}
	}

	// Composer strips the @stability flag from the constraint and discards it
	// when it is "stable" (see VersionParser::parseConstraint). The remaining
	// flags only influence the implicit prerelease suffix appended below, so a
	// trailing "@stable" must behave exactly like no flag at all.
	if stability == "stable" {
		stability = ""
	}

	version, err := stripDevReference(version, v)
	if err != nil {
		return nil, err
	}

	version = normalizeConstraintVersionTypos(version)
	version = applyStabilitySuffix(version, stability, operator)
	origSegments := countVersionSegments(version)
	stableBound := hasStableModifier(version)

	// Handle wildcards in version numbers
	if isWildcardConstraintVersion(version) && !isNumericDevBranch(version) {
		return parseWildcardConstraint(operator, version, v, "")
	}

	check, err := NewVersion(version)
	if err != nil {
		return nil, err
	}

	return &Constraint{
		f:            constraintOperators[operator],
		check:        check,
		original:     v,
		origSegments: origSegments,
		operator:     operator,
		stableBound:  stableBound,
	}, nil
}

func normalizeConstraintVersionTypos(version string) string {
	version = strings.TrimSpace(version)
	version = strings.ReplaceAll(version, "..dev", "-dev")
	version = strings.ReplaceAll(version, "..DEV", "-DEV")
	version = strings.ReplaceAll(version, "-.dev", "-dev")
	version = strings.ReplaceAll(version, "-.DEV", "-DEV")
	version = strings.ReplaceAll(version, "_-dev", "-dev")
	version = strings.ReplaceAll(version, "_-DEV", "-DEV")
	return strings.TrimRight(version, ".")
}

func standaloneStabilityConstraint(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if len(s) < 2 || s[0] != '@' || !allAlpha(s[1:]) {
		return "", false
	}
	stability := strings.ToLower(s[1:])
	if _, ok := stabilityLevels[stability]; !ok {
		return "", false
	}
	return stability, true
}

func applyStabilitySuffix(version, stability, operator string) string {
	if stability == "" || !operatorUsesStabilitySuffix(operator) || hasNormalizedPrerelease(version) {
		return version
	}
	return version + "-" + stability
}

func operatorUsesStabilitySuffix(operator string) bool {
	switch operator {
	case ">", ">=", "<", "<=", "^", "~":
		return true
	default:
		return false
	}
}

func hasNormalizedPrerelease(version string) bool {
	normalized, err := normalizeVersion(version)
	if err != nil {
		return false
	}
	return strings.Contains(normalized, "-")
}

func stripDevReference(version, original string) (string, error) {
	index := strings.Index(version, "#")
	if index < 0 {
		return version, nil
	}

	base := version[:index]
	if strings.HasPrefix(strings.ToLower(base), "dev-") || isNumericDevBranch(base) {
		return base, nil
	}

	return "", fmt.Errorf("malformed constraint: %s", original)
}

func splitConstraintParts(raw string) (operator, version, stability string, ok bool) {
	s := strings.TrimSpace(raw)
	for _, op := range []string{"==", "!=", "<>", ">=", "<=", "^", "~", ">", "<", "="} {
		if strings.HasPrefix(s, op) {
			operator = op
			s = strings.TrimSpace(s[len(op):])
			break
		}
	}
	if s == "" {
		return "", "", "", false
	}

	if at := strings.LastIndex(s, "@"); at > 0 && at < len(s)-1 && allAlpha(s[at+1:]) {
		stability = strings.ToLower(s[at+1:])
		s = strings.TrimSpace(s[:at])
	}
	if s == "" {
		return "", "", "", false
	}

	if _, ok := constraintOperators[operator]; !ok {
		return "", "", "", false
	}

	return operator, s, stability, true
}

func allAlpha(s string) bool {
	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}
	return true
}

func hasStableModifier(version string) bool {
	version = strings.TrimSpace(version)
	if i := strings.IndexAny(version, "+@"); i >= 0 {
		version = version[:i]
	}
	lower := strings.ToLower(version)
	return strings.HasSuffix(lower, "-stable") || strings.HasSuffix(lower, ".stable") || strings.HasSuffix(lower, "_stable")
}

func countVersionSegments(version string) int {
	version = strings.TrimSpace(version)
	if len(version) > 1 && (version[0] == 'v' || version[0] == 'V') {
		version = version[1:]
	}
	if i := strings.IndexAny(version, "-+@"); i >= 0 {
		version = version[:i]
	}
	if version == "" {
		return 1
	}
	return strings.Count(version, ".") + 1
}

func isNumericDevBranch(version string) bool {
	lower := strings.ToLower(strings.TrimSpace(version))
	if !strings.HasSuffix(lower, "-dev") {
		return false
	}

	base := strings.TrimSuffix(version, version[len(version)-4:])
	if len(base) > 1 && (base[0] == 'v' || base[0] == 'V') {
		base = base[1:]
	}
	parts := strings.Split(base, ".")
	if len(parts) == 0 || len(parts) > 4 {
		return false
	}
	seenWildcard := false
	for _, part := range parts {
		if part == "" {
			return false
		}
		if isWildcardSegment(part) {
			if seenWildcard {
				return false
			}
			seenWildcard = true
			continue
		}
		if seenWildcard {
			return false
		}
		if _, err := strconv.ParseInt(part, 10, 64); err != nil {
			return false
		}
	}
	return true
}

func isWildcardConstraintVersion(version string) bool {
	base := strings.TrimSpace(version)
	if len(base) > 1 && (base[0] == 'v' || base[0] == 'V') {
		base = base[1:]
	}
	if i := strings.IndexAny(base, "-+"); i >= 0 {
		base = base[:i]
	}
	for _, part := range strings.Split(base, ".") {
		if isWildcardSegment(part) {
			return true
		}
	}
	return false
}

func isWildcardSegment(part string) bool {
	return part == "*" || strings.EqualFold(part, "x")
}

// parseWildcardConstraint handles parsing of version constraints containing wildcards.
// It validates the wildcard pattern and creates appropriate constraint functions.
func parseWildcardConstraint(operator, version, original, stability string) (*Constraint, error) {
	version = strings.TrimSpace(version)
	if len(version) > 1 && (version[0] == 'v' || version[0] == 'V') {
		version = version[1:]
	}
	if operator == "^" || operator == "~" {
		return nil, fmt.Errorf("malformed constraint: %s", original)
	}
	if strings.ContainsAny(version, "-+") {
		return nil, fmt.Errorf("malformed constraint: %s", original)
	}

	parts := strings.Split(version, ".")
	wildcardIndex := -1
	fixed := []int64{}
	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("malformed constraint: %s", original)
		}
		if isWildcardSegment(part) {
			if wildcardIndex == -1 {
				wildcardIndex = i
			}
			continue
		}
		if wildcardIndex != -1 {
			return nil, fmt.Errorf("malformed constraint: %s", original)
		}
		segment, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("malformed constraint: %s", original)
		}
		fixed = append(fixed, segment)
	}

	if wildcardIndex == -1 {
		return nil, fmt.Errorf("malformed constraint: %s", original)
	}

	if len(fixed) == 0 {
		if operator != "" && operator != "=" && operator != "==" {
			return nil, fmt.Errorf("malformed constraint: %s", original)
		}
		if strings.TrimSpace(original) != "*" {
			check, err := NewVersion("0.0.0-dev")
			if err != nil {
				return nil, err
			}
			return &Constraint{
				f:            constraintGreaterThanEqual,
				check:        check,
				original:     original,
				stability:    stability,
				operator:     ">=",
				origSegments: 1,
			}, nil
		}
		return &Constraint{
			f:            constraintWildcard,
			check:        nil,
			original:     original,
			stability:    stability,
			operator:     operator,
			origSegments: 1,
		}, nil
	}

	checkParts := make([]string, len(fixed))
	for i, segment := range fixed {
		checkParts[i] = strconv.FormatInt(segment, 10)
	}
	for len(checkParts) < 3 {
		checkParts = append(checkParts, "0")
	}
	check, err := NewVersion(strings.Join(checkParts, "."))
	if err != nil {
		return nil, err
	}

	fixedLen := len(fixed)
	return &Constraint{
		f: func(v, c *Version, origSegments int) bool {
			cmp := comparePrefix(v.segments, c.segments[:fixedLen])
			switch operator {
			case ">=":
				return cmp >= 0
			case ">":
				return cmp > 0
			case "<=":
				return cmp <= 0
			case "<":
				return cmp < 0
			case "!=", "<>":
				return cmp != 0
			default:
				return cmp == 0
			}
		},
		check:        check,
		original:     original,
		stability:    stability,
		operator:     operator,
		origSegments: fixedLen + 1,
	}, nil
}

func comparePrefix(versionSegments, prefix []int64) int {
	for i, expected := range prefix {
		if i >= len(versionSegments) {
			if expected == 0 {
				continue
			}
			return -1
		}
		if versionSegments[i] < expected {
			return -1
		}
		if versionSegments[i] > expected {
			return 1
		}
	}
	return 0
}

func prereleaseCheck(v, c *Version) bool {
	switch vPre, cPre := v.Prerelease() != "", c.Prerelease() != ""; {
	case cPre && vPre:
		return true

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

func constraintEqual(v, c *Version, origSegments int) bool {
	return v.Equal(c)
}

func constraintNotEqual(v, c *Version, origSegments int) bool {
	return !v.Equal(c)
}

func constraintGreaterThan(v, c *Version, origSegments int) bool {
	return (bothNotPreRelease(v, c) || prereleaseCheck(v, c)) && v.GreaterThan(c)
}

func constraintLessThan(v, c *Version, origSegments int) bool {
	if v.IsPrerelease() && !c.IsPrerelease() && reflect.DeepEqual(v.Segments64(), c.Segments64()) {
		return false
	}
	return (bothNotPreRelease(v, c) || prereleaseCheck(v, c)) && v.LessThan(c)
}

func constraintGreaterThanEqual(v, c *Version, origSegments int) bool {
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

	return (bothNotPreRelease(v, c) || prereleaseCheck(v, c)) && v.GreaterThanOrEqual(c)
}

func constraintLessThanEqual(v, c *Version, origSegments int) bool {
	return (bothNotPreRelease(v, c) || prereleaseCheck(v, c)) && v.LessThanOrEqual(c)
}

func constraintCaret(v, c *Version, origSegments int) bool {
	compareVersion := v
	compareConstraint := c

	if v.IsPrerelease() && !c.IsPrerelease() {
		compareVersion = &Version{
			segments: v.segments,
			si:       v.si,
		}
		compareConstraint = &Version{
			segments: c.segments,
			si:       c.si,
		}
	}

	// If the version being checked is naturally less than the constraint, then there
	// is no way for the version to be valid against the constraint
	if compareVersion.LessThan(compareConstraint) {
		return false
	}

	if c.segments[0] != 0 {
		return v.segments[0] == c.segments[0]
	}

	if origSegments <= 1 {
		return v.segments[0] == 0
	}

	if c.segments[1] != 0 || origSegments <= 2 {
		return v.segments[0] == 0 && v.segments[1] == c.segments[1]
	}

	return v.segments[0] == 0 && v.segments[1] == 0 && v.segments[2] == c.segments[2]
}

func constraintTilde(v, c *Version, origSegments int) bool {
	// For tilde operator with prerelease versions, we need to compare without the prerelease tag first
	vNoPrerelease := &Version{
		segments: v.segments,
		si:       v.si,
	}
	cNoPrerelease := &Version{
		segments: c.segments,
		si:       c.si,
	}

	if c.IsPrerelease() {
		if v.LessThan(c) {
			return false
		}
	} else {
		// If the version without prerelease is less than the constraint without prerelease,
		// then there is no way for the version to be valid against the constraint
		if vNoPrerelease.LessThan(cNoPrerelease) {
			return false
		}
	}

	// Check the major version
	if v.segments[0] != c.segments[0] {
		return false
	}

	// Tilde constraint behavior in Composer:
	// ~X.Y.Z (3 segments) allows patch-level changes: >=X.Y.Z <X.(Y+1).0
	// ~X.Y (2 segments) allows minor-level changes: >=X.Y.0 <(X+1).0.0

	// For constraints with 4+ segments (~X.Y.Z.W), patch version must match.
	if origSegments >= 4 {
		if v.segments[1] != c.segments[1] || v.segments[2] != c.segments[2] {
			return false
		}
	} else if origSegments >= 3 {
		// For constraints with 3 segments (~X.Y.Z), minor version must match.
		if v.segments[1] != c.segments[1] {
			return false
		}
	}
	// For constraints with 2 segments (~X.Y), only major version must match
	// (already checked above)

	// Composer compiles ~X.Y.Z into >=X.Y.Z(-dev) <upper, so the implicit
	// lower bound carries dev stability and prereleases that share the base
	// segments (e.g. 1.2.3-beta for ~1.2.3) are in range. The only exclusion
	// is for explicitly stable constraints, which Check() applies centrally
	// via stableBound. The lower and upper bounds were already enforced above.
	return true
}

func constraintWildcard(v, c *Version, origSegments int) bool {
	return true
}

func bothNotPreRelease(v, c *Version) bool {
	return !v.IsPrerelease() || !c.IsPrerelease()
}

func getVersionStability(v *Version) string {
	if !v.IsPrerelease() {
		return "stable"
	}
	pre := strings.ToLower(v.Prerelease())
	switch {
	case strings.HasPrefix(pre, "dev"):
		return "dev"
	case strings.HasPrefix(pre, "alpha") || strings.HasPrefix(pre, "a"):
		return "alpha"
	case strings.HasPrefix(pre, "beta") || strings.HasPrefix(pre, "b"):
		return "beta"
	case strings.HasPrefix(pre, "rc"):
		return "rc"
	case strings.HasPrefix(pre, "patch") || strings.HasPrefix(pre, "p"):
		return "stable"
	default:
		return "dev"
	}
}
