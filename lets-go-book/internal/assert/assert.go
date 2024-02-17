package assert

import (
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, got, want T) {
	// This line is needed to tell the test suite that this method is a helper.
	// This way when it fails the line number reported will be in our function call 
	// rather than inside this helper (which is not helpful).
	t.Helper() 

	if got != want {
		t.Errorf("got %v | want %v", got, want)
	}
}

func StringContains(t *testing.T, s, substr string) {
	t.Helper()

	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}