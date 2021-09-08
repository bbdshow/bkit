package caches

import (
	"container/list"
	"sync"
	"time"
)

const (
	GCInterval        = 60 * time.Minute
	CacheGcMaxRemoved = 30
)

type LRUCache struct {
	mutex      sync.Mutex
	eleList    *list.List
	eleIndex   map[string]*list.Element
	store      CacheStore
	MaxElement int
}

// NewLRUMemory memory store
func NewLRUMemory(maxElement int) *LRUCache {
	return NewLRUCache(NewMemoryStore(), maxElement)
}

// NewLRUCache lru cache maxElement min set 100
func NewLRUCache(store CacheStore, maxElement int) *LRUCache {
	m := &LRUCache{
		mutex:      sync.Mutex{},
		eleList:    list.New(),
		eleIndex:   make(map[string]*list.Element),
		store:      store,
		MaxElement: 100,
	}
	if maxElement > m.MaxElement {
		m.MaxElement = maxElement
	}
	m.runGC()
	return m
}

func (m *LRUCache) runGC() {
	go func() {
		for {
			time.Sleep(GCInterval)
			m.GC()
		}
	}()
}

//  if storage not enough, del little visit
func (m *LRUCache) GC() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.eleList.Len() < m.MaxElement {
		return
	}
	removedNum := 0
	// precedence del ttlKey
	for removedNum < CacheGcMaxRemoved {
		if !m.delTTLKey() {
			break
		}
		removedNum++
	}
	for removedNum < CacheGcMaxRemoved {
		if !m.delFrontKey() {
			break
		}
		removedNum++
	}
}

func (m *LRUCache) Get(key string) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if ele, ok := m.eleIndex[key]; ok {
		n := ele.Value.(*iNode)
		if n.Expired(time.Now().Unix()) {
			// expired
			m.eleList.Remove(ele)
			delete(m.eleIndex, key)
			_ = m.store.Del(key)
			return nil, ErrNotFound
		}
		// if visit, move to end, not GC
		n.lastVisit = time.Now()
		m.eleList.MoveToBack(ele)
		return m.store.Get(key)
	}
	return nil, ErrNotFound
}

func (m *LRUCache) Set(key string, val interface{}) error {
	return m.SetWithTTL(key, val, -1)
}

func (m *LRUCache) SetWithTTL(key string, val interface{}, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if ele, ok := m.eleIndex[key]; !ok {
		e := m.eleList.PushBack(newNode(key, ttl))
		m.eleIndex[key] = e
	} else {
		expired := int64(-1)
		if ttl > 0 {
			expired = time.Now().Add(ttl).Unix()
		}
		n := ele.Value.(*iNode)
		n.expired = expired
		n.lastVisit = time.Now()
	}
	_ = m.store.Set(key, val)

	if m.eleList.Len() > m.MaxElement {
		if !m.delTTLKey() {
			// if without ttl key, just front key
			m.delFrontKey()
		}
	}
	return nil
}

func (m *LRUCache) delTTLKey() bool {
	now := time.Now().Unix()
	for ele := m.eleList.Front(); ele != nil; {
		n := ele.Value.(*iNode)
		if n.Expired(now) {
			m.eleList.Remove(ele)
			delete(m.eleIndex, n.key)
			_ = m.store.Del(n.key)
			return true
		}
		ele = ele.Next()
	}
	return false
}

func (m *LRUCache) delFrontKey() bool {
	ele := m.eleList.Front()
	if ele == nil {
		return false
	}
	n := ele.Value.(*iNode)
	m.eleList.Remove(ele)
	delete(m.eleIndex, n.key)
	_ = m.store.Del(n.key)
	return true
}

func (m *LRUCache) Del(key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if ele, ok := m.eleIndex[key]; ok {
		m.eleList.Remove(ele)
		delete(m.eleIndex, key)
	}
	return m.store.Del(key)
}

func (m *LRUCache) Close() error {
	return nil
}

type iNode struct {
	key       string
	expired   int64
	lastVisit time.Time
}

func (n *iNode) Expired(now int64) bool {
	if n.expired == -1 {
		return false
	}
	return now > n.expired
}

func newNode(key string, ttl time.Duration) *iNode {
	n := &iNode{
		key:       key,
		expired:   -1,
		lastVisit: time.Now(),
	}
	if ttl > 0 {
		n.expired = time.Now().Add(ttl).Unix()
	}
	return n
}
