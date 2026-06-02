package version

import "testing"

// Representative inputs spanning the parser's distinct paths: short/full
// semver, prereleases, the v-prefix, dev branches, and date-style versions.
var benchVersions = []string{
	"1.2.3",
	"1.2.3.4",
	"v2.0.0",
	"1.0.0-beta.5",
	"1.0.0-alpha1",
	"2.1.x-dev",
	"dev-main",
	"20100102",
	"1.2.3+build.123",
}

var benchConstraints = []string{
	"1.2.3",
	"^1.0",
	"~1.2.3",
	">=1.0,<2.0",
	"1.2.*",
	"^1.0 || ^2.0",
	">=1.0@stable",
	"1.0 - 2.0",
}

func BenchmarkNewVersion(b *testing.B) {
	for _, v := range benchVersions {
		b.Run(v, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := NewVersion(v); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkNormalizeComposerVersion(b *testing.B) {
	for _, v := range benchVersions {
		b.Run(v, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := NormalizeComposerVersion(v); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkNewConstraint(b *testing.B) {
	for _, c := range benchConstraints {
		b.Run(c, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := NewConstraint(c); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkCompare(b *testing.B) {
	pairs := [][2]string{
		{"1.2.3", "1.2.3"},        // equal fast path
		{"1.2.3", "1.2.4"},        // differing segment
		{"1.0.0-beta", "1.0.0"},   // prerelease vs stable
		{"1.0.0-alpha", "1.0.0-beta"}, // prerelease ordering
		{"1.2.3", "1.2.3.0"},      // jagged specificity
	}
	for _, p := range pairs {
		left := Must(NewVersion(p[0]))
		right := Must(NewVersion(p[1]))
		b.Run(p[0]+"_vs_"+p[1], func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = left.Compare(right)
			}
		})
	}
}

func BenchmarkConstraintCheck(b *testing.B) {
	for _, c := range benchConstraints {
		constraint, err := NewConstraint(c)
		if err != nil {
			b.Fatal(err)
		}
		v := Must(NewVersion("1.2.3"))
		b.Run(c, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = constraint.Check(v)
			}
		})
	}
}

// BenchmarkSatisfies measures the full parse-and-check path that most callers
// actually use, including version and constraint parsing on every call.
func BenchmarkSatisfies(b *testing.B) {
	cases := [][2]string{
		{"1.2.3", "^1.0"},
		{"1.2.3", ">=1.0,<2.0"},
		{"1.0.0-beta", "~1.0"},
		{"2.5.0", "1.2.*"},
	}
	for _, tc := range cases {
		b.Run(tc[0]+"_"+tc[1], func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := Satisfies(tc[0], tc[1]); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkConstraintIntersects(b *testing.B) {
	cases := [][2]string{
		{"^1.0", ">=1.0,<2.0"},
		{"^6.4", ">=6.4.0,<6.4.29"},
		{"1.2.*", ">=1.2.5,<1.3.0"},
		{"^1.0 || ^2.0", "^1.5"},
	}
	for _, tc := range cases {
		b.Run(tc[0]+"_"+tc[1], func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := ConstraintIntersects(tc[0], tc[1]); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
