package apucache

import "testing"

// struct T from testing package
func TestSetGet(t *testing.T) {
	s := New(100) //can call any function in same package
	s.Set("a", "1")

	v, ok := s.Get("a")
	if !ok || v != "1" {
		t.Fatalf("expected 1, got %q (ok=%v)", v, ok) //fails test, prints error, stops it.
		//f = string has formatted values %something
	}

	if _, ok := s.Get("missing"); ok {
		t.Fatal("expected missing key to not exist")
	} //should be false, if it exists, something wrong
}
