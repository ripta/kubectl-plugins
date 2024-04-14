package dynaward

import (
	"bytes"
	"sync"

	"k8s.io/utils/lru"
)

type TraceStore struct {
	cache *lru.Cache
	ids   map[string]struct{}
	mut   sync.RWMutex
}

func NewTraceStore(size int) *TraceStore {
	t := &TraceStore{
		ids: map[string]struct{}{},
		mut: sync.RWMutex{},
	}

	t.cache = lru.NewWithEvictionFunc(size, t.onEvict)
	return t
}

func (t *TraceStore) Add(id string, trace *RoundTripTrace) {
	t.cache.Add(id, trace)

	t.mut.Lock()
	defer t.mut.Unlock()
	t.ids[id] = struct{}{}
}

func (t *TraceStore) Clear() {
	// no lock necessary, as it goes through eviction and calls onEvict
	t.cache.Clear()
}

func (t *TraceStore) Get(id string) *RoundTripTrace {
	t.mut.RLock()
	defer t.mut.RUnlock()

	trace, ok := t.cache.Get(id)
	if !ok {
		return nil
	}

	return trace.(*RoundTripTrace)
}

func (t *TraceStore) List() []string {
	t.mut.RLock()
	defer t.mut.RUnlock()

	ids := []string{}
	for id := range t.ids {
		ids = append(ids, id)
	}

	return ids
}

func (t *TraceStore) onEvict(key lru.Key, _ interface{}) {
	t.mut.Lock()
	defer t.mut.Unlock()
	delete(t.ids, key.(string))
}

type RoundTripTrace struct {
	Host     string
	Request  *bytes.Buffer
	Response *bytes.Buffer
}

func NewRoundTripTrace() *RoundTripTrace {
	return &RoundTripTrace{
		Request:  &bytes.Buffer{},
		Response: &bytes.Buffer{},
	}
}
