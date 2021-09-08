package caches

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound         = errors.New("bkit/caches: not found")
	ErrSizeOverCapacity = errors.New("bkit/caches: memory size over set the capacity")
)

func IsNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	// compatibility redis not found
	if err == ErrNotFound || err.Error() == "redis: nil" {
		return true
	}
	return false
}

type Cacher interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	SetWithTTL(key string, value interface{}, ttl time.Duration) error
	Del(key string) error

	Close() error
}

// redis.Client Cacher implement

type CacheStore interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	Del(key string) error
	Range(f func(key, value interface{}) bool)
}

var _ CacheStore = NewMemoryStore()

// MemoryStore represents in-memory store
type MemoryStore struct {
	store map[interface{}]interface{}
	mutex sync.RWMutex
}

// NewMemoryStore creates a new store in memory
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{store: make(map[interface{}]interface{})}
}

// Set sets object into store
func (s *MemoryStore) Set(key string, value interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.store[key] = value
	return nil
}

// Get gets object from store
func (s *MemoryStore) Get(key string) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if v, ok := s.store[key]; ok {
		return v, nil
	}
	return nil, ErrNotFound
}

// Del deletes object
func (s *MemoryStore) Del(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.store, key)
	return nil
}

func (s *MemoryStore) Range(f func(key, value interface{}) bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for k, v := range s.store {
		if !f(k, v) {
			break
		}
	}
}
