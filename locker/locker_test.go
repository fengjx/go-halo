package locker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryLocker_TryLock(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()

	// 测试成功获取锁
	acquired, err := locker.TryLock(ctx, "key1")
	require.NoError(t, err)
	assert.True(t, acquired, "应该成功获取锁")

	// 测试锁已被占用
	acquired2, err := locker.TryLock(ctx, "key1")
	require.NoError(t, err)
	assert.False(t, acquired2, "锁已被占用，应该返回 false")

	// 测试释放锁后可以再次获取
	locker.Unlock("key1")
	acquired3, err := locker.TryLock(ctx, "key1")
	require.NoError(t, err)
	assert.True(t, acquired3, "释放后应该可以再次获取锁")

	// 测试不同 key 之间互不影响
	acquired4, err := locker.TryLock(ctx, "key2")
	require.NoError(t, err)
	assert.True(t, acquired4, "不同 key 应该可以同时获取锁")
}

func TestMemoryLocker_TryLock_ContextCancel(t *testing.T) {
	locker := NewMemoryLocker()
	ctx, cancel := context.WithCancel(context.Background())

	// 先获取锁
	acquired, err := locker.TryLock(ctx, "key1")
	require.NoError(t, err)
	assert.True(t, acquired)

	// 取消 context
	cancel()

	// 尝试获取锁，应该返回 context 错误
	acquired2, err := locker.TryLock(ctx, "key1")
	assert.Error(t, err)
	assert.False(t, acquired2)
	assert.Equal(t, context.Canceled, err)
}

func TestMemoryLocker_Lock(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()

	// 测试成功获取锁
	err := locker.Lock(ctx, "key1", time.Second)
	assert.NoError(t, err, "应该成功获取锁")

	// 测试锁已被占用时等待超时
	done := make(chan bool)
	go func() {
		err := locker.Lock(ctx, "key1", 100*time.Millisecond)
		assert.Error(t, err)
		assert.Equal(t, ErrLockTimeout, err)
		done <- true
	}()

	select {
	case <-done:
		// 测试通过
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Lock 应该在超时后返回错误")
	}

	// 释放锁
	locker.Unlock("key1")
}

func TestMemoryLocker_Lock_ContextCancel(t *testing.T) {
	locker := NewMemoryLocker()
	ctx, cancel := context.WithCancel(context.Background())

	// 先获取锁
	err := locker.Lock(ctx, "key1", time.Second)
	require.NoError(t, err)

	// 启动 goroutine 等待锁
	done := make(chan bool)
	go func() {
		err := locker.Lock(ctx, "key1", time.Second)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		done <- true
	}()

	// 等待一小段时间确保 goroutine 已经开始等待
	time.Sleep(10 * time.Millisecond)

	// 取消 context
	cancel()

	select {
	case <-done:
		// 测试通过
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Lock 应该在 context 取消后立即返回")
	}

	locker.Unlock("key1")
}

func TestMemoryLocker_Lock_SuccessAfterRelease(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()

	// 先获取锁
	err := locker.Lock(ctx, "key1", time.Second)
	require.NoError(t, err)

	// 启动 goroutine 等待锁
	done := make(chan bool)
	go func() {
		err := locker.Lock(ctx, "key1", time.Second)
		assert.NoError(t, err, "锁释放后应该可以成功获取")
		done <- true
	}()

	// 等待一小段时间确保 goroutine 已经开始等待
	time.Sleep(10 * time.Millisecond)

	// 释放锁
	locker.Unlock("key1")

	select {
	case <-done:
		// 测试通过
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Lock 应该在锁释放后立即成功")
	}
}

func TestMemoryLocker_Concurrent(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()
	key := "concurrent_key"

	// 并发计数器
	var counter int
	var mu sync.Mutex

	// 启动多个 goroutine 争抢锁
	const goroutineCount = 10
	const operationsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				err := locker.Lock(ctx, key, time.Second)
				require.NoError(t, err)

				// 临界区：修改共享计数器
				mu.Lock()
				counter++
				mu.Unlock()

				// 模拟一些工作
				time.Sleep(1 * time.Millisecond)

				// 再次检查计数器（应该仍然是递增后的值）
				mu.Lock()
				currentCounter := counter
				mu.Unlock()

				locker.Unlock(key)

				// 验证计数器没有被其他 goroutine 修改
				mu.Lock()
				assert.Equal(t, currentCounter, counter, "锁应该保证互斥性")
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// 验证最终计数器值
	assert.Equal(t, goroutineCount*operationsPerGoroutine, counter)
}

func TestMemoryLocker_DifferentKeys(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()

	// 不同 key 应该可以同时获取锁
	err1 := locker.Lock(ctx, "key1", time.Second)
	require.NoError(t, err1)

	err2 := locker.Lock(ctx, "key2", time.Second)
	require.NoError(t, err2)

	err3 := locker.Lock(ctx, "key3", time.Second)
	require.NoError(t, err3)

	// 释放所有锁
	locker.Unlock("key1")
	locker.Unlock("key2")
	locker.Unlock("key3")
}

func TestMemoryLocker_Release_NotHeld(t *testing.T) {
	locker := NewMemoryLocker()

	// 释放一个未持有的锁不应该 panic
	assert.NotPanics(t, func() {
		locker.Unlock("non_existent_key")
	})
}

func TestMemoryLocker_Release_MultipleTimes(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()

	// 获取锁
	err := locker.Lock(ctx, "key1", time.Second)
	require.NoError(t, err)

	// 释放锁
	locker.Unlock("key1")

	// 多次释放不应该 panic
	assert.NotPanics(t, func() {
		locker.Unlock("key1")
		locker.Unlock("key1")
	})

	// 应该可以再次获取锁
	err = locker.Lock(ctx, "key1", time.Second)
	assert.NoError(t, err)
}
