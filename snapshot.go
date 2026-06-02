package version

import (
	"fmt"
	"sort"
)

// This file holds the snapshot layer: serializable representations of domains,
// intervals, and bounds plus the helpers that render them. These are used to
// summarize and format constraint domains.

type boundSnapshot struct {
	Version   string
	Inclusive bool
}

type intervalSnapshot struct {
	Start string
	End   string
}

type branchSnapshot struct {
	Names   []string
	Exclude bool
}

type domainSnapshot struct {
	Numeric  []intervalSnapshot
	Branches branchSnapshot
}

func constraintBoundsSnapshot(operator, version string) (boundSnapshot, boundSnapshot, error) {
	v, err := NewVersion(version)
	if err != nil {
		return boundSnapshot{}, boundSnapshot{}, err
	}

	if v.branch != "" && operator != "" && operator != "=" && operator != "==" {
		return zeroBoundSnapshot(), positiveInfinityBoundSnapshot(), nil
	}

	bound := versionBoundSnapshot(v, true)
	switch operator {
	case "", "=", "==":
		return bound, bound, nil
	case "<":
		return zeroBoundSnapshot(), versionBoundSnapshot(v, false), nil
	case "<=":
		return zeroBoundSnapshot(), bound, nil
	case ">":
		return versionBoundSnapshot(v, false), positiveInfinityBoundSnapshot(), nil
	case ">=":
		return bound, positiveInfinityBoundSnapshot(), nil
	case "!=", "<>":
		return zeroBoundSnapshot(), positiveInfinityBoundSnapshot(), nil
	default:
		return boundSnapshot{}, boundSnapshot{}, nil
	}
}

func domainUnionBoundsSnapshot(domains []constraintDomain) (boundSnapshot, boundSnapshot) {
	intervals := unionNumericIntervals(domains)
	if len(intervals) == 0 {
		return zeroBoundSnapshot(), positiveInfinityBoundSnapshot()
	}

	var lower *versionBound
	var upper *versionBound
	initialized := false
	for _, interval := range intervals {
		if !initialized {
			lower = interval.lower
			upper = interval.upper
			initialized = true
			continue
		}
		lower = minLowerBound(lower, interval.lower)
		upper = maxUpper(upper, interval.upper)
	}

	return snapshotLowerBound(lower), snapshotUpperBound(upper)
}

func combineRawConstraintBoundsSnapshot(parts []string, conjunctive bool) (boundSnapshot, boundSnapshot, error) {
	if len(parts) == 0 {
		return zeroBoundSnapshot(), positiveInfinityBoundSnapshot(), nil
	}

	var lower boundSnapshot
	var upper boundSnapshot
	for i, part := range parts {
		operator, version, _, ok := splitConstraintParts(part)
		if !ok {
			return boundSnapshot{}, boundSnapshot{}, fmt.Errorf("malformed constraint: %s", part)
		}
		partLower, partUpper, err := constraintBoundsSnapshot(operator, version)
		if err != nil {
			return boundSnapshot{}, boundSnapshot{}, err
		}

		if i == 0 {
			lower = partLower
			upper = partUpper
			continue
		}

		if conjunctive {
			lower = maxLowerSnapshot(lower, partLower)
			upper = minUpperSnapshot(upper, partUpper)
		} else {
			lower = minLowerSnapshot(lower, partLower)
			upper = maxUpperSnapshot(upper, partUpper)
		}
	}

	return lower, upper, nil
}

func maxLowerSnapshot(left, right boundSnapshot) boundSnapshot {
	cmp := compareBoundSnapshotVersions(left.Version, right.Version)
	if cmp > 0 {
		return left
	}
	if cmp < 0 {
		return right
	}
	return boundSnapshot{Version: left.Version, Inclusive: left.Inclusive && right.Inclusive}
}

func minLowerSnapshot(left, right boundSnapshot) boundSnapshot {
	cmp := compareBoundSnapshotVersions(left.Version, right.Version)
	if cmp < 0 {
		return left
	}
	if cmp > 0 {
		return right
	}
	return boundSnapshot{Version: left.Version, Inclusive: left.Inclusive || right.Inclusive}
}

func minUpperSnapshot(left, right boundSnapshot) boundSnapshot {
	cmp := compareBoundSnapshotVersions(left.Version, right.Version)
	if cmp < 0 {
		return left
	}
	if cmp > 0 {
		return right
	}
	return boundSnapshot{Version: left.Version, Inclusive: left.Inclusive && right.Inclusive}
}

func maxUpperSnapshot(left, right boundSnapshot) boundSnapshot {
	cmp := compareBoundSnapshotVersions(left.Version, right.Version)
	if cmp > 0 {
		return left
	}
	if cmp < 0 {
		return right
	}
	return boundSnapshot{Version: left.Version, Inclusive: left.Inclusive || right.Inclusive}
}

func compareBoundSnapshotVersions(left, right string) int {
	if left == right {
		return 0
	}
	if left == "+Inf" {
		return 1
	}
	if right == "+Inf" {
		return -1
	}
	leftVersion := Must(NewVersion(left))
	rightVersion := Must(NewVersion(right))
	return leftVersion.Compare(rightVersion)
}

func snapshotLowerBound(bound *versionBound) boundSnapshot {
	if bound == nil {
		return zeroBoundSnapshot()
	}
	return versionBoundSnapshot(bound.version, bound.inclusive)
}

func snapshotUpperBound(bound *versionBound) boundSnapshot {
	if bound == nil {
		return positiveInfinityBoundSnapshot()
	}
	return versionBoundSnapshot(bound.version, bound.inclusive)
}

func versionBoundSnapshot(v *Version, inclusive bool) boundSnapshot {
	return boundSnapshot{Version: v.NormalizedString(), Inclusive: inclusive}
}

func zeroBoundSnapshot() boundSnapshot {
	return boundSnapshot{Version: "0.0.0.0-dev", Inclusive: true}
}

func positiveInfinityBoundSnapshot() boundSnapshot {
	return boundSnapshot{Version: "+Inf", Inclusive: false}
}

func domainUnionSnapshot(domains []constraintDomain) domainSnapshot {
	return domainSnapshot{
		Numeric:  snapshotNumericIntervals(unionNumericIntervals(domains)),
		Branches: snapshotBranches(domains),
	}
}

func snapshotNumericIntervals(intervals []versionInterval) []intervalSnapshot {
	segments := splitIntervalExclusions(intervals)
	if len(segments) == 0 {
		return nil
	}

	sort.Slice(segments, func(i, j int) bool {
		return lowerBefore(segments[i].lower, segments[j].lower)
	})

	merged := make([]versionInterval, 0, len(segments))
	for _, interval := range segments {
		if !intervalHasAnyVersion(interval) {
			continue
		}
		if len(merged) == 0 {
			merged = append(merged, interval)
			continue
		}
		last := &merged[len(merged)-1]
		if intervalsCanMerge(*last, interval) {
			last.upper = maxUpper(last.upper, interval.upper)
			continue
		}
		merged = append(merged, interval)
	}

	snapshots := make([]intervalSnapshot, 0, len(merged))
	for _, interval := range merged {
		snapshots = append(snapshots, intervalSnapshot{
			Start: formatLowerBound(interval.lower),
			End:   formatUpperBound(interval.upper),
		})
	}
	return snapshots
}

func splitIntervalExclusions(intervals []versionInterval) []versionInterval {
	var result []versionInterval
	for _, interval := range intervals {
		if !intervalHasAnyVersion(interval) {
			continue
		}

		exclusions := intervalExclusionsInside(interval)
		if len(exclusions) == 0 {
			result = append(result, versionInterval{lower: interval.lower, upper: interval.upper})
			continue
		}

		lower := interval.lower
		for _, exclusion := range exclusions {
			before := versionInterval{
				lower: lower,
				upper: exclusiveBound(exclusion),
			}
			if intervalHasAnyVersion(before) {
				result = append(result, before)
			}
			lower = exclusiveBound(exclusion)
		}

		after := versionInterval{
			lower: lower,
			upper: interval.upper,
		}
		if intervalHasAnyVersion(after) {
			result = append(result, after)
		}
	}
	return result
}

func intervalExclusionsInside(interval versionInterval) []*Version {
	seen := map[string]*Version{}
	for _, exclusion := range interval.exclusions {
		if versionInInterval(exclusion, interval, false) {
			seen[exclusion.NormalizedString()] = exclusion
		}
	}

	exclusions := make([]*Version, 0, len(seen))
	for _, exclusion := range seen {
		exclusions = append(exclusions, exclusion)
	}
	sort.Slice(exclusions, func(i, j int) bool {
		return exclusions[i].LessThan(exclusions[j])
	})
	return exclusions
}

func formatLowerBound(bound *versionBound) string {
	if bound == nil {
		return ">= 0.0.0.0-dev"
	}
	operator := ">="
	if !bound.inclusive {
		operator = ">"
	}
	return operator + " " + bound.version.NormalizedString()
}

func formatUpperBound(bound *versionBound) string {
	if bound == nil {
		return "< +Inf"
	}
	operator := "<="
	if !bound.inclusive {
		operator = "<"
	}
	return operator + " " + bound.version.NormalizedString()
}

func snapshotBranches(domains []constraintDomain) branchSnapshot {
	if unionHasAnyBranch(domains) {
		names := stringSetKeys(branchesExcludedByEveryAnyBranch(domains))
		return branchSnapshot{Names: names, Exclude: true}
	}

	names := map[string]struct{}{}
	for _, domain := range domains {
		for branch := range domain.branches {
			names[branch] = struct{}{}
		}
	}
	return branchSnapshot{Names: stringSetKeys(names)}
}

func stringSetKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
