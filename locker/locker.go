package locker

import (
	"context"
	"time"
)

// Locker 锁接口，提供基于 key 的加锁和解锁功能
// 支持多种实现方式，如内存锁、Redis 锁等
type Locker interface {
	// TryLock 尝试获取锁，获取不到不会阻塞，立即返回
	// 返回 true 表示获取锁成功，false 表示锁已被占用
	TryLock(ctx context.Context, key string) (bool, error)

	// Lock 尝试获取锁，获取不到则等待指定时长
	// ctx 用于取消操作，timeout 指定等待超时时间
	// 如果超时或 ctx 被取消，返回错误
	Lock(ctx context.Context, key string, timeout time.Duration) error

	// Unlock 释放指定 key 的锁
	Unlock(key string)
}

// NewMemoryLocker 创建一个基于内存的锁实现
func NewMemoryLocker() Locker {
	return &memoryLocker{
		locks: make(map[string]*lockEntry),
	}
}
