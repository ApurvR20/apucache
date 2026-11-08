package apucache

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
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

func TestRecovery(t *testing.T) { // does data survive a restart
	path := filepath.Join(t.TempDir(), "cache.aof")

	s, err := Open(100, path)
	if err != nil {
		t.Fatal(err)
	}
	s.Set("a", "1")
	s.Set("b", "2")
	s.Close()

	s2, err := Open(100, path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	if v, ok := s2.Get("a"); !ok || v != "1" {
		t.Fatalf("expected a=1 after restart, got %q (ok=%v)", v, ok)
	}
	if v, ok := s2.Get("b"); !ok || v != "2" {
		t.Fatalf("expected b=2 after restart, got %q (ok=%v)", v, ok)
	}
}

func TestTruncatedTail(t *testing.T) { //does a half-written record get handled while writing to log
	path := filepath.Join(t.TempDir(), "cache.aof")

	s, _ := Open(100, path)
	s.Set("a", "1")
	s.Close()

	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString(`{"k":"b","v":`)
	f.Close()

	s2, err := Open(100, path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	if v, ok := s2.Get("a"); !ok || v != "1" {
		t.Fatalf("expected a=1, got %q (ok=%v)", v, ok)
	}
	if _, ok := s2.Get("b"); ok {
		t.Fatal("half-written record should have been discarded")
	}
}

/*It starts 8 goroutines that all `Set` and `Get` the same 100 keys 1000 times, then waits for them to finish. It has no assertions. Its job is to create real concurrent traffic, so `go test -race` can fail it if any access isn't protected by a lock.*/

func TestConcurrent(t *testing.T) {
	s := New(1000)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				k := strconv.Itoa(i % 100)
				s.Set(k, "v")
				s.Get(k)
			}
		}()
	}
	wg.Wait()
}
