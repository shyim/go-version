package version

import (
	"strconv"
	"strings"
)

// This file holds the constraint-domain model: the union-of-intervals plus
// branch representation of a parsed constraint, and the operations that build,
// intersect, and test subset relationships between domains.

type constraintDomain struct {
	numeric       []versionInterval
	anyBranch     bool
	branches      map[string]struct{}
	branchExclude map[string]struct{}
}

func constraintsDomain(constraints []*Constraint) (constraintDomain, error) {
	domain := allConstraintDomain()
	for _, constraint := range constraints {
		next, err := singleConstraintDomain(constraint)
		if err != nil {
			return constraintDomain{}, err
		}
		domain = intersectDomains(domain, next)
		if domainEmpty(domain) {
			return domain, nil
		}
	}
	return domain, nil
}

func singleConstraintDomain(c *Constraint) (constraintDomain, error) {
	if c.check == nil {
		return allConstraintDomain(), nil
	}

	if c.check.branch != "" {
		switch c.operator {
		case "", "=", "==":
			return branchOnlyDomain(c.check.branch), nil
		case "!=", "<>":
			domain := allConstraintDomain()
			domain.branchExclude[c.check.branch] = struct{}{}
			return domain, nil
		default:
			return constraintDomain{}, nil
		}
	}

	if isConstraintWildcard(c) {
		return wildcardConstraintDomain(c), nil
	}

	switch c.operator {
	case "", "=", "==":
		return numericOnlyDomain(versionInterval{
			lower: inclusiveBound(c.check),
			upper: inclusiveBound(c.check),
		}), nil
	case "!=",
		"<>":
		domain := allConstraintDomain()
		domain.numeric[0].exclusions = append(domain.numeric[0].exclusions, c.check)
		return domain, nil
	case ">":
		return numericOnlyDomain(versionInterval{lower: exclusiveBound(c.check)}), nil
	case ">=":
		return numericOnlyDomain(versionInterval{lower: inclusiveBound(implicitDevLowerBound(c))}), nil
	case "<":
		return numericOnlyDomain(versionInterval{upper: exclusiveBound(implicitDevUpperBound(c))}), nil
	case "<=":
		return numericOnlyDomain(versionInterval{upper: inclusiveBound(c.check)}), nil
	case "^":
		return caretConstraintDomain(c), nil
	case "~":
		return tildeConstraintDomain(c), nil
	default:
		return constraintDomain{}, nil
	}
}

func allConstraintDomain() constraintDomain {
	return constraintDomain{
		numeric:       []versionInterval{{}},
		anyBranch:     true,
		branches:      map[string]struct{}{},
		branchExclude: map[string]struct{}{},
	}
}

func numericOnlyDomain(interval versionInterval) constraintDomain {
	return constraintDomain{
		numeric:       []versionInterval{interval},
		branches:      map[string]struct{}{},
		branchExclude: map[string]struct{}{},
	}
}

func branchOnlyDomain(branch string) constraintDomain {
	return constraintDomain{
		branches:      map[string]struct{}{branch: {}},
		branchExclude: map[string]struct{}{},
	}
}

func isConstraintWildcard(c *Constraint) bool {
	if c.check == nil {
		return true
	}
	_, version, _, ok := splitConstraintParts(c.original)
	return ok && c.origSegments > 1 && isWildcardConstraintVersion(version) && !isNumericDevBranch(version)
}

func wildcardConstraintDomain(c *Constraint) constraintDomain {
	fixedLen := c.origSegments - 1
	if fixedLen <= 0 || c.check == nil {
		return allConstraintDomain()
	}

	lower := implicitDevVersion(prefixStart(c.check, fixedLen))
	upper := implicitDevVersion(prefixEnd(c.check, fixedLen))

	switch c.operator {
	case ">=":
		return numericOnlyDomain(versionInterval{lower: inclusiveBound(lower)})
	case ">":
		return numericOnlyDomain(versionInterval{lower: inclusiveBound(upper)})
	case "<=":
		return numericOnlyDomain(versionInterval{upper: exclusiveBound(upper)})
	case "<":
		return numericOnlyDomain(versionInterval{upper: exclusiveBound(lower)})
	case "!=", "<>":
		return allConstraintDomain()
	default:
		return numericOnlyDomain(versionInterval{
			lower: inclusiveBound(lower),
			upper: exclusiveBound(upper),
		})
	}
}

func caretConstraintDomain(c *Constraint) constraintDomain {
	return numericOnlyDomain(versionInterval{
		lower: inclusiveBound(implicitDevLowerBound(c)),
		upper: exclusiveBound(implicitDevVersion(caretUpperBound(c))),
	})
}

func caretUpperBound(c *Constraint) *Version {
	segments := c.check.Segments64()
	if segments[0] == 0 && isMajorZeroNumericDevBranchWildcard(c) {
		return numericVersion(1, 0, 0)
	}
	switch {
	case segments[0] != 0:
		return numericVersion(segments[0]+1, 0, 0)
	case c.origSegments <= 1:
		return numericVersion(1, 0, 0)
	case segments[1] != 0 || c.origSegments <= 2:
		return numericVersion(0, segments[1]+1, 0)
	default:
		return numericVersion(0, 0, segments[2]+1)
	}
}

func isMajorZeroNumericDevBranchWildcard(c *Constraint) bool {
	_, version, _, ok := splitConstraintParts(c.original)
	if !ok || !isNumericDevBranch(version) {
		return false
	}
	version = strings.TrimSpace(version)
	if len(version) > 1 && (version[0] == 'v' || version[0] == 'V') {
		version = version[1:]
	}
	version = strings.TrimSuffix(version, version[len(version)-4:])
	parts := strings.Split(version, ".")
	return len(parts) == 2 && parts[0] == "0" && isWildcardSegment(parts[1])
}

func tildeConstraintDomain(c *Constraint) constraintDomain {
	segments := c.check.Segments64()
	upper := numericVersion(segments[0]+1, 0, 0)
	if c.origSegments >= 4 {
		upper = numericVersion(segments[0], segments[1], segments[2]+1)
	} else if c.origSegments >= 3 {
		upper = numericVersion(segments[0], segments[1]+1, 0)
	}

	return numericOnlyDomain(versionInterval{
		lower: inclusiveBound(implicitDevLowerBound(c)),
		upper: exclusiveBound(implicitDevVersion(upper)),
	})
}

func implicitDevLowerBound(c *Constraint) *Version {
	if c.stableBound || c.check.IsPrerelease() {
		return c.check
	}
	return implicitDevVersion(c.check)
}

func implicitDevUpperBound(c *Constraint) *Version {
	if c.stableBound || c.check.IsPrerelease() {
		return c.check
	}
	return implicitDevVersion(c.check)
}

func implicitDevVersion(v *Version) *Version {
	if v.IsPrerelease() || v.branch != "" {
		return v
	}
	return &Version{
		segments: v.Segments64(),
		si:       v.si,
		original: v.original,
		pre:      "dev",
	}
}

func prefixStart(v *Version, fixedLen int) *Version {
	segments := v.Segments64()
	result := append([]int64{}, segments[:fixedLen]...)
	for len(result) < 3 {
		result = append(result, 0)
	}
	return numericVersion(result...)
}

func prefixEnd(v *Version, fixedLen int) *Version {
	segments := append([]int64{}, v.Segments64()...)
	for len(segments) < fixedLen {
		segments = append(segments, 0)
	}
	result := append([]int64{}, segments[:fixedLen]...)
	result[len(result)-1]++
	for len(result) < 3 {
		result = append(result, 0)
	}
	return numericVersion(result...)
}

func numericVersion(segments ...int64) *Version {
	copied := append([]int64{}, segments...)
	for len(copied) < 4 {
		copied = append(copied, 0)
	}
	parts := make([]string, len(copied))
	for i, segment := range copied {
		parts[i] = strconv.FormatInt(segment, 10)
	}
	return &Version{segments: copied, si: len(copied), original: joinVersionParts(parts)}
}

func joinVersionParts(parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "."
		}
		result += part
	}
	return result
}

func intersectDomains(left, right constraintDomain) constraintDomain {
	result := constraintDomain{
		branches:      map[string]struct{}{},
		branchExclude: mergeBranchExclusions(left.branchExclude, right.branchExclude),
	}

	for _, leftInterval := range left.numeric {
		for _, rightInterval := range right.numeric {
			if interval, ok := intersectIntervals(leftInterval, rightInterval); ok {
				result.numeric = append(result.numeric, interval)
			}
		}
	}

	result.anyBranch = left.anyBranch && right.anyBranch
	for branch := range left.branches {
		if right.anyBranch || hasBranch(right.branches, branch) {
			addBranchIfAllowed(result.branches, branch, result.branchExclude)
		}
	}
	for branch := range right.branches {
		if left.anyBranch || hasBranch(left.branches, branch) {
			addBranchIfAllowed(result.branches, branch, result.branchExclude)
		}
	}

	return result
}

func mergeBranchExclusions(left, right map[string]struct{}) map[string]struct{} {
	result := map[string]struct{}{}
	for branch := range left {
		result[branch] = struct{}{}
	}
	for branch := range right {
		result[branch] = struct{}{}
	}
	return result
}

func hasBranch(branches map[string]struct{}, branch string) bool {
	_, ok := branches[branch]
	return ok
}

func addBranchIfAllowed(branches map[string]struct{}, branch string, excluded map[string]struct{}) {
	if _, ok := excluded[branch]; ok {
		return
	}
	branches[branch] = struct{}{}
}

func domainsIntersect(left, right constraintDomain) bool {
	intersection := intersectDomains(left, right)
	if len(intersection.numeric) > 0 {
		return true
	}
	if len(intersection.branches) > 0 {
		return true
	}
	return intersection.anyBranch
}

func domainEmpty(domain constraintDomain) bool {
	return len(domain.numeric) == 0 && !domain.anyBranch && len(domain.branches) == 0
}

func domainSubsetOfUnion(left constraintDomain, rights []constraintDomain) bool {
	rightIntervals := unionNumericIntervals(rights)
	for _, leftInterval := range left.numeric {
		if !intervalSubsetOfUnion(leftInterval, rightIntervals) {
			return false
		}
	}

	for branch := range left.branches {
		if !unionAllowsBranch(rights, branch) {
			return false
		}
	}

	if left.anyBranch && !anyBranchSubsetOfUnion(left, rights) {
		return false
	}

	return true
}

func domainUnionSubsetOfUnion(left, right []constraintDomain) bool {
	for _, leftDomain := range left {
		if domainEmpty(leftDomain) {
			continue
		}
		if !domainSubsetOfUnion(leftDomain, right) {
			return false
		}
	}
	return true
}

func domainUnionsEquivalent(left, right []constraintDomain) bool {
	return domainUnionSubsetOfUnion(left, right) && domainUnionSubsetOfUnion(right, left)
}

func domainUnionEmpty(domains []constraintDomain) bool {
	for _, domain := range domains {
		if !domainEmpty(domain) {
			return false
		}
	}
	return true
}

func anyBranchSubsetOfUnion(left constraintDomain, rights []constraintDomain) bool {
	if !unionHasAnyBranch(rights) {
		return false
	}

	for branch := range branchesExcludedByEveryAnyBranch(rights) {
		if _, leftExcludes := left.branchExclude[branch]; !leftExcludes {
			return false
		}
	}

	return true
}

func unionHasAnyBranch(domains []constraintDomain) bool {
	for _, domain := range domains {
		if domain.anyBranch {
			return true
		}
	}
	return false
}

func branchesExcludedByEveryAnyBranch(domains []constraintDomain) map[string]struct{} {
	var excluded map[string]struct{}
	for _, domain := range domains {
		if !domain.anyBranch {
			continue
		}
		if excluded == nil {
			excluded = copyStringSet(domain.branchExclude)
			continue
		}
		for branch := range excluded {
			if _, ok := domain.branchExclude[branch]; !ok {
				delete(excluded, branch)
			}
		}
	}
	if excluded == nil {
		return map[string]struct{}{}
	}
	for _, domain := range domains {
		for branch := range domain.branches {
			delete(excluded, branch)
		}
	}
	return excluded
}

func copyStringSet(input map[string]struct{}) map[string]struct{} {
	result := map[string]struct{}{}
	for value := range input {
		result[value] = struct{}{}
	}
	return result
}

func unionAllowsBranch(domains []constraintDomain, branch string) bool {
	for _, domain := range domains {
		if domain.anyBranch {
			if _, excluded := domain.branchExclude[branch]; !excluded {
				return true
			}
		}
		if _, ok := domain.branches[branch]; ok {
			return true
		}
	}
	return false
}
