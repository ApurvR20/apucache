package apucache

import (
	"hash/fnv"
	"sync"
)

const numShards = 16

// creating multiple maps(shards) with their own locks, so goroutines less likely wait on each other, while holding different keys
type shard struct {
	mu   sync.RWMutex
	data map[string]string
}

// router for shards
type Store struct {
	shards [numShards]*shard
}

func New() *Store {
	s := &Store{}
	for i := range s.shards {
		s.shards[i] = &shard{data: make(map[string]string)}
	}
	return s
}

// convert key into bytes, bytes into a number, %16, and pick one of 0-15 shards to slot the key, to spread the traffic across locks
func (s *Store) shardFor(key string) *shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return s.shards[h.Sum32()%numShards]
}

func (s *Store) Set(key, value string) {
	sh := s.shardFor(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	sh := s.shardFor(key)
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	v, ok := sh.data[key]
	return v, ok
}
