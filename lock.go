package bkit

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DistributedLocker interface {
	AcquireLock(name string, ttl time.Duration) (bool, error)
	ReleaseLock(name string) error
}

type _distributedLock struct {
	LockName  string
	Owner     string
	Expire    time.Time
	CreatedAt time.Time
}

// 通过Mysql 实现的分布式锁
type MysqlDistributedLocker struct {
	db        *sql.DB
	tableName string
	owner     string

	mutex sync.Mutex
	m     map[string]_distributedLock
}

// NewMysqlLocker 创建一个基于mysql的分布式锁, 适用于不太频繁的场景
func NewMysqlDistributedLocker(dataSourceName, owner string) (*MysqlDistributedLocker, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	lock := &MysqlDistributedLocker{
		db:        db,
		tableName: "_distributed_locks",
		owner:     owner,
		m:         make(map[string]_distributedLock),
	}
	// 初始化表, lock_key 为锁的key, expire 为锁的过期时间
	_, err = lock.db.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			lock_name VARCHAR(255) PRIMARY KEY,
            owner  VARCHAR(255) NOT NULL,
			expire BIGINT NOT NULL,
			created_at BIGINT NOT NULL
		)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用于分布式锁的日志表';`, lock.tableName))
	if err != nil {
		return nil, fmt.Errorf("create table _lock_log: %w", err)
	}
	return lock, nil
}
func (lock *MysqlDistributedLocker) Close() error {
	if lock.db == nil {
		return nil
	}
	return lock.db.Close()
}

// Lock 加锁
func (lock *MysqlDistributedLocker) AcquireLock(name string, ttl time.Duration) (bool, error) {
	if ttl < time.Second {
		ttl = time.Second
	}
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	//  先本地判断是否已经加锁
	if v, ok := lock.m[name]; ok {
		if v.Expire.After(time.Now()) {
			return true, nil
		}
		// 过期了, 删除
		delete(lock.m, name)
	}
	expire := time.Now().Add(ttl)
	// 本地没有, 则去数据库获取锁
	ok, err := lock.tryAcquireLock(name, expire)
	if err != nil {
		return false, err
	}

	if ok {
		lock.m[name] = _distributedLock{
			LockName:  name,
			Owner:     lock.owner,
			Expire:    expire,
			CreatedAt: time.Now(),
		}
	}
	return ok, nil
}
func (lock *MysqlDistributedLocker) tryAcquireLock(name string, expire time.Time) (bool, error) {
	result, err := lock.db.Exec(fmt.Sprintf(`
	INSERT INTO %s (lock_name, owner, expire, created_at)
	VALUES (?, ?, ?, UNIX_TIMESTAMP())
	ON DUPLICATE KEY UPDATE
		owner = IF(expire < UNIX_TIMESTAMP(), VALUES(owner), owner),
		created_at = IF(expire < UNIX_TIMESTAMP(), VALUES(created_at), created_at),
		expire = IF(expire < UNIX_TIMESTAMP(), VALUES(expire), expire)`, lock.tableName),
		name, lock.owner, expire.Unix())
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func (lock *MysqlDistributedLocker) ReleaseLock(key string) error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	if _, ok := lock.m[key]; !ok {
		return fmt.Errorf("%s lock not found", lock.owner)
	}

	// 存在, 则删除
	err := lock.tryRelease(lock.m[key])
	if err != nil {
		return err
	}
	delete(lock.m, key)

	return nil
}

func (lock *MysqlDistributedLocker) tryRelease(v _distributedLock) error {
	_, err := lock.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE lock_name = ? AND owner = ?", lock.tableName), v.LockName, lock.owner)
	return err
}
