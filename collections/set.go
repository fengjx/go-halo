package collections

import (
	"fmt"
	"sort"
)

// Set 是一个基于 map 实现的集合数据结构
type Set[T comparable] struct {
	data map[T]bool
}

// NewSet 创建一个新的空集合
func NewSet[T comparable]() *Set[T] {
	return NewSetWithCap[T](4)
}

// NewSetWithCap 创建一个新的空集合，并指定初始容量
func NewSetWithCap[T comparable](cap int) *Set[T] {
	return &Set[T]{
		data: make(map[T]bool, cap),
	}
}

// NewSetWithValues 创建一个包含指定值的集合
func NewSetWithValues[T comparable](values ...T) *Set[T] {
	s := NewSet[T]()
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// Add 向集合中添加元素
func (s *Set[T]) Add(value T) {
	s.data[value] = true
}

// AddAll 向集合中添加多个元素
func (s *Set[T]) AddAll(values ...T) {
	for _, v := range values {
		s.Add(v)
	}
}

// Rem 从集合中移除元素
func (s *Set[T]) Rem(value T) {
	delete(s.data, value)
}

// RemAll 从集合中移除多个元素
func (s *Set[T]) RemAll(values ...T) {
	for _, v := range values {
		s.Rem(v)
	}
}

// Has 检查集合是否包含指定元素
func (s *Set[T]) Has(value T) bool {
	_, exists := s.data[value]
	return exists
}

// HasAll 检查集合是否包含所有指定元素
func (s *Set[T]) HasAll(values ...T) bool {
	for _, v := range values {
		if !s.Has(v) {
			return false
		}
	}
	return true
}

// Size 返回集合中元素的数量
func (s *Set[T]) Size() int {
	return len(s.data)
}

// IsEmpty 检查集合是否为空
func (s *Set[T]) IsEmpty() bool {
	return s.Size() == 0
}

// Clear 清空集合
func (s *Set[T]) Clear() {
	s.data = make(map[T]bool)
}

// ToSlice 将集合转换为切片
func (s *Set[T]) ToSlice() []T {
	result := make([]T, 0, s.Size())
	for k := range s.data {
		result = append(result, k)
	}
	return result
}

// ToSortedSlice 将集合转换为排序后的切片
func (s *Set[T]) ToSortedSlice(less func(a, b T) bool) []T {
	slice := s.ToSlice()
	sort.Slice(slice, func(i, j int) bool {
		return less(slice[i], slice[j])
	})
	return slice
}

// Union 返回当前集合与另一个集合的并集
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	
	// 添加当前集合的所有元素
	for k := range s.data {
		result.Add(k)
	}
	
	// 添加另一个集合的所有元素
	for k := range other.data {
		result.Add(k)
	}
	
	return result
}

// Intersection 返回当前集合与另一个集合的交集
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	
	// 遍历较小的集合以提高效率
	if s.Size() <= other.Size() {
		for k := range s.data {
			if other.Has(k) {
				result.Add(k)
			}
		}
	} else {
		for k := range other.data {
			if s.Has(k) {
				result.Add(k)
			}
		}
	}
	
	return result
}

// Diff 返回当前集合与另一个集合的差集（当前集合中有但另一个集合中没有的元素）
func (s *Set[T]) Diff(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	
	for k := range s.data {
		if !other.Has(k) {
			result.Add(k)
		}
	}
	
	return result
}

// SymmetricDifference 返回当前集合与另一个集合的对称差集（只在一个集合中存在的元素）
func (s *Set[T]) SymmetricDiff(other *Set[T]) *Set[T] {
	union := s.Union(other)
	intersection := s.Intersection(other)
	return union.Diff(intersection)
}

// IsSubset 检查当前集合是否是另一个集合的子集
func (s *Set[T]) IsSubset(other *Set[T]) bool {
	if s.Size() > other.Size() {
		return false
	}
	
	for k := range s.data {
		if !other.Has(k) {
			return false
		}
	}
	
	return true
}

// IsSuperset 检查当前集合是否是另一个集合的超集
func (s *Set[T]) IsSuperset(other *Set[T]) bool {
	return other.IsSubset(s)
}

// IsDisjoint 检查当前集合与另一个集合是否不相交（没有共同元素）
func (s *Set[T]) IsDisjoint(other *Set[T]) bool {
	for k := range s.data {
		if other.Has(k) {
			return false
		}
	}
	return true
}

// Equals 检查当前集合是否与另一个集合相等
func (s *Set[T]) Equals(other *Set[T]) bool {
	if s.Size() != other.Size() {
		return false
	}
	
	for k := range s.data {
		if !other.Has(k) {
			return false
		}
	}
	
	return true
}

// Clone 创建当前集合的副本
func (s *Set[T]) Clone() *Set[T] {
	result := NewSet[T]()
	for k := range s.data {
		result.Add(k)
	}
	return result
}

// Filter 根据条件过滤集合中的元素
func (s *Set[T]) Filter(predicate func(T) bool) *Set[T] {
	result := NewSet[T]()
	for k := range s.data {
		if predicate(k) {
			result.Add(k)
		}
	}
	return result
}

// ForEach 对集合中的每个元素执行指定操作
func (s *Set[T]) ForEach(action func(T)) {
	for k := range s.data {
		action(k)
	}
}

// String 返回集合的字符串表示
func (s *Set[T]) String() string {
	if s.IsEmpty() {
		return "Set{}"
	}
	
	slice := s.ToSlice()
	return fmt.Sprintf("Set{%v}", slice)
}

