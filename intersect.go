package version

import (
	"strings"
)

// ConstraintIntersects reports whether two Composer constraints can match at
// least one common version.
func ConstraintIntersects(left, right string) (bool, error) {
	left = normalizeConstraintInput(left)
	right = normalizeConstraintInput(right)

	leftDomains, err := constraintUnionDomains(left)
	if err != nil {
		return false, err
	}
	rightDomains, err := constraintUnionDomains(right)
	if err != nil {
		return false, err
	}

	for _, leftDomain := range leftDomains {
		for _, rightDomain := range rightDomains {
			if domainsIntersect(leftDomain, rightDomain) {
				return true, nil
			}
		}
	}

	return false, nil
}

// ConstraintSubsetOf reports whether every version matched by left is also
// matched by right.

// ConstraintSubsetOf reports whether every version matched by left is also
// matched by right.
func ConstraintSubsetOf(left, right string) (bool, error) {
	leftDomains, err := constraintUnionDomains(normalizeConstraintInput(left))
	if err != nil {
		return false, err
	}
	rightDomains, err := constraintUnionDomains(normalizeConstraintInput(right))
	if err != nil {
		return false, err
	}

	for _, leftDomain := range leftDomains {
		if domainEmpty(leftDomain) {
			continue
		}
		if !domainSubsetOfUnion(leftDomain, rightDomains) {
			return false, nil
		}
	}

	return true, nil
}

func normalizeConstraintInput(constraint string) string {
	constraint = strings.TrimSpace(constraint)
	if constraint == "" {
		return "*"
	}
	return constraint
}

func constraintUnionDomains(constraint string) ([]constraintDomain, error) {
	parsed, err := NewConstraint(constraint)
	if err != nil {
		return nil, err
	}

	domains := make([]constraintDomain, 0, len(parsed))
	for _, andConstraints := range parsed {
		domain, err := constraintsDomain(andConstraints)
		if err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}
	return domains, nil
}

func composeConstraintDomains(parts []string, conjunctive bool) ([]constraintDomain, error) {
	if len(parts) == 0 {
		return []constraintDomain{allConstraintDomain()}, nil
	}

	if !conjunctive {
		var domains []constraintDomain
		for _, part := range parts {
			partDomains, err := constraintUnionDomains(normalizeConstraintInput(part))
			if err != nil {
				return nil, err
			}
			domains = append(domains, partDomains...)
		}
		return domains, nil
	}

	domains := []constraintDomain{allConstraintDomain()}
	for _, part := range parts {
		partDomains, err := constraintUnionDomains(normalizeConstraintInput(part))
		if err != nil {
			return nil, err
		}

		var next []constraintDomain
		for _, left := range domains {
			for _, right := range partDomains {
				intersection := intersectDomains(left, right)
				if !domainEmpty(intersection) {
					next = append(next, intersection)
				}
			}
		}
		domains = next
		if len(domains) == 0 {
			break
		}
	}

	return domains, nil
}
