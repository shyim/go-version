package version

import "testing"

// parityCorpus is a broad set of version strings spanning every parser path,
// used to assert that the optimized comparison fast-path stays behaviorally
// identical to the canonical NormalizedString-based comparison.
var parityCorpus = []string{
	"0", "1", "1.0", "1.2", "1.2.3", "1.2.3.4", "1.2.3.0",
	"v1.2.3", "V2.0.0", "01.02.03", "000000", "00.00",
	"1.0.0-dev", "1.0.0-alpha", "1.0.0-alpha1", "1.0.0-beta", "1.0.0-beta.5",
	"1.0.0-rc1", "1.0.0-RC1", "1.0.0-patch1", "1.0.0-p1", "1.0.0-stable",
	"1.0.0", "1.0.0.0", "1.2", "1.2.0", "1.2.0.0",
	"20100102", "201903.0", "1.2.3+build.123", "1.0.0-beta+exp.sha",
	"dev-main", "dev-master", "dev-feature/x", "2.1.x-dev",
	"100000.00.0.00", "100000.0.0",
}

// referenceCompare replicates the pre-optimization Compare exactly (the
// NormalizedString-string fast-path followed by the jagged, allZero-aware
// segment loop). It is the behavioral oracle: the optimized Compare must return
// the identical result for every pair, so the fast-path rewrite is provably
// behavior-preserving rather than merely "matches NormalizedString".
func referenceCompare(v, other *Version) int {
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
	ss := v.Segments64()
	so := other.Segments64()
	eq := len(ss) == len(so)
	if eq {
		for i := range ss {
			if ss[i] != so[i] {
				eq = false
				break
			}
		}
	}
	if eq {
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
	lenSelf := len(ss)
	lenOther := len(so)
	hS := lenSelf
	if lenSelf < lenOther {
		hS = lenOther
	}
	for i := 0; i < hS; i++ {
		if i > lenSelf-1 {
			if !allZero(so[i:]) {
				return -1
			}
			break
		} else if i > lenOther-1 {
			if !allZero(ss[i:]) {
				return 1
			}
			break
		}
		if ss[i] == so[i] {
			continue
		} else if ss[i] < so[i] {
			return -1
		}
		return 1
	}
	return 0
}

// TestCompareParity is the core safety net for the optimized Compare: it must
// return the identical result to the reference (pre-optimization) implementation
// for every pair in the corpus, including the 6-digit "date-like" segment edge
// cases where NormalizedString and the jagged loop disagree.
func TestCompareParity(t *testing.T) {
	versions := make([]*Version, 0, len(parityCorpus))
	for _, s := range parityCorpus {
		v, err := NewVersion(s)
		if err != nil {
			t.Fatalf("seed %q failed to parse: %v", s, err)
		}
		versions = append(versions, v)
	}

	for i, a := range versions {
		for j, b := range versions {
			if got, want := a.Compare(b), referenceCompare(a, b); got != want {
				t.Errorf("Compare(%q, %q) = %d, reference = %d",
					parityCorpus[i], parityCorpus[j], got, want)
			}
		}
	}
}

// TestCompareDevBranchReflexivity guards the trap that sank several rejected
// optimization variants: dev branches all share pre="dev" and segments=[0,0,0],
// so a naive segments+pre fast-path would treat distinct branches as equal.
func TestCompareDevBranchReflexivity(t *testing.T) {
	same := Must(NewVersion("dev-master"))
	sameAgain := Must(NewVersion("dev-master"))
	if same.Compare(sameAgain) != 0 {
		t.Errorf("Compare(dev-master, dev-master) = %d, want 0", same.Compare(sameAgain))
	}

	foo := Must(NewVersion("dev-foo"))
	bar := Must(NewVersion("dev-bar"))
	if foo.Compare(bar) == 0 {
		t.Error("Compare(dev-foo, dev-bar) == 0, want non-zero (distinct branches)")
	}

	// A branch and a numeric version must never be equal.
	numeric := Must(NewVersion("0.0.0"))
	if foo.Compare(numeric) == 0 {
		t.Error("Compare(dev-foo, 0.0.0) == 0, want non-zero")
	}
}

// TestNormalizedStringIdempotency pins the previously-uncovered idempotency
// behavior so any future caching of NormalizedString cannot silently regress.
func TestNormalizedStringIdempotency(t *testing.T) {
	cases := map[string]string{
		"000000":         "0.0.0.0",
		"0":              "0.0.0.0",
		"1.2":            "1.2.0.0",
		"1.2.3":          "1.2.3.0",
		"20100102":       "20100102", // date version: must NOT be padded
		"v1.2.3":         "1.2.3.0",
		"1.0.0-beta.5":   "1.0.0.0-beta5",
	}
	for input, want := range cases {
		v, err := NewVersion(input)
		if err != nil {
			t.Errorf("NewVersion(%q) failed: %v", input, err)
			continue
		}
		got := v.NormalizedString()
		if got != want {
			t.Errorf("NormalizedString(%q) = %q, want %q", input, got, want)
		}
		// Re-parsing the output must yield the same canonical form.
		v2 := Must(NewVersion(got))
		if v2.NormalizedString() != got {
			t.Errorf("NormalizedString not idempotent for %q: %q -> %q", input, got, v2.NormalizedString())
		}
	}
}
