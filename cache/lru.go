package cache

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/fengjx/go-halo/cache/internal/lru"
)

// LRUCached 使用 LRU 实现缓存
type lruCache[K comparable, V any] struct {
	lru        *lru.Cache
	mu         sync.RWMutex
	ttl        time.Duration
	cacheEmpty bool
}

type lruItem[K comparable, V any] struct {
	key      K
	val      V
	expireAt int64
}

// isExpire 判断是否过期
// 返回 true 表示已过期
func (i lruItem[K, V]) isExpire() bool {
	return i.expireAt <= time.Now().UnixNano()
}

// NewLRUCache 创建一个 LRU 缓存
func NewLRUCache[K comparable, V any](capacity int, ttl time.Duration, opts ...Option) Cache[K, V] {
	opt := &Options{}
	for _, o := range opts {
		o(opt)
	}
	cache := &lruCache[K, V]{
		lru:        lru.New(capacity),
		ttl:        ttl,
		cacheEmpty: opt.cacheEmpty,
	}
	return cache
}

func (c *lruCache[K, V]) Get(ctx context.Context, key K) *Result[V] {
	c.mu.RLock()
	v, ok := c.getCache(key)
	c.mu.RUnlock()
	if ok {
		return &Result[V]{val: v, err: nil}
	}
	return &Result[V]{}
}

func (c *lruCache[K, V]) GetWithFallback(ctx context.Context, key K, fn Fallback[K, V]) *Result[V] {
	c.mu.RLock()
	v, ok := c.getCache(key)
	c.mu.RUnlock()

	if ok {
		return &Result[V]{val: v, err: nil}
	}

	vf, err := fn(ctx, key)
	if err != nil {
		return &Result[V]{err: err}
	}
	if !c.cacheEmpty && isEmpty(vf) {
		return &Result[V]{val: vf, err: nil}
	}

	c.mu.Lock()
	c.setCache(key, vf)
	c.mu.Unlock()

	return &Result[V]{val: vf, err: nil}
}

func (c *lruCache[K, V]) getCache(key K) (v V, ok bool) {
	val, ok := c.lru.Get(key)
	if ok {
		item := val.(*lruItem[K, V])
		if item.isExpire() {
			// 过期了，需要删除，但这里不能直接删除，因为是在读锁下
			// 返回false表示未找到，让调用方处理
			ok = false
			return
		}
		return item.val, true
	}
	return
}

func (c *lruCache[K, V]) GetMulti(ctx context.Context, keys []K) *Result[map[K]V] {
	c.mu.RLock()
	vals := make(map[K]V)
	for _, k := range keys {
		v, ok := c.getCache(k)
		if ok {
			vals[k] = v
		}
	}
	c.mu.RUnlock()
	return &Result[map[K]V]{val: vals, err: nil}
}

func (c *lruCache[K, V]) GetMultiWithFallback(ctx context.Context, keys []K, fn FallbackMulti[K, V]) *Result[map[K]V] {
	c.mu.RLock()
	vals := make(map[K]V)
	var missKeys []K
	for _, k := range keys {
		v, ok := c.getCache(k)
		if ok {
			vals[k] = v
		} else {
			missKeys = append(missKeys, k)
		}
	}
	c.mu.RUnlock()

	if len(missKeys) == 0 {
		return &Result[map[K]V]{val: vals, err: nil}
	}

	m, err := fn(ctx, missKeys)
	if err != nil {
		return &Result[map[K]V]{err: err}
	}

	c.mu.Lock()
	for _, k := range missKeys {
		v := m[k]
		if !c.cacheEmpty && isEmpty(v) {
			continue
		}
		c.setCache(k, v)
		vals[k] = v
	}
	c.mu.Unlock()
	return &Result[map[K]V]{val: vals, err: nil}
}

func (c *lruCache[K, V]) Set(ctx context.Context, key K, val V) *Result[bool] {
	c.setCache(key, val)
	return &Result[bool]{true, nil}
}

func (c *lruCache[K, V]) setCache(key K, val V) {
	item := &lruItem[K, V]{
		key:      key,
		val:      val,
		expireAt: time.Now().Add(c.ttl).UnixNano(),
	}
	c.lru.Add(key, item)
}

func (c *lruCache[K, V]) SetMulti(ctx context.Context, values map[K]V) *Result[bool] {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range values {
		c.setCache(k, v)
	}
	return &Result[bool]{val: true, err: nil}
}

func (c *lruCache[K, V]) Del(ctx context.Context, keys ...K) *Result[int] {
	c.mu.Lock()
	defer c.mu.Unlock()
	cnt := 0
	for _, k := range keys {
		exist := c.lru.Remove(k)
		if exist {
			cnt++
		}
	}
	return &Result[int]{val: cnt, err: nil}
}

func (c *lruCache[K, V]) Has(ctx context.Context, key K) *Result[bool] {
	c.mu.RLock()
	_, ok := c.getCache(key)
	c.mu.RUnlock()
	return &Result[bool]{val: ok, err: nil}
}

func (c *lruCache[K, V]) Clear(ctx context.Context) *Result[bool] {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Clear()
	return &Result[bool]{val: true, err: nil}
}

// isEmpty 判断一个值是否是空值
func isEmpty(v any) bool {
	vl := reflect.ValueOf(v)
	return vl.Kind() == reflect.Pointer && vl.IsNil()
}
