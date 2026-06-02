# go-version

[![Go Reference](https://pkg.go.dev/badge/github.com/shyim/go-version.svg)](https://pkg.go.dev/github.com/shyim/go-version)
[![Go](https://github.com/shyim/go-version/actions/workflows/go.yml/badge.svg)](https://github.com/shyim/go-version/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.23-blue)](https://go.dev/dl/)

A zero-dependency Go library that ports the **PHP Composer** version and constraint parsing semantics to Go — including caret (`^`), tilde (`~`), wildcards (`*`), hyphen ranges, stability flags, and prerelease comparison.

Forked from [hashicorp/go-version](https://github.com/hashicorp/go-version) and fundamentally extended with full Composer compatibility.

## Installation

```bash
go get github.com/shyim/go-version
```

Requires Go 1.23+.

## Quick Start

```go
import "github.com/shyim/go-version"

// Check if a version satisfies a constraint
ok, err := version.Satisfies("1.2.3", "^1.0")   // true
ok, err  = version.Satisfies("7.0.0", "^1.0")   // false

// Parse a version
v, err := version.NewVersion("1.2.3-beta1")
fmt.Println(v.Major())            // 1
fmt.Println(v.Minor())            // 2
fmt.Println(v.Patch())            // 3
fmt.Println(v.Prerelease())       // "beta1"
fmt.Println(v.NormalizedString()) // "1.2.3.0-beta1"

// Compare versions
a := version.Must(version.NewVersion("1.2.3-rc2"))
b := version.Must(version.NewVersion("1.2.3-rc1"))
a.GreaterThan(b)  // true — numeric suffixes compared correctly

// Parse a constraint with multiple conditions
c, err := version.NewConstraint(">=1.0,<2.0")
c.Check(v)  // true

// Check constraint intersection (for dependency solvers)
ok, err = version.ConstraintIntersects("^1.0", ">=1.0,<2.0")  // true
ok, err = version.ConstraintIntersects("^1.0", "^2.0")         // false

// Check subset relationship
ok, err = version.ConstraintSubsetOf("^1.2", "^1.0 || ^2.0")   // true
```

## API Reference

### Version Parsing & Comparison

| Function / Method | Description |
|---|---|
| `NewVersion(v string) (*Version, error)` | Parse a version string (supports Composer formats) |
| `Must(v *Version, err error) *Version` | Panic-on-error convenience wrapper |
| `v.Compare(other *Version) int` | Compare two versions: -1, 0, 1 |
| `v.Equal(other *Version) bool` | Exact equality |
| `v.GreaterThan(other *Version) bool` | Greater-than |
| `v.GreaterThanOrEqual(other *Version) bool` | Greater-than-or-equal |
| `v.LessThan(other *Version) bool` | Less-than |
| `v.LessThanOrEqual(other *Version) bool` | Less-than-or-equal |
| `v.Major() int` | Major version segment |
| `v.Minor() int` | Minor version segment |
| `v.Patch() int` | Patch version segment |
| `v.Segments() []int` | All numeric segments as `[]int` |
| `v.Segments64() []int64` | All numeric segments as `[]int64` |
| `v.Prerelease() string` | Prerelease identifier (e.g. `"beta2"`) |
| `v.IsPrerelease() bool` | Whether the version has prerelease info |
| `v.NormalizedString() string` | Canonical string representation |
| `v.Original() string` | Original parsed string |
| `v.IncreaseMajor()` | Bump major, reset minor/patch/build |
| `v.IncreaseMinor()` | Bump minor, reset patch/build |
| `v.IncreasePatch()` | Bump patch, reset build |

### Constraint Parsing & Checking

| Function / Method | Description |
|---|---|
| `NewConstraint(cs string) (Constraints, error)` | Parse a constraint string |
| `MustConstraints(c Constraints, err error) Constraints` | Panic-on-error convenience wrapper |
| `cs.Check(v *Version) bool` | Test if a version satisfies the constraints |
| `cs.String() string` | String representation of constraints |
| `c.Check(v *Version) bool` | Test a single constraint against a version |
| `c.Prerelease() bool` | Whether the constraint target has a prerelease |
| `c.String() string` | Original constraint string |

### Convenience API

| Function | Description |
|---|---|
| `Satisfies(version, constraint string) (bool, error)` | One-shot: parse version, parse constraint, check |
| `NormalizeComposerVersion(version string) (string, error)` | Normalize a Composer version string |
| `Stability(version string) string` | Returns `"dev"`, `"alpha"`, `"beta"`, `"RC"`, or `"stable"` |

### Constraint Intersection (for dependency solvers)

| Function | Description |
|---|---|
| `ConstraintIntersects(left, right string) (bool, error)` | Do two constraints share at least one version? |
| `ConstraintSubsetOf(left, right string) (bool, error)` | Does `left`'s version set fall entirely within `right`'s? |

### Sorting

| Type | Description |
|---|---|
| `Collection []*Version` | Implements `sort.Interface` for stable version sorting |

### Stability Constants

```go
const (
    StabilityDev    = "dev"
    StabilityAlpha  = "alpha"
    StabilityBeta   = "beta"
    StabilityRC     = "RC"
    StabilityStable = "stable"
)
```

## Feature Overview

### Version Formats

The library parses a wide range of version strings following Composer's conventions:

| Format | Example | Notes |
|---|---|---|
| Classic semver | `1.2.3`, `v1.2.3`, `1.2.3.4` | Padded to 4 segments |
| Prereleases | `1.0.0-alpha1`, `1.0.0-beta.2`, `1.0.0-RC1` | Abbreviations: `a1`, `b2`, `p1`/`pl3` |
| Build metadata | `1.2.3+build.123` | Stripped, does not affect comparison |
| Date versions | `20100102`, `2010.01.02`, `201903.0` | Composer's CalVer support |
| dev branches | `dev-main`, `dev-feature/x` | Treated as unordered (equality only) |
| Numeric branches | `2.1.x-dev`, `1.*-dev` | Wildcards mapped to `9999999` |
| Stability suffixes | `1.0.0@beta`, `1.2.3@stable` | Stripped during normalization |
| Aliases | `1.2.3 as 1.2.3-alias` | Source version extracted |

### Constraint Operators

| Operator | Example | Meaning |
|---|---|---|
| `=` / `==` | `=1.2.3` | Exact version match |
| `!=` / `<>` | `!=1.2.3` | Not equal |
| `>` | `>1.0` | Greater than |
| `>=` | `>=1.0` | Greater than or equal |
| `<` | `<2.0` | Less than |
| `<=` | `<=2.0` | Less than or equal |
| `^` | `^1.2.3` | Caret: compatible updates (Composer-style) |
| `~` | `~1.2.3` | Tilde: next significant release |
| `*` | `*`, `2.*`, `1.2.*` | Wildcard matching |
| (none) | `1.2.3` | Bare version = exact match |
| `-` | `1.0 - 2.0` | Hyphen range |

### Composer-Specific Semantics

**Prerelease rank ordering** — versions sort in Composer's stability order:
`dev < alpha < beta < RC < stable`

With numeric suffixes compared numerically: `1.0.0-rc1 < 1.0.0-rc2 < 1.0.0-rc10`

**Caret (`^`)** — allows changes that do not modify the leftmost non-zero digit:

| Constraint | Equivalent Range | Rationale |
|---|---|---|
| `^1.2.3` | `>=1.2.3 <2.0.0` | Major > 0, lock major |
| `^0.2.3` | `>=0.2.3 <0.3.0` | Major = 0, minor > 0, lock minor |
| `^0.0.3` | `>=0.0.3 <0.0.4` | Both zero, lock patch |
| `^1.2` | `>=1.2.0 <2.0.0` | Equivalent to `^1.2.0` |

**Tilde (`~`)** — allows changes at the last specified segment:

| Constraint | Equivalent Range | Rationale |
|---|---|---|
| `~1.2.3` | `>=1.2.3 <1.3.0` | 3 segments: lock minor |
| `~1.2` | `>=1.2.0 <2.0.0` | 2 segments: lock major |
| `~1.2.3.4` | `>=1.2.3.4 <1.2.4.0` | 4 segments: lock patch |

**Hyphen ranges** — Composer-style range with intelligent upper bound logic:

| Constraint | Equivalent |
|---|---|
| `1.0 - 2.0` | `>=1.0 <2.1` |
| `1.2.3 - 2.3.4` | `>=1.2.3 <=2.3.4` |
| `1.0 - 2.0.*` | `>=1.0 <2.1` |

**Stability flags** (`@stable`, `@beta`, `@dev`, `@alpha`, `@RC`):
```go
c := version.MustConstraints(version.NewConstraint(">=1.0@stable"))
c.Check(version.Must(version.NewVersion("1.0.0")))       // true
c.Check(version.Must(version.NewVersion("1.0.0-beta")))  // false
```

**Stability-appended operators** — `>=1.0@beta` expands to `>=1.0.0-beta`, making the minimum-stable constraint automatically aware of its stability level.

**Branch versions** (`dev-*`) — treated as unordered for ordering operators; only equality (`=`/`==`) and inequality (`!=`) give meaningful results. Wildcards (`*`) and `anyBranch` domains include all branch versions.

**Constraint composition** — supports both `||` and `|` as OR separators, with `,` and whitespace as AND:
- `>=1.0,<2.0` — AND: version must satisfy both
- `^1.0 || ^2.0` — OR: version must satisfy at least one
- `>=1.0 <2.0` — AND (whitespace delimiter)

## Differences from hashicorp/go-version

| Feature | hashicorp/go-version | shyim/go-version |
|---|---|---|
| Version model | Strict semver (MAJOR.MINOR.PATCH) | Composer (multi-segment, branches, date versions) |
| Operators | `=`, `!=`, `>`, `<`, `>=`, `<=` | + `^`, `~`, `*`, hyphen ranges |
| Prerelease comparison | Basic string compare | Ranked: dev < alpha < beta < RC < stable, numeric suffixes |
| Caret / tilde | Not supported | Full Composer semantics |
| Stability flags | Not supported | `@stable`, `@beta`, `@alpha`, `@dev`, `@RC` |
| Branch versions | Not supported | `dev-*` with equality-only semantics |
| Constraint intersection | Not supported | Full domain algebra |
| Wildcards | `*` only | `1.*`, `1.2.*`, `2.x`, `x.x.x` |
| Normalization | None | Aliases, stability flags, branches, date versions |
| Dependencies | None | None (preserved) |

## Performance

The library includes an optimized fast path for the common case of plain numeric versions (`1.2.3`, `v1.2`), avoiding regexp overhead. Benchmarks are available:

```bash
go test -bench=. -benchmem ./...
```

## Testing

The library has comprehensive test coverage:

| Area | Tests |
|---|---|
| Version parsing | 24 test functions, table-driven + property-based |
| Constraint checking | 27 test functions, 900+ operator-matrix combinations |
| Normalization | 2 test functions + fast-path parity verification |
| Constraint intersection | 7 test functions + exhaustive domain algebra tests |
| API surface | 4 test functions |
| Semver behavior | 35 test functions, 100KB+ conformance suite |

Fuzz tests are included for all public entry points:

```bash
go test -run=Fuzz -fuzz=FuzzNewVersion -fuzztime=30s ./...
go test -run=Fuzz -fuzz=FuzzNewConstraint -fuzztime=30s ./...
go test -run=Fuzz -fuzz=FuzzSatisfies -fuzztime=30s ./...
go test -run=Fuzz -fuzz=FuzzConstraintIntersects -fuzztime=30s ./...
```

Run all tests:

```bash
go test ./...
```

## Architecture

The library follows a clean two-phase pipeline:

```
normalizer.go → version.go → constraint.go → intersect.go
                         ↕
              version_collection.go
```

- **`normalizer.go`** — Canonicalizes raw version strings (aliases, stability suffixes, branches, date versions, classical semver modifiers)
- **`version.go`** — `Version` type, regex parsing, ranked prerelease comparison
- **`constraint.go`** — Disjunctive Normal Form constraint model, all operators, hyphen ranges, stability constraints
- **`domain.go`** / **`interval.go`** / **`snapshot.go`** / **`intersect.go`** — Domain algebra for constraint intersection and subset queries
- **`version_collection.go`** — `sort.Interface` for version slices
- **`api.go`** — Public convenience API (`Satisfies`, `NormalizeComposerVersion`, `Stability`)

Zero external dependencies.

## License

MIT — see [LICENSE](LICENSE) for details. Forked from [hashicorp/go-version](https://github.com/hashicorp/go-version) (also MIT licensed).
