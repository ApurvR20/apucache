package apucache

import (
	"container/list"
	"hash/fnv"
	"sync"
)

const numShards = 16

type entry struct {
	key   string
	value string
}

type shard struct {
	mu    sync.Mutex //plain mutex instead of RW. Get changes order, so reads are writes as well
	data  map[string]*list.Element
	order *list.List
	cap   int
}

type Store struct {
	shards [numShards]*shard //array of maps. hash the req, %16, search the correspoding shard
	log    *aof
}

func New(capacity int) *Store { //splits cap across 16 shards, and assigns default data
	perShard := capacity / numShards
	if perShard < 1 {
		perShard = 1
	}
	s := &Store{}
	for i := range s.shards {
		s.shards[i] = &shard{
			data:  make(map[string]*list.Element),
			order: list.New(),
			cap:   perShard,
		}
	}
	return s
}

func (s *Store) shardFor(key string) *shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return s.shards[h.Sum32()%numShards]
}

func (s *Store) set(key, value string) {
	sh := s.shardFor(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	if el, ok := sh.data[key]; ok { //move the found key:value to front of list
		el.Value.(*entry).value = value
		sh.order.MoveToFront(el)
		return
	}

	sh.data[key] = sh.order.PushFront(&entry{key, value})
	if sh.order.Len() > sh.cap { //if cap is exceeded, we get the LRU node, get its key from entry (node stores both), node is removed from list,
		oldest := sh.order.Back()
		sh.order.Remove(oldest)
		delete(sh.data, oldest.Value.(*entry).key)
	}
}

func (s *Store) Get(key string) (string, bool) {
	sh := s.shardFor(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	el, ok := sh.data[key]
	if !ok {
		return "", false
	}
	sh.order.MoveToFront(el)
	return el.Value.(*entry).value, true
}
