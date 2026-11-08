/*
Tradeoffs

- There’s no fsync, so the last few writes can be lost if the machine loses power (not if just the process dies).
- The log only grows. Compaction or snapshots would fix that.
- Evicted keys come back on replay, then get evicted again as the log is read.
*/

package apucache

import (
	"encoding/json"
	"os"
	"sync"
)

type record struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

type aof struct {
	mu   sync.Mutex
	file *os.File
	enc  *json.Encoder
}

/*
- Creates the Store, then opens the file (create if missing, append mode).
- Replay loop: decodes one record at a time and calls set for each, until decoding fails at end of file.
- good tracks the end of the last complete record. If a crash left a half-written record at the end, Truncate(good) cuts it off so the log stays clean.
- Attaches the log to the store and returns it.
*/

func Open(capacity int, path string) (*Store, error) {
	s := New(capacity)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	var good int64
	dec := json.NewDecoder(f)
	for {
		var r record
		if err := dec.Decode(&r); err != nil {
			break
		}
		s.set(r.Key, r.Value)
		good = dec.InputOffset()
	}
	if err := f.Truncate(good); err != nil {
		f.Close()
		return nil, err
	}

	s.log = &aof{file: f, enc: json.NewEncoder(f)}
	return s, nil
}

func (s *Store) Set(key, value string) {
	s.set(key, value)
	if s.log != nil {
		s.log.mu.Lock()
		s.log.enc.Encode(record{key, value})
		s.log.mu.Unlock()
	}
}

func (s *Store) Close() error {
	if s.log == nil {
		return nil
	}
	return s.log.file.Close()
}
