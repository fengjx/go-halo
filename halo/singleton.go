package halo

import (
	"sync"
)

// Singleton 是一个泛型的结构体，支持任意类型的单例模式
type Singleton[T any] struct {
	instance *T
	once     sync.Once
	factory  func() *T
}

// Get 提供一个获取单例对象的通用方法
func (s *Singleton[T]) Get() *T {
	// 确保初始化只执行一次
	s.once.Do(func() {
		s.instance = s.factory() // 使用工厂函数创建实例
	})
	return s.instance
}

func NewSingleton[T any](factory func() *T) *Singleton[T] {
	return &Singleton[T]{
		factory: factory,
	}
}
