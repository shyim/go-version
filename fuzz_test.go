package version

import (
	"strings"
	"testing"
)

// These fuzz targets exercise the untrusted-input entry points of the library
// (version strings and constraint strings). The properties they assert are
// deliberately simple and robust:
//
//   - parsing arbitrary input must never panic, only return an error;
//   - normalization is idempotent and its output re-parses;
//   - comparison is reflexive and antisymmetric;
//   - Satisfies agrees with NewConstraint(...).Check(...).
//
// Run a single target with, e.g.:
//
//	go test -run=Fuzz -fuzz=FuzzNewVersion -fuzztime=30s
//
// The static seed corpus below keeps these useful as fast regression tests
// under plain `go test` even when fuzzing is not engaged.

// versionSeeds covers well-formed, degenerate, and adversarial version inputs.
var versionSeeds = []string{
	"",
	" ",
	"0",
	"1",
	"1.2",
	"1.2.3",
	"1.2.3.4",
	"v1.2.3",
	"V1.2.3",
	"1.0.0-alpha",
	"1.0.0-alpha.1",
	"1.0.0-beta2",
	"1.0.0-rc.1",
	"1.0.0-RC1",
	"1.0.0-patch1",
	"1.0.0-p1",
	"1.0.0-pl3-dev",
	"1.0.0+build.123",
	"1.0.0-beta.5+foo",
	"dev-main",
	"dev-feature/x",
	"2.1.x-dev",
	"20100102-203040",
	"201903.0-p2",
	"1.0.0.0.0",
	"9999999999999999999999999.0.0",
	"99999.99999.99999",
	"1.",
	".1",
	"1..2",
	"-",
	"+",
	"@",
	"~",
	"^",
	"1.0.0-",
	"1.0.0+",
	"1.0.0@",
	"\x00",
	"1.0.0\n",
	"  1.2.3  ",
}

// constraintSeeds covers operators, ranges, wildcards, unions, and junk.
var constraintSeeds = []string{
	"",
	"*",
	"1.2.3",
	"=1.2.3",
	"==1.2.3",
	"!=1.2.3",
	">1.0",
	">=1.0",
	"<2.0",
	"<=2.0",
	"^1.0",
	"^0.2.0",
	"~1.2",
	"~1.2.3",
	"1.2.*",
	"2.*",
	">=1.0,<2.0",
	">=1.0 <2.0",
	"1.0 - 2.0",
	"^1.0 || ^2.0",
	">=1.0@stable",
	">=1.0@beta",
	"dev-main",
	"!= dev-foo",
	">>1.0",
	"^.*",
	"~x.y",
	"||",
	",",
	" - ",
	"1 - ",
	" - 2",
	"@stable",
	">=1.0.0@foo",
}

func FuzzNewVersion(f *testing.F) {
	for _, s := range versionSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		v, err := NewVersion(s)
		if err != nil {
			// A parse error is a valid, expected outcome for junk input.
			return
		}
		if v == nil {
			t.Fatalf("NewVersion(%q) returned nil version and nil error", s)
		}

		// A parsed version must compare equal to itself.
		if cmp := v.Compare(v); cmp != 0 {
			t.Fatalf("NewVersion(%q).Compare(self) = %d, want 0", s, cmp)
		}

		// The normalized form must re-parse, and re-parsing it must be stable.
		norm := v.NormalizedString()
		v2, err := NewVersion(norm)
		if err != nil {
			t.Fatalf("re-parsing NormalizedString()=%q of %q failed: %v", norm, s, err)
		}
		if cmp := v.Compare(v2); cmp != 0 {
			t.Fatalf("NewVersion(%q) != NewVersion(NormalizedString()=%q): Compare=%d", s, norm, cmp)
		}
		if norm2 := v2.NormalizedString(); norm2 != norm {
			t.Fatalf("NormalizedString not idempotent for %q: %q vs %q", s, norm, norm2)
		}
	})
}

func FuzzNormalizeComposerVersion(f *testing.F) {
	for _, s := range versionSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		norm, err := NormalizeComposerVersion(s)
		if err != nil {
			return
		}

		// The normalizer mirrors Composer at the string level. Two Composer
		// quirks mean we cannot assert stronger properties here:
		//   - it is NOT strictly idempotent for some date-style inputs
		//     ("0000-00" -> "0000.00" -> "0000.00.0.0"); and
		//   - it can emit a numeric segment wider than int64
		//     ("9227000000000000000"), which NewVersion later rejects.
		// The genuine safety property is simply that feeding the output back in
		// never panics. (Re-normalization may legitimately return an error.)
		_, _ = NormalizeComposerVersion(norm)
	})
}

func FuzzNewConstraint(f *testing.F) {
	for _, s := range constraintSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		c, err := NewConstraint(s)
		if err != nil {
			return
		}

		// A parsed constraint must be checkable against any valid version
		// without panicking.
		for _, vs := range []string{"1.0.0", "0.0.1", "2.5.3", "1.0.0-beta", "dev-main"} {
			v, verr := NewVersion(vs)
			if verr != nil {
				continue
			}
			_ = c.Check(v)
		}
	})
}

func FuzzSatisfies(f *testing.F) {
	// Seed with the cross product of a few versions and constraints.
	for _, vs := range []string{"1.2.3", "1.0.0-beta", "2.0.0", "dev-main"} {
		for _, cs := range []string{"^1.0", ">=1.0,<2.0", "*", "~1.2", "1.0 - 2.0"} {
			f.Add(vs, cs)
		}
	}

	f.Fuzz(func(t *testing.T, versionString, constraintString string) {
		got, err := Satisfies(versionString, constraintString)
		if err != nil {
			return
		}

		// Satisfies must agree with the lower-level NewConstraint/Check path.
		// Mirror Satisfies' own special-casing of empty/"*" constraints.
		v, verr := NewVersion(versionString)
		if verr != nil {
			t.Fatalf("Satisfies(%q, %q) succeeded but NewVersion failed: %v", versionString, constraintString, verr)
		}

		want := got
		if trimmed := strings.TrimSpace(constraintString); trimmed != "" && trimmed != "*" {
			c, cerr := NewConstraint(trimmed)
			if cerr != nil {
				t.Fatalf("Satisfies(%q, %q) succeeded but NewConstraint failed: %v", versionString, constraintString, cerr)
			}
			want = c.Check(v)
		} else {
			want = true
		}

		if got != want {
			t.Fatalf("Satisfies(%q, %q)=%v disagrees with NewConstraint().Check()=%v", versionString, constraintString, got, want)
		}
	})
}

func FuzzConstraintIntersects(f *testing.F) {
	for _, a := range constraintSeeds {
		for _, b := range []string{"^1.0", ">=1.0,<2.0", "1.2.*", "dev-main"} {
			f.Add(a, b)
		}
	}

	f.Fuzz(func(t *testing.T, left, right string) {
		// Both ConstraintIntersects and ConstraintSubsetOf must never panic on
		// arbitrary input; errors are an acceptable outcome.
		_, _ = ConstraintIntersects(left, right)
		_, _ = ConstraintSubsetOf(left, right)
	})
}
