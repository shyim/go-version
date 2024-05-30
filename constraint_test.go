package version

import (
	"testing"
)

func TestConstraints(t *testing.T) {
	MustConstraints(NewConstraint(">=1.0.0"))
	MustConstraints(NewConstraint(">=1.0.0 || <2.0.0"))
	MustConstraints(NewConstraint(">=1.0.0,<2.0.0"))
}

func TestConstraintParsingWhitespaceAnd(t *testing.T) {
	c, err := NewConstraint(">=1.0 <2.0")
	if err != nil { t.Errorf("Expected no error, got %v", err) }

	if c.String() != ">=1.0,<2.0" { t.Errorf("Expected '>=1.0,<2.0', got %s", c.String()) }
	if !c.Check(Must(NewVersion("1.0.0"))) { t.Errorf("Expected true, got false") }
	if c.Check(Must(NewVersion("2.0.0"))) { t.Errorf("Expected false, got true") }
}

func TestConstraintParsingWhitespaceAndOr(t *testing.T) {
	c, err := NewConstraint("~6.4 >=6.4.20.0 || ~6.5")
	if err != nil { t.Errorf("Expected no error, got %v", err) }

	if c.String() != "~6.4,>=6.4.20.0||~6.5" { t.Errorf("Expected '~6.4,>=6.4.20.0||~6.5', got %s", c.String()) }
	if !c.Check(Must(NewVersion("6.4.20"))) { t.Errorf("Expected true, got false") }
	if !c.Check(Must(NewVersion("6.4.20.0"))) { t.Errorf("Expected true, got false") }
	if !c.Check(Must(NewVersion("6.5.0"))) { t.Errorf("Expected true, got false") }
	if c.Check(Must(NewVersion("6.4.0.0"))) { t.Errorf("Expected false, got true") }
}

func TestConstraintWithoutWhitespace(t *testing.T) {
	c, err := NewConstraint("<6.6.1.0||>=6.3.5.0")
	if err != nil { t.Errorf("Expected no error, got %v", err) }

	if c.String() != "<6.6.1.0||>=6.3.5.0" { t.Errorf("Expected '<6.6.1.0||>=6.3.5.0', got %s", c.String()) }
	if !c.Check(Must(NewVersion("6.4.0.0"))) { t.Errorf("Expected true, got false") }
}

func TestConstraintVersionNumber(t *testing.T) {
	c, err := NewConstraint("1.0.0")
	if err != nil { t.Errorf("Expected no error, got %v", err) }

	if c.String() != "1.0.0" { t.Errorf("Expected '1.0.0', got %s", c.String()) }
	if !c.Check(Must(NewVersion("1.0.0"))) { t.Errorf("Expected true, got false") }
	if c.Check(Must(NewVersion("1.0.1"))) { t.Errorf("Expected false, got true") }
}
