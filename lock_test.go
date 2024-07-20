package bkit

import (
	"fmt"
	"testing"
	"time"
)

func TestMysqlDistributedLocker(t *testing.T) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		"root", "123456", "127.0.0.1", "33060", "test_db")
	lock, err := NewMysqlDistributedLocker(dsn, "owner_1")
	if err != nil {
		t.Fatal(err)
	}

	key := "test_key"
	ok, err := lock.AcquireLock(key, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("AcquireLock failed")
	}
	lock2, err := NewMysqlDistributedLocker(dsn, "owner_2")
	if err != nil {
		t.Fatal(err)
	}
	ok, err = lock2.AcquireLock(key, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("AcquireLock 不应该获取到锁")
	}

	time.Sleep(25 * time.Second)
	ok, err = lock2.AcquireLock(key, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("AcquireLock 应该获取到锁")
	}
}
