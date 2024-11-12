package mgo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.woa.com/csm/fault_track/pkg/bkit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 基于mongodb实现的分布式锁
type _distributedLock struct {
	LockName  string    `bson:"lock_name"`
	Owner     string    `bson:"owner"`
	Expire    time.Time `bson:"expire"`
	CreatedAt time.Time `bson:"created_at"`
	Ver       string    `bson:"ver"`
}

// 通过MongoDB 实现的分布式锁
type MongoDistributedLocker struct {
	db         *Database
	collection string
	owner      string

	mutex sync.Mutex
	m     map[string]_distributedLock
}

// NewMongoDistributedLocker 创建一个基于MongoDB的分布式锁, 适用于不太频繁的场景
func NewMongoDistributedLocker(ctx context.Context, conn DBConn, owner string) (*MongoDistributedLocker, error) {
	db, err := NewDatabase(ctx, conn)
	if err != nil {
		return nil, err
	}

	lock := &MongoDistributedLocker{
		db:         db,
		collection: "_distributed_locks",
		owner:      owner,
		m:          make(map[string]_distributedLock),
	}

	// 创建索引
	if err := db.CreateIndexes([]Index{
		{
			Collection: lock.collection,
			Name:       "unique_lock_name",
			Keys: []primitive.E{
				{Key: "lock_name", Value: 1},
			},
			Unique: true,
		},
	}); err != nil {
		return nil, err
	}

	return lock, nil
}

func (lock *MongoDistributedLocker) Close() error {
	if lock.db == nil {
		return nil
	}
	return lock.db.Client().Disconnect(context.Background())
}

// AcquireLock 加锁
func (lock *MongoDistributedLocker) AcquireLock(name string, ttl time.Duration) error {
	if ttl < time.Second {
		ttl = time.Second
	}
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	// 先本地判断是否已经加锁
	if v, ok := lock.m[name]; ok {
		if v.Expire.After(time.Now()) {
			return nil
		}
		// 过期了, 删除
		delete(lock.m, name)
	}
	expire := time.Now().Local().Add(ttl)
	// 本地没有, 则去数据库获取锁
	var (
		ok  bool
		err error
	)

	bkit.Retry.RetryN(1, 100*time.Millisecond, func() error {
		ok, err = lock.tryAcquireLock(name, expire)
		if err != nil {
			return err
		}
		return nil
	})

	if ok {
		lock.m[name] = _distributedLock{
			LockName:  name,
			Owner:     lock.owner,
			Expire:    expire,
			CreatedAt: time.Now().Local(),
		}
		return nil
	}
	return bkit.ErrAcquireLockFailed
}

func (lock *MongoDistributedLocker) tryAcquireLock(name string, expire time.Time) (bool, error) {
	isLock := false

	err := lock.db.Transaction(context.Background(), func(session SessionContext) error {
		doc := &_distributedLock{}
		exists, err := lock.db.FindOne(session, lock.collection, bson.M{"lock_name": name}, doc)
		if err != nil {
			return err
		}
		v := _distributedLock{
			LockName:  name,
			Owner:     lock.owner,
			Expire:    expire,
			CreatedAt: time.Now().Local(),
			Ver:       fmt.Sprintf("%d", time.Now().UnixNano()),
		}
		if !exists {
			_, err := lock.db.Collection(lock.collection).InsertOne(session, v)
			if err != nil {
				if mongo.IsDuplicateKeyError(err) {
					// 其他协程已经插入了
					return nil
				}
				return err
			}
			// 获取锁成功
			isLock = true
			return nil
		}

		// 已经存在
		if doc.Expire.Unix() > time.Now().Unix() {
			// 未过期
			return nil
		}
		filter := bson.M{
			"lock_name": name,
			"ver":       doc.Ver,
		}
		up := bson.M{
			"$set": bson.M{
				"owner":      lock.owner,
				"expire":     expire,
				"created_at": time.Now().Local(),
				"ver":        v.Ver,
			},
		}
		// 已经过期, 去更新
		ret, err := lock.db.Collection(lock.collection).UpdateOne(session, filter, up)
		if err != nil {
			return err
		}
		if ret.MatchedCount == 0 {
			// 数据已经被其他协程更新
			return nil
		}
		if ret.ModifiedCount > 0 {
			// 更新成功
			isLock = true
		}
		return nil
	})

	return isLock, err
}

// ReleaseLock 释放锁
func (lock *MongoDistributedLocker) ReleaseLock(name string) error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	if _, ok := lock.m[name]; !ok {
		return fmt.Errorf("%s lock not found", lock.owner)
	}
	// 存在, 则删除
	var err error
	bkit.Retry.RetryN(1, 100*time.Millisecond, func() error {
		err = lock.tryRelease(lock.m[name])
		return err
	})
	if err != nil {
		return err
	}
	delete(lock.m, name)

	return nil
}

func (lock *MongoDistributedLocker) tryRelease(v _distributedLock) error {
	filter := bson.M{
		"lock_name": v.LockName,
		"owner":     lock.owner,
	}
	_, err := lock.db.Collection(lock.collection).DeleteOne(context.TODO(), filter)
	return err
}

// ReleaseAllLocks 释放所有锁
func (lock *MongoDistributedLocker) ReleaseAllLocks() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	for _, v := range lock.m {
		err := lock.tryRelease(v)
		if err != nil {
			return err
		}
	}
	lock.m = make(map[string]_distributedLock)
	return nil
}
