package ufo

import "testing"

func TestBoxRange_accessorsAndToString(t *testing.T) {
	r := NewBoxRange(3, 7)
	if r.GetStart() != 3 || r.GetEnd() != 7 {
		t.Fatalf("got start=%d end=%d, want 3 and 7", r.GetStart(), r.GetEnd())
	}
	if got, want := r.ToString(), "[start=3, end=7]"; got != want {
		t.Errorf("ToString() = %q, want %q", got, want)
	}
	if got, want := r.String(), "[start=3, end=7]"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
