package validators

import (
	"testing"
)

func TestPasswordValidator_Valid(t *testing.T) {
	v := DefaultPasswordValidator()
	// Default requires uppercase, lowercase, digits, min 8
	if err := v.Validate("Abcd1234"); err != nil {
		t.Fatalf("expected valid password, got error: %v", err)
	}
}

func TestPasswordValidator_InvalidShort(t *testing.T) {
	v := DefaultPasswordValidator()
	if err := v.Validate("A1b"); err == nil {
		t.Fatalf("expected error for short password")
	}
}

func TestContainsSQLInjection(t *testing.T) {
	if !ContainsSQLInjection("select * from users") {
		t.Fatalf("expected sql injection detected")
	}
	if ContainsSQLInjection("normaltext@example.com") {
		t.Fatalf("did not expect sql injection for normal text")
	}
}

func TestContainsXSS(t *testing.T) {
	if !ContainsXSS("<script>alert(1)</script>") {
		t.Fatalf("expected xss detected")
	}
	if ContainsXSS("hello world") {
		t.Fatalf("did not expect xss for normal text")
	}
}
