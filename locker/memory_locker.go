package locker

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrLockTimeout 锁获取超时错误
	ErrLockTimeout = errors.New("locker: lock timeout")
)

// memoryLocker 基于内存的锁实现
type memoryLocker struct {
	mu    sync.Mutex
	locks map[string]*lockEntry
}

// lockEntry 每个 key 对应的锁条目
// 使用容量为 1 的 channel 作为信号量，channel 中有值表示锁可用
type lockEntry struct {
	ch chan struct{}
}

// getOrCreateEntry 获取或创建锁条目
func (m *memoryLocker) getOrCreateEntry(key string) *lockEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.locks[key]
	if !ok {
		entry = &lockEntry{
			ch: make(chan struct{}, 1),
		}
		// 初始化时发送一个值，表示锁当前空闲
		entry.ch <- struct{}{}
		m.locks[key] = entry
	}
	return entry
}

// TryLock 尝试获取锁，获取不到不会阻塞
func (m *memoryLocker) TryLock(ctx context.Context, key string) (bool, error) {
	entry := m.getOrCreateEntry(key)

	select {
	case <-entry.ch:
		// 成功从 channel 中接收到值，表示获取锁成功
		return true, nil
	case <-ctx.Done():
		// context 被取消
		return false, ctx.Err()
	default:
		// channel 中没有值，锁已被占用
		return false, nil
	}
}

// Lock 尝试获取锁，获取不到则等待指定时长
func (m *memoryLocker) Lock(ctx context.Context, key string, timeout time.Duration) error {
	entry := m.getOrCreateEntry(key)

	// 创建超时定时器
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	select {
	case <-entry.ch:
		// 成功获取锁
		return nil
	case <-timeoutTimer.C:
		// 超时
		return ErrLockTimeout
	case <-ctx.Done():
		// context 被取消或 deadline 超时
		return ctx.Err()
	}
}

// Unlock 释放指定 key 的锁
func (m *memoryLocker) Unlock(key string) {
	m.mu.Lock()
	entry, ok := m.locks[key]
	m.mu.Unlock()

	if !ok {
		// key 不存在，无需释放
		return
	}

	// 尝试发送值到 channel，表示归还锁
	// 使用 select + default 避免阻塞（如果 channel 已满，说明锁已经被释放）
	select {
	case entry.ch <- struct{}{}:
		// 成功归还锁
	default:
		// channel 已满，说明锁已经被释放，忽略重复释放
	}
}
