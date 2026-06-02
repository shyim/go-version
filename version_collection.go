package version

// Collection is a type that implements the sort.Interface interface
// so that versions can be sorted.
type Collection []*Version

func (v Collection) Len() int {
	return len(v)
}

func (v Collection) Less(i, j int) bool {
	return compareForSort(v[i], v[j]) < 0
}

func (v Collection) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func compareForSort(left, right *Version) int {
	return sortVersion(left).Compare(sortVersion(right))
}

func sortVersion(v *Version) *Version {
	switch v.branch {
	case "dev-master", "dev-default", "dev-trunk":
		return &Version{
			pre:      "dev",
			segments: []int64{9999999, 0, 0, 0},
			si:       1,
			original: v.original,
		}
	default:
		return v
	}
}
