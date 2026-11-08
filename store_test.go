package apucache

import (
	"strconv"
	"testing"
)

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

func TestEviction(t *testing.T) {
	s := New(32)
	for i := 0; i < 1000; i++ { //try to set 1000 keys in store of 32
		s.Set(strconv.Itoa(i), "v")
	}

	count := 0
	for i := 0; i < 1000; i++ { // count no of current keys
		if _, ok := s.Get(strconv.Itoa(i)); ok {
			count++
		}
	}
	if count > 32 {
		t.Fatalf("expected at most 32 entries, got %d", count)
	}

	if _, ok := s.Get("999"); !ok {
		t.Fatal("expected the newest key to still be present")
	}
	if _, ok := s.Get("0"); ok {
		t.Fatal("expected the oldest key to be evicted")
	}
}
