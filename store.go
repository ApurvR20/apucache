package apucache

import "sync"

type Store struct {
	lock sync.RWMutex
	data map[string]string
}

func New() *Store {
	return &Store{data: make(map[string]string)} //literal helps fill data field with a map, not nil
}

func (s *Store) Set(k, v string) { //Only one goroutine can set
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[k] = v
}

func (s *Store) Get(k string) (string, bool) { //Only one goroutine can set
	s.lock.RLock() //allows shared read locks, but no writes
	defer s.lock.RUnlock()
	v, ok := s.data[k]
	return v, ok
}
