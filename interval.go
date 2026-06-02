package version

import "sort"

// This file holds the numeric version-interval primitives: half-open ranges
// with inclusive/exclusive bounds, and the lower/upper bound arithmetic used to
// intersect and cover intervals.

type versionInterval struct {
	lower      *versionBound
	upper      *versionBound
	exclusions []*Version
}

type versionBound struct {
	version   *Version
	inclusive bool
}

func minLowerBound(left, right *versionBound) *versionBound {
	if left == nil || right == nil {
		return nil
	}
	cmp := left.version.Compare(right.version)
	if cmp < 0 {
		return left
	}
	if cmp > 0 {
		return right
	}
	return &versionBound{version: left.version, inclusive: left.inclusive || right.inclusive}
}

func inclusiveBound(v *Version) *versionBound {
	return &versionBound{version: v, inclusive: true}
}

func exclusiveBound(v *Version) *versionBound {
	return &versionBound{version: v}
}

func intersectIntervals(left, right versionInterval) (versionInterval, bool) {
	interval := versionInterval{
		lower:      maxLower(left.lower, right.lower),
		upper:      minUpper(left.upper, right.upper),
		exclusions: append(append([]*Version{}, left.exclusions...), right.exclusions...),
	}
	if !intervalHasAnyVersion(interval) {
		return versionInterval{}, false
	}
	return interval, true
}

func maxLower(left, right *versionBound) *versionBound {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	cmp := left.version.Compare(right.version)
	if cmp > 0 {
		return left
	}
	if cmp < 0 {
		return right
	}
	return &versionBound{version: left.version, inclusive: left.inclusive && right.inclusive}
}

func minUpper(left, right *versionBound) *versionBound {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	cmp := left.version.Compare(right.version)
	if cmp < 0 {
		return left
	}
	if cmp > 0 {
		return right
	}
	return &versionBound{version: left.version, inclusive: left.inclusive && right.inclusive}
}

func maxUpper(left, right *versionBound) *versionBound {
	if left == nil || right == nil {
		return nil
	}
	cmp := left.version.Compare(right.version)
	if cmp > 0 {
		return left
	}
	if cmp < 0 {
		return right
	}
	return &versionBound{version: left.version, inclusive: left.inclusive || right.inclusive}
}

func intervalHasAnyVersion(interval versionInterval) bool {
	if interval.lower != nil && interval.upper != nil {
		cmp := interval.lower.version.Compare(interval.upper.version)
		if cmp > 0 {
			return false
		}
		if cmp == 0 {
			if !interval.lower.inclusive || !interval.upper.inclusive {
				return false
			}
			return !versionExcluded(interval.lower.version, interval.exclusions)
		}
	}
	return true
}

func versionExcluded(v *Version, exclusions []*Version) bool {
	for _, excluded := range exclusions {
		if v.Equal(excluded) {
			return true
		}
	}
	return false
}

func intervalsCanMerge(left, right versionInterval) bool {
	if left.upper == nil || right.lower == nil {
		return true
	}
	cmp := left.upper.version.Compare(right.lower.version)
	if cmp > 0 {
		return true
	}
	if cmp < 0 {
		return false
	}
	return left.upper.inclusive || right.lower.inclusive
}

func unionNumericIntervals(domains []constraintDomain) []versionInterval {
	var intervals []versionInterval
	for _, domain := range domains {
		intervals = append(intervals, domain.numeric...)
	}
	return intervals
}

func intervalSubsetOfUnion(left versionInterval, rights []versionInterval) bool {
	if !intervalCoveredByUnion(left, rights) {
		return false
	}

	for _, excluded := range unionNumericExclusions(rights) {
		if !versionInInterval(excluded, left, false) {
			continue
		}
		if versionExcluded(excluded, left.exclusions) {
			continue
		}
		if !unionAllowsNumericVersion(rights, excluded) {
			return false
		}
	}

	return true
}

func intervalCoveredByUnion(left versionInterval, rights []versionInterval) bool {
	if len(rights) == 0 {
		return false
	}

	sorted := append([]versionInterval{}, rights...)
	sort.Slice(sorted, func(i, j int) bool {
		return lowerBefore(sorted[i].lower, sorted[j].lower)
	})

	cursor := left.lower
	for _, right := range sorted {
		if upperBeforeCursor(right.upper, cursor) {
			continue
		}
		if !lowerCoversCursor(right.lower, cursor) {
			return false
		}
		if upperReaches(right.upper, left.upper) {
			return true
		}
		if right.upper == nil {
			return true
		}
		cursor = &versionBound{version: right.upper.version, inclusive: !right.upper.inclusive}
	}

	return false
}

func lowerBefore(left, right *versionBound) bool {
	if left == nil {
		return right != nil
	}
	if right == nil {
		return false
	}
	cmp := left.version.Compare(right.version)
	if cmp != 0 {
		return cmp < 0
	}
	if left.inclusive == right.inclusive {
		return false
	}
	return left.inclusive
}

func upperBeforeCursor(upper, cursor *versionBound) bool {
	if upper == nil || cursor == nil {
		return false
	}
	cmp := upper.version.Compare(cursor.version)
	if cmp != 0 {
		return cmp < 0
	}
	return !upper.inclusive || !cursor.inclusive
}

func lowerCoversCursor(lower, cursor *versionBound) bool {
	if lower == nil {
		return true
	}
	if cursor == nil {
		return false
	}
	cmp := lower.version.Compare(cursor.version)
	if cmp != 0 {
		return cmp < 0
	}
	if cursor.inclusive && !lower.inclusive {
		return false
	}
	return true
}

func upperReaches(upper, target *versionBound) bool {
	if target == nil {
		return upper == nil
	}
	if upper == nil {
		return true
	}
	cmp := upper.version.Compare(target.version)
	if cmp != 0 {
		return cmp > 0
	}
	if target.inclusive && !upper.inclusive {
		return false
	}
	return true
}

func unionNumericExclusions(intervals []versionInterval) []*Version {
	var exclusions []*Version
	for _, interval := range intervals {
		exclusions = append(exclusions, interval.exclusions...)
	}
	return exclusions
}

func unionAllowsNumericVersion(intervals []versionInterval, version *Version) bool {
	for _, interval := range intervals {
		if versionInInterval(version, interval, true) {
			return true
		}
	}
	return false
}

func versionInInterval(version *Version, interval versionInterval, respectExclusions bool) bool {
	if interval.lower != nil {
		cmp := version.Compare(interval.lower.version)
		if cmp < 0 || (cmp == 0 && !interval.lower.inclusive) {
			return false
		}
	}
	if interval.upper != nil {
		cmp := version.Compare(interval.upper.version)
		if cmp > 0 || (cmp == 0 && !interval.upper.inclusive) {
			return false
		}
	}
	return !respectExclusions || !versionExcluded(version, interval.exclusions)
}
