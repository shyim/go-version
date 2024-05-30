# Versioning Library for Go

This is a fork of [hashicorp/go-version](https://github.com/hashicorp/go-version) with some improvments.


## Difference to the original library

- Added caret (^) and tilde (~) support
- Comparing of pre-release versions (e.g. 1.0.0-alpha.1 < 1.0.0-alpha.2)
- Constriant to a specific version (e.g. 1.0.0) without operators
- Allow to parse constraints with only one delimitier `|` instead of `||`
