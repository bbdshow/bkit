package mgo

import (
	"context"
	"testing"
	"time"

	"git.woa.com/csm/fault_track/pkg/bkit"
)

func TestMongoDistributedLocker(t *testing.T) {
	conn := DBConn{
		URI:      "mongodb://root:111111@127.0.0.1:27017/admin?replicaSet=rs0&authSource=admin",
		Database: "test_db",
	}
	lock, err := NewMongoDistributedLocker(context.TODO(), conn, "owner_1")
	if err != nil {
		t.Fatal(err)
	}

	key := "test_key"
	if err := lock.AcquireLock(key, 5*time.Second); err != nil {
		if err != bkit.ErrAcquireLockFailed {
			t.Fatal(err)
		}
		if err == bkit.ErrAcquireLockFailed {
			t.Fatal("AcquireLock failed")
		}
	}

	lock2, err := NewMongoDistributedLocker(context.TODO(), conn, "owner_2")
	if err != nil {
		t.Fatal(err)
	}
	if err = lock2.AcquireLock(key, 5*time.Second); err != nil {
		if err != bkit.ErrAcquireLockFailed {
			t.Fatal(err)
		}

	} else {
		t.Fatal("AcquireLock 不应该获取到锁")
	}

	time.Sleep(10 * time.Second)
	if err = lock2.AcquireLock(key, 5*time.Second); err != nil {
		if err != bkit.ErrAcquireLockFailed {
			t.Fatal(err)
		}
		if err == bkit.ErrAcquireLockFailed {
			t.Fatal("AcquireLock 应该获取到锁")
		}
	}
}
