package caches

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

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
	expire := val.ExpiredTime.Sub(time.Now())
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
		return ErrSizeOverCapacity
	}
	return nil
}

func (m *LimitMemoryCache) scanExpiredKeyAndDel() int32 {
	var size int32
	m.rwMutex.RLock()
	dels := make([]struct {
		Key string
		Val IValue
	}, 0)
	now := time.Now().Unix()
	for k, v := range m.store {
		if v.Expired(now) {
			dels = append(dels, struct {
				Key string
				Val IValue
			}{Key: k, Val: v})
		}
	}
	m.rwMutex.RUnlock()

	for _, del := range dels {
		size += m.delete(del.Key, true)
	}
	return size
}

// runGC interval scan expired key
// when a large number expired keys are set, is suggest for use
func (m *LimitMemoryCache) runGC() {
	go func() {
		time.Sleep(GCInterval)
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

// GetCurrentDir -
func GetCurrentDir() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) // absolution filepath.Dir(os.Args[0])
	if err != nil {
		return "", err
	}
	return strings.Replace(dir, "\\", "/", -1), nil // just \ replace to /
}

func FilenameExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}
