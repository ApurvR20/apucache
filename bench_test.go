package apucache

import (
	"strconv"
	"testing"
)

func benchmarkStore(b *testing.B, shards int) {
	s := newStore(100000, shards)
	b.RunParallel(func(pb *testing.PB) { //runParellel : testing helper
		i := 0
		for pb.Next() { //pb.Next() returns true while there is still work to do and false when Go has measured enough, measured by time
			k := strconv.Itoa(i % 1000)
			if i%10 == 0 {
				s.set(k, "v")
			} else {
				s.Get(k)
			}
			i++
		}
	})
}

func BenchmarkOneLock(b *testing.B)       { benchmarkStore(b, 1) }
func BenchmarkSixteenShards(b *testing.B) { benchmarkStore(b, 16) }
