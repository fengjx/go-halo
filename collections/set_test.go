package collections

import (
	"testing"
)

func TestNewSet(t *testing.T) {
	s := NewSet[string]()
	if !s.IsEmpty() {
		t.Error("新创建的集合应该为空")
	}
	if s.Size() != 0 {
		t.Error("新创建的集合大小应该为0")
	}
}

func TestNewSetWithValues(t *testing.T) {
	s := NewSetWithValues("a", "b", "c")
	if s.Size() != 3 {
		t.Errorf("期望大小为3，实际为%d", s.Size())
	}
	if !s.Has("a") || !s.Has("b") || !s.Has("c") {
		t.Error("集合应该包含所有初始值")
	}
}

func TestAddAndRemove(t *testing.T) {
	s := NewSet[int]()
	
	// 测试添加
	s.Add(1)
	s.Add(2)
	s.Add(3)
	
	if s.Size() != 3 {
		t.Errorf("期望大小为3，实际为%d", s.Size())
	}
	
	if !s.Has(1) || !s.Has(2) || !s.Has(3) {
		t.Error("集合应该包含所有添加的元素")
	}
	
	// 测试重复添加
	s.Add(1)
	if s.Size() != 3 {
		t.Error("重复添加不应该改变集合大小")
	}
	
	// 测试移除
	s.Rem(2)
	if s.Size() != 2 {
		t.Errorf("移除后期望大小为2，实际为%d", s.Size())
	}
	if s.Has(2) {
		t.Error("移除的元素不应该存在于集合中")
	}
}

func TestAddAllAndRemoveAll(t *testing.T) {
	s := NewSet[string]()
	
	// 测试批量添加
	s.AddAll("a", "b", "c", "d")
	if s.Size() != 4 {
		t.Errorf("期望大小为4，实际为%d", s.Size())
	}
	
	// 测试批量移除
	s.RemAll("b", "d")
	if s.Size() != 2 {
		t.Errorf("期望大小为2，实际为%d", s.Size())
	}
	if s.Has("b") || s.Has("d") {
		t.Error("移除的元素不应该存在于集合中")
	}
}

func TestContainsAll(t *testing.T) {
	s := NewSetWithValues(1, 2, 3, 4, 5)
	
	if !s.HasAll(1, 2, 3) {
		t.Error("集合应该包含所有指定元素")
	}
	
	if s.HasAll(1, 2, 6) {
		t.Error("集合不应该包含不存在的元素")
	}
}

func TestClear(t *testing.T) {
	s := NewSetWithValues(1, 2, 3)
	s.Clear()
	
	if !s.IsEmpty() {
		t.Error("清空后集合应该为空")
	}
	if s.Size() != 0 {
		t.Error("清空后集合大小应该为0")
	}
}

func TestToSlice(t *testing.T) {
	s := NewSetWithValues(3, 1, 2)
	slice := s.ToSlice()
	
	if len(slice) != 3 {
		t.Errorf("期望切片长度为3，实际为%d", len(slice))
	}
	
	// 检查切片包含所有元素
	expected := map[int]bool{1: true, 2: true, 3: true}
	for _, v := range slice {
		if !expected[v] {
			t.Errorf("切片包含意外元素: %d", v)
		}
	}
}

func TestToSortedSlice(t *testing.T) {
	s := NewSetWithValues(3, 1, 4, 2)
	slice := s.ToSortedSlice(func(a, b int) bool {
		return a < b
	})
	
	expected := []int{1, 2, 3, 4}
	for i, v := range slice {
		if v != expected[i] {
			t.Errorf("期望位置%d的值为%d，实际为%d", i, expected[i], v)
		}
	}
}

func TestUnion(t *testing.T) {
	s1 := NewSetWithValues(1, 2, 3)
	s2 := NewSetWithValues(3, 4, 5)
	union := s1.Union(s2)
	
	expected := NewSetWithValues(1, 2, 3, 4, 5)
	if !union.Equals(expected) {
		t.Errorf("并集结果不正确，期望%v，实际%v", expected.ToSlice(), union.ToSlice())
	}
}

func TestIntersection(t *testing.T) {
	s1 := NewSetWithValues(1, 2, 3, 4)
	s2 := NewSetWithValues(3, 4, 5, 6)
	intersection := s1.Intersection(s2)
	
	expected := NewSetWithValues(3, 4)
	if !intersection.Equals(expected) {
		t.Errorf("交集结果不正确，期望%v，实际%v", expected.ToSlice(), intersection.ToSlice())
	}
}

func TestDifference(t *testing.T) {
	s1 := NewSetWithValues(1, 2, 3, 4)
	s2 := NewSetWithValues(3, 4, 5, 6)
	difference := s1.Diff(s2)
	
	expected := NewSetWithValues(1, 2)
	if !difference.Equals(expected) {
		t.Errorf("差集结果不正确，期望%v，实际%v", expected.ToSlice(), difference.ToSlice())
	}
}

func TestSymmetricDifference(t *testing.T) {
	s1 := NewSetWithValues(1, 2, 3)
	s2 := NewSetWithValues(3, 4, 5)
	symmetricDiff := s1.SymmetricDiff(s2)
	
	expected := NewSetWithValues(1, 2, 4, 5)
	if !symmetricDiff.Equals(expected) {
		t.Errorf("对称差集结果不正确，期望%v，实际%v", expected.ToSlice(), symmetricDiff.ToSlice())
	}
}

func TestIsSubset(t *testing.T) {
	s1 := NewSetWithValues(1, 2)
	s2 := NewSetWithValues(1, 2, 3, 4)
	s3 := NewSetWithValues(1, 5)
	
	if !s1.IsSubset(s2) {
		t.Error("s1应该是s2的子集")
	}
	
	if s1.IsSubset(s3) {
		t.Error("s1不应该是s3的子集")
	}
	
	if !s1.IsSubset(s1) {
		t.Error("集合应该是自己的子集")
	}
}

func TestIsSuperset(t *testing.T) {
	s1 := NewSetWithValues(1, 2, 3, 4)
	s2 := NewSetWithValues(1, 2)
	
	if !s1.IsSuperset(s2) {
		t.Error("s1应该是s2的超集")
	}
	
	if s2.IsSuperset(s1) {
		t.Error("s2不应该是s1的超集")
	}
}

func TestIsDisjoint(t *testing.T) {
	s1 := NewSetWithValues(1, 2)
	s2 := NewSetWithValues(3, 4)
	s3 := NewSetWithValues(2, 3)
	
	if !s1.IsDisjoint(s2) {
		t.Error("s1和s2应该不相交")
	}
	
	if s1.IsDisjoint(s3) {
		t.Error("s1和s3应该相交")
	}
}

func TestEquals(t *testing.T) {
	s1 := NewSetWithValues(1, 2, 3)
	s2 := NewSetWithValues(3, 1, 2)
	s3 := NewSetWithValues(1, 2)
	
	if !s1.Equals(s2) {
		t.Error("s1和s2应该相等")
	}
	
	if s1.Equals(s3) {
		t.Error("s1和s3不应该相等")
	}
}

func TestClone(t *testing.T) {
	s1 := NewSetWithValues(1, 2, 3)
	s2 := s1.Clone()
	
	if !s1.Equals(s2) {
		t.Error("克隆的集合应该与原集合相等")
	}
	
	// 修改原集合，确保克隆不受影响
	s1.Add(4)
	if s2.Has(4) {
		t.Error("修改原集合不应该影响克隆")
	}
}

func TestFilter(t *testing.T) {
	s := NewSetWithValues(1, 2, 3, 4, 5, 6)
	filtered := s.Filter(func(x int) bool {
		return x%2 == 0
	})
	
	expected := NewSetWithValues(2, 4, 6)
	if !filtered.Equals(expected) {
		t.Errorf("过滤结果不正确，期望%v，实际%v", expected.ToSlice(), filtered.ToSlice())
	}
}

func TestForEach(t *testing.T) {
	s := NewSetWithValues(1, 2, 3)
	sum := 0
	s.ForEach(func(x int) {
		sum += x
	})
	
	if sum != 6 {
		t.Errorf("期望和为6，实际为%d", sum)
	}
}

func TestString(t *testing.T) {
	s := NewSetWithValues(1, 2, 3)
	str := s.String()
	
	if str != "Set{[1 2 3]}" && str != "Set{[1 3 2]}" && str != "Set{[2 1 3]}" && 
	   str != "Set{[2 3 1]}" && str != "Set{[3 1 2]}" && str != "Set{[3 2 1]}" {
		t.Errorf("字符串表示不正确: %s", str)
	}
	
	empty := NewSet[int]()
	if empty.String() != "Set{}" {
		t.Errorf("空集合字符串表示不正确: %s", empty.String())
	}
}

// 基准测试
func BenchmarkAdd(b *testing.B) {
	s := NewSet[int]()
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
}

func BenchmarkContains(b *testing.B) {
	s := NewSet[int]()
	for i := 0; i < 1000; i++ {
		s.Add(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Has(i % 1000)
	}
}

func BenchmarkUnion(b *testing.B) {
	s1 := NewSet[int]()
	s2 := NewSet[int]()
	for i := 0; i < 1000; i++ {
		s1.Add(i)
		s2.Add(i + 500)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s1.Union(s2)
	}
}
