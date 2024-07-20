package bkit

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrCacheNotFound         = errors.New("bkit/caches: not found")
	ErrCacheSizeOverCapacity = errors.New("bkit/caches: memory size over set the capacity")
)

func IsNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	// compatibility redis not found
	if err == ErrCacheNotFound || err.Error() == "redis: nil" {
		return true
	}
	return false
}

type Cacher interface {
	Get(key string) (interface{}, error)
	Range(f func(key string, value interface{}) bool) // Range Locks will be added. Watch for deadlocks

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

const (
	CacheGCInterval   = 60 * time.Minute
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
			time.Sleep(CacheGCInterval)
			m.GC()
		}
	}()
}

// GC if storage not enough, del little visit
func (m *LRUCache) GC() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	_ = m.delAllExpiredKey()

	if m.eleList.Len() < m.MaxElement {
		return
	}
	removedNum := 0
	for removedNum < CacheGcMaxRemoved {
		if !m.delFrontKey() {
			break
		}
		removedNum++
	}
}

func (m *LRUCache) Get(key string) (interface{}, error) {
	m.mutex.Lock()
	v, err := m.get(key, false)
	m.mutex.Unlock()
	return v, err
}

// goroutine not safety
func (m *LRUCache) get(key string, isRange bool) (interface{}, error) {
	if ele, ok := m.eleIndex[key]; ok {
		n := ele.Value.(*cacheNode)
		if n.Expired(time.Now().Unix()) {
			// expired
			m.eleList.Remove(ele)
			delete(m.eleIndex, key)
			_ = m.store.Del(key)
			return nil, ErrNotFound
		}
		// if visit, move to end, not GC
		// if isRange , not change visit state
		if !isRange {
			n.lastVisit = time.Now()
			m.eleList.MoveToBack(ele)
		}
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
		e := m.eleList.PushBack(newCacheNode(key, ttl))
		m.eleIndex[key] = e
	} else {
		expired := int64(-1)
		if ttl > 0 {
			expired = time.Now().Add(ttl).Unix()
		}
		n := ele.Value.(*cacheNode)
		n.expired = expired
		n.lastVisit = time.Now()
	}
	_ = m.store.Set(key, val)

	if m.eleList.Len() > m.MaxElement {
		delCount := m.delAllExpiredKey()
		if delCount <= 0 {
			// if without ttl key, just front key
			m.delFrontKey()
		}
	}
	return nil
}

func (m *LRUCache) Range(fn func(key string, value interface{}) bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for k := range m.eleIndex {
		v, err := m.get(k, true)
		if err == nil {
			next := fn(k, v)
			if !next {
				return
			}
		}
	}
}

// delete all expired key
func (m *LRUCache) delAllExpiredKey() int {
	delCount := 0
	now := time.Now().Unix()
	for ele := m.eleList.Front(); ele != nil; {
		n := ele.Value.(*cacheNode)
		if n.Expired(now) {
			m.eleList.Remove(ele)
			delete(m.eleIndex, n.key)
			_ = m.store.Del(n.key)
			delCount++
		}
		ele = ele.Next()
	}
	return delCount
}

func (m *LRUCache) delFrontKey() bool {
	ele := m.eleList.Front()
	if ele == nil {
		return false
	}
	n := ele.Value.(*cacheNode)
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

type cacheNode struct {
	key       string
	expired   int64
	lastVisit time.Time
}

func (n *cacheNode) Expired(now int64) bool {
	if n.expired == -1 {
		return false
	}
	return now > n.expired
}

func newCacheNode(key string, ttl time.Duration) *cacheNode {
	n := &cacheNode{
		key:       key,
		expired:   -1,
		lastVisit: time.Now(),
	}
	if ttl > 0 {
		n.expired = time.Now().Add(ttl).Unix()
	}
	return n
}

// LimitMemoryCache limit Memory use, implement cacher, suitable high performance, storage small.
type LimitMemoryCache struct {
	rwMutex sync.RWMutex
	store   map[string]IValue

	// storage size， not limit = -1
	size        int32
	currentSize int32

	// cache value, save write file
	filename string
}

// NewMemCache
// size=-1 not limit,   1<<20(1MB)
// filename if pass in, Close will save to file
func NewLimitMemoryCache(size int32, filename ...string) *LimitMemoryCache {
	m := &LimitMemoryCache{
		store:       make(map[string]IValue),
		size:        -1,
		currentSize: 0,
		filename:    "",
	}
	if size > 0 {
		m.size = size
	}
	if len(filename) > 0 && filename[0] != "" {
		m.filename = filename[0]
	}

	if err := m.load(); err != nil {
		log.Printf("WARNING: load file cache error %s \n", err.Error())
	}

	m.runGC()

	return m
}

type IValue struct {
	Value       interface{} `json:"v"`
	ExpiredTime time.Time   `json:"e"`
	Size        int32       `json:"s"`
}

func (val *IValue) SetExpiredTime(t time.Duration) {
	if t <= -1 {
		val.ExpiredTime = time.Now().AddDate(100, 0, 0)
		return
	}
	val.ExpiredTime = time.Now().Add(t)
}

func (val *IValue) TTL() time.Duration {
	expire := time.Until(val.ExpiredTime)
	if expire < 0 {
		expire = 0
	}
	return expire
}

func (val *IValue) Expired(now int64) bool {
	return val.ExpiredTime.Unix() < now
}

func (m *LimitMemoryCache) Get(key string) (interface{}, error) {
	m.rwMutex.RLock()
	v, err := m.get(key)
	m.rwMutex.RUnlock()
	return v, err
}

// goroutine not safety
func (m *LimitMemoryCache) get(key string) (interface{}, error) {
	iVal, ok := m.store[key]
	if ok {
		// if expired, just del
		if iVal.Expired(time.Now().Unix()) {
			m.delete(key, true)
			return nil, ErrNotFound
		}
		return iVal.Value, nil
	}
	return nil, ErrNotFound
}

func (m *LimitMemoryCache) Range(f func(key string, value interface{}) bool) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()

	for k := range m.store {
		v, err := m.get(k)
		if err == nil {
			next := f(k, v)
			if !next {
				return
			}
		}
	}
}

func (m *LimitMemoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	val := IValue{
		Value: value,
	}
	if m.size > 0 {
		val.Size = int32(len(key) + len(fmt.Sprint(value)))
	}
	val.SetExpiredTime(ttl)

	if err := m.set(key, val); err != nil {
		return err
	}

	return nil
}

func (m *LimitMemoryCache) Set(key string, value interface{}) error {
	return m.SetWithTTL(key, value, -1)
}

func (m *LimitMemoryCache) set(key string, val IValue) error {
	m.rwMutex.Lock()
	addSize := int32(0)
	oldVal, ok := m.store[key]
	if !ok {
		if err := m.isOverSize(val.Size); err != nil {
			return err
		}
		addSize = val.Size
	} else {
		// calc storage size
		subSize := val.Size - oldVal.Size
		if err := m.isOverSize(subSize); err != nil {
			return err
		}
		addSize = subSize
	}
	m.store[key] = val

	atomic.AddInt32(&m.currentSize, addSize)

	m.rwMutex.Unlock()

	return nil
}

func (m *LimitMemoryCache) Del(key string) error {
	m.delete(key, false)
	return nil
}

// return  data size
func (m *LimitMemoryCache) delete(key string, isExpired bool) int32 {
	var size int32
	m.rwMutex.Lock()
	val, ok := m.store[key]
	if ok {
		if (isExpired && val.Expired(time.Now().Unix())) || !isExpired {
			// if is expired del action, verify expired time
			// if not expired del action, just del
			delete(m.store, key)
			atomic.AddInt32(&m.currentSize, -val.Size)
			size = val.Size
		}
	}
	m.rwMutex.Unlock()
	return size
}

// Close if enable save to file, save()
func (m *LimitMemoryCache) Close() error {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	if m.filename != "" {
		err := m.save()
		m.store = make(map[string]IValue)
		return err
	}
	return nil
}

func (m *LimitMemoryCache) isOverSize(size int32) error {
	if m.size <= 0 {
		return nil
	}

	if atomic.LoadInt32(&m.currentSize)+size > m.size {
		// scan expired del, just del
		if m.scanExpiredKeyAndDel() >= size {
			return nil
		}
		return ErrCacheSizeOverCapacity
	}
	return nil
}

func (m *LimitMemoryCache) scanExpiredKeyAndDel() int32 {
	var size int32
	m.rwMutex.RLock()
	delList := make([]struct {
		Key string
		Val IValue
	}, 0)
	now := time.Now().Unix()
	for k, v := range m.store {
		if v.Expired(now) {
			delList = append(delList, struct {
				Key string
				Val IValue
			}{Key: k, Val: v})
		}
	}
	m.rwMutex.RUnlock()

	for _, del := range delList {
		size += m.delete(del.Key, true)
	}
	return size
}

// runGC interval scan expired key
// when a large number expired keys are set, is suggest for use
func (m *LimitMemoryCache) runGC() {
	go func() {
		time.Sleep(CacheGCInterval)
		m.scanExpiredKeyAndDel()
	}()
}

func (m *LimitMemoryCache) save() error {
	if m.filename == "" {
		return nil
	}

	disk, err := NewDisk(m.filename)
	if err != nil {
		return err
	}
	byt, err := json.Marshal(m.store)
	if err != nil {
		return err
	}
	err = disk.WriteToFile(byt)
	return err
}

func (m *LimitMemoryCache) load() error {
	if m.filename == "" {
		return nil
	}
	values := make(map[string]IValue, 0)
	disk, err := NewDisk(m.filename)
	if err != nil {
		return err
	}
	byt, err := disk.ReadFromFile()
	if err != nil {
		return err
	}

	if len(byt) == 0 {
		return nil
	}
	if err := json.Unmarshal(byt, &values); err != nil {
		return err
	}
	now := time.Now().Unix()
	for k, v := range values {
		if !v.Expired(now) {
			_ = m.set(k, v)
		}
	}
	return nil
}

type Disk struct {
	filename string
}

// NewDisk  filename
func NewDisk(filename string) (*Disk, error) {

	f, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	filename = f

	d := Disk{
		filename: filename,
	}
	return &d, nil
}

// WriteToFile if file exists, just del
func (d *Disk) WriteToFile(data []byte) error {
	if FilenameExists(d.filename) {
		if err := os.Remove(d.filename); err != nil {
			return err
		}
	} else {
		dir := filepath.Dir(d.filename)
		if err := os.MkdirAll(dir, os.FileMode(0666)); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(d.filename, os.O_RDWR|os.O_CREATE, os.FileMode(0666))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// ReadFromFile
func (d *Disk) ReadFromFile() ([]byte, error) {
	data := make([]byte, 0)
	if !FilenameExists(d.filename) {
		return data, nil
	}

	file, err := os.Open(d.filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err = io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CacheGetOrSet 从缓存中获取数据，如果不存在则调用fetchFn获取数据并写入缓存。
// valPtr 必须是指针类型, 并且不能是指针nil, 不能是指针的指针
// valPtr 的值修改，不会影响缓存中的值
// fetchFn 返回的值如果是指针类型，会将指针解引用后赋值给valPtr, 缓存不会保存指针类型的值，外部修改不影响缓存中的值
func CacheGetOrSet(c Cacher, ctx context.Context, key string, valPtr interface{}, fetchFn func(context.Context) (interface{}, error), ttl ...time.Duration) error {
	// valPtr 必须是指针类型, 并且不能是指针nil, 不能是指针的指针
	if reflect.TypeOf(valPtr).Kind() != reflect.Ptr || reflect.ValueOf(valPtr).IsNil() || reflect.TypeOf(valPtr).Elem().Kind() == reflect.Ptr {
		return fmt.Errorf("valPtr must be a non-nil pointer to a non-pointer type, but got %T", valPtr)
	}
	v, err := c.Get(key)
	if err == nil {
		// 检查v的类型是否可以赋值给valPtr指向的基础类型
		vType := reflect.TypeOf(v)
		valPtrType := reflect.TypeOf(valPtr).Elem() // valPtr指向的类型

		if vType.AssignableTo(valPtrType) {
			reflect.ValueOf(valPtr).Elem().Set(reflect.ValueOf(v))
			return nil
		}
	}
	// 如果缓存不存在或者类型不匹配，调用fetchFn
	if fetchFn != nil {
		fetchVal, err := fetchFn(ctx)
		if err != nil {
			return err
		}
		if fetchVal == nil {
			return nil
		}
		if reflect.TypeOf(fetchVal).Kind() == reflect.Ptr {
			fetchVal = reflect.ValueOf(fetchVal).Elem().Interface()
		}
		reflect.ValueOf(valPtr).Elem().Set(reflect.ValueOf(fetchVal))
		// 将fetchFn的结果写入缓存
		if len(ttl) > 0 {
			_ = c.SetWithTTL(key, fetchVal, ttl[0])
		} else {
			_ = c.Set(key, fetchVal)
		}
	}
	return nil
}
