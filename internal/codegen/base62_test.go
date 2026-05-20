package codegen

import "testing"

func TestRandomBase62(t *testing.T) {
	code, err := RandomBase62(7)
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 7 {
		t.Fatalf("expected length 7, got %d", len(code))
	}
}
