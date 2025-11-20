package memkv

import (
	"sync"
	"time"
)

type item struct {
	value     interface{}
	expiresAt time.Time
}

type InMemoryStore struct {
	data              map[string]item
	defaultExpiration time.Duration
	mu                sync.RWMutex
	stopJanitor       chan struct{}
}

var Store *InMemoryStore

func init() {
	Store = NewStore(10 * time.Minute)
}

func NewStore(defaultExpiration time.Duration) *InMemoryStore {
	s := &InMemoryStore{
		data:              make(map[string]item),
		defaultExpiration: defaultExpiration,
		stopJanitor:       make(chan struct{}),
	}

	go s.janitor(1 * time.Minute)

	return s
}

func (s *InMemoryStore) Set(key string, value interface{}, expiration ...time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exp := s.defaultExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	var expiresAt time.Time
	if exp > 0 {
		expiresAt = time.Now().Add(exp)
	}

	s.data[key] = item{
		value:     value,
		expiresAt: expiresAt,
	}
}

func (s *InMemoryStore) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	it, found := s.data[key]
	s.mu.RUnlock()

	if !found {
		return nil, false
	}

	if !it.expiresAt.IsZero() && time.Now().After(it.expiresAt) {
		s.Delete(key)
		return nil, false
	}

	return it.value, true
}

func (s *InMemoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *InMemoryStore) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopJanitor:
			return
		}
	}
}

func (s *InMemoryStore) cleanup() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for k, it := range s.data {
		if !it.expiresAt.IsZero() && now.After(it.expiresAt) {
			delete(s.data, k)
		}
	}
}

func (s *InMemoryStore) Close() {
	close(s.stopJanitor)
}
