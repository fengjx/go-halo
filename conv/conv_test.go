package conv

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

// 测试结构体定义

type UserEntity struct {
	ID   int         `json:"id" db:"id"`
	Name string      `json:"name" db:"name"`
	Addr *AddrEntity `json:"addr" db:"addr"`
}

type AddrEntity struct {
	City string `json:"city" db:"city"`
}

type UserDTO struct {
	ID   int      `json:"id"`
	Name string   `json:"name"`
	Addr *AddrDTO `json:"addr"`
}

type AddrDTO struct {
	City string `json:"city"`
}

// 用于测试 time.Time <-> int64 互转
type TimeEntity struct {
	T1 int64
	T2 int64
}

type TimeDTO struct {
	T1 time.Time
	T2 time.Time
}

type TimeEntity2 struct {
	T1 time.Time
	T2 time.Time
}

type TimeDTO2 struct {
	T1 int64
	T2 int64
}

func TestConvert_Basic(t *testing.T) {
	entity := &UserEntity{
		ID:   1,
		Name: "Tom",
		Addr: &AddrEntity{City: "Beijing"},
	}
	result := Convert[*UserEntity, *UserDTO](entity)
	dto := result.Val()
	if result.Error() != nil {
		t.Fatalf("转换失败: %v", result.Error())
	}
	printJSON(t, "dto", dto)
	if dto.ID != entity.ID || dto.Name != entity.Name || dto.Addr == nil || dto.Addr.City != entity.Addr.City {
		t.Errorf("字段未正确拷贝: %+v", dto)
	}
}

type AddrDTODB struct {
	City string `db:"city"`
}
type UserDTODB struct {
	ID   int        `db:"id"`
	Name string     `db:"name"`
	Addr *AddrDTODB `db:"addr"`
}

func TestConvert_WithTagName(t *testing.T) {
	entity := UserEntity{
		ID:   2,
		Name: "Jerry",
		Addr: &AddrEntity{City: "Shanghai"},
	}
	result := Convert[UserEntity, UserDTODB](entity, WithTag("db"))
	dto := result.Val()
	if result.Error() != nil {
		t.Fatalf("tagName转换失败: %v", result.Error())
	}
	if dto.ID != entity.ID || dto.Name != entity.Name || dto.Addr == nil || dto.Addr.City != entity.Addr.City {
		t.Errorf("tagName字段未正确拷贝: %+v", dto)
	}
}

func TestConvert_NilPointer(t *testing.T) {
	var entity *UserEntity = nil
	result := Convert[*UserEntity, UserDTO](entity)
	dto := result.Val()
	if result.Error() != nil {
		t.Fatalf("nil指针转换失败: %v", result.Error())
	}
	if !reflect.ValueOf(dto).IsZero() {
		t.Errorf("nil输入应返回零值: %+v", dto)
	}
}

func TestConvert_CustomConverter(t *testing.T) {
	Register[UserEntity, UserDTO](func(src UserEntity) (UserDTO, error) {
		return UserDTO{ID: 100, Name: "custom"}, nil
	})
	entity := UserEntity{ID: 1, Name: "Tom"}
	result := Convert[UserEntity, UserDTO](entity)
	dto := result.Val()
	if result.Error() != nil {
		t.Fatalf("自定义转换器失败: %v", result.Error())
	}
	if dto.ID != 100 || dto.Name != "custom" {
		t.Errorf("自定义转换器未生效: %+v", dto)
	}
}

// 类型别名测试
type MyInt int
type MyString = string

type AliasEntity struct {
	A MyInt
	B MyString
}

type AliasDTO struct {
	A int
	B string
}

func TestConvert_AliasType(t *testing.T) {
	entity := AliasEntity{
		A: 123,
		B: "hello",
	}
	result := Convert[AliasEntity, AliasDTO](entity)
	dto := result.Val()
	if result.Error() != nil {
		t.Fatalf("alias type 转换失败: %v", result.Error())
	}
	if dto.A != int(entity.A) || dto.B != string(entity.B) {
		t.Errorf("alias type 字段未正确拷贝: %+v", dto)
	}
}

func TestConvert_TimeInt64_S_Default(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	entity := TimeEntity{
		T1: now.Unix(),
		T2: now.Add(time.Hour).Unix(),
	}
	dto := Convert[TimeEntity, TimeDTO](entity).Val()
	if !dto.T1.Equal(now) || !dto.T2.Equal(now.Add(time.Hour)) {
		t.Errorf("time.Time <- int64(s, default) 转换失败: %+v", dto)
	}
	// 反向
	entity2 := TimeEntity2{T1: now, T2: now.Add(time.Hour)}
	dto2 := Convert[TimeEntity2, TimeDTO2](entity2).Val()
	if dto2.T1 != now.Unix() || dto2.T2 != now.Add(time.Hour).Unix() {
		t.Errorf("int64(s, default) <- time.Time 转换失败: %+v", dto2)
	}
}

func TestConvert_TimeInt64_S(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	entity := TimeEntity{
		T1: now.Unix(),
		T2: now.Add(time.Hour).Unix(),
	}
	dto := Convert[TimeEntity, TimeDTO](entity, WithTimeUnit(time.Second)).Val()
	if !dto.T1.Equal(now) || !dto.T2.Equal(now.Add(time.Hour)) {
		t.Errorf("time.Time <- int64(s) 转换失败: %+v", dto)
	}
	// 反向
	entity2 := TimeEntity2{T1: now, T2: now.Add(time.Hour)}
	dto2 := Convert[TimeEntity2, TimeDTO2](entity2, WithTimeUnit(time.Second)).Val()
	if dto2.T1 != now.Unix() || dto2.T2 != now.Add(time.Hour).Unix() {
		t.Errorf("int64(s) <- time.Time 转换失败: %+v", dto2)
	}
}

func TestConvert_TimeInt64_NS(t *testing.T) {
	now := time.Now().Truncate(time.Nanosecond)
	entity := TimeEntity{
		T1: now.UnixNano(),
		T2: now.Add(time.Hour).UnixNano(),
	}
	dto := Convert[TimeEntity, TimeDTO](entity, WithTimeUnit(time.Nanosecond)).Val()
	if !dto.T1.Equal(now) || !dto.T2.Equal(now.Add(time.Hour)) {
		t.Errorf("time.Time <- int64(ns) 转换失败: %+v", dto)
	}
	// 反向
	entity2 := TimeEntity2{T1: now, T2: now.Add(time.Hour)}
	dto2 := Convert[TimeEntity2, TimeDTO2](entity2, WithTimeUnit(time.Nanosecond)).Val()
	if dto2.T1 != now.UnixNano() || dto2.T2 != now.Add(time.Hour).UnixNano() {
		t.Errorf("int64(ns) <- time.Time 转换失败: %+v", dto2)
	}
}

func TestConvert_TimeInt64_US(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	entity := TimeEntity{
		T1: now.UnixMicro(),
		T2: now.Add(time.Hour).UnixMicro(),
	}
	dto := Convert[TimeEntity, TimeDTO](entity, WithTimeUnit(time.Microsecond)).Val()
	if !dto.T1.Equal(now) || !dto.T2.Equal(now.Add(time.Hour)) {
		t.Errorf("time.Time <- int64(us) 转换失败: %+v", dto)
	}
	// 反向
	entity2 := TimeEntity2{T1: now, T2: now.Add(time.Hour)}
	dto2 := Convert[TimeEntity2, TimeDTO2](entity2, WithTimeUnit(time.Microsecond)).Val()
	if dto2.T1 != now.UnixMicro() || dto2.T2 != now.Add(time.Hour).UnixMicro() {
		t.Errorf("int64(us) <- time.Time 转换失败: %+v", dto2)
	}
}

func BenchmarkConvert_Struct(b *testing.B) {
	entity := UserEntity{
		ID:   1,
		Name: "Tom",
		Addr: &AddrEntity{City: "Beijing"},
	}
	for i := 0; i < b.N; i++ {
		result := Convert[UserEntity, UserDTO](entity)
		if result.Error() != nil {
			b.Fatal(result.Error())
		}
	}
}

func BenchmarkConvert_Ptr(b *testing.B) {
	entity := &UserEntity{
		ID:   1,
		Name: "Tom",
		Addr: &AddrEntity{City: "Beijing"},
	}
	for i := 0; i < b.N; i++ {
		result := Convert[*UserEntity, UserDTO](entity)
		if result.Error() != nil {
			b.Fatal(result.Error())
		}
	}
}

// 打印结果辅助
func ExampleConvert() {
	entity := UserEntity{ID: 1, Name: "Tom", Addr: &AddrEntity{City: "Beijing"}}
	dto := Convert[UserEntity, UserDTO](entity).Val()
	fmt.Println(dto.ID, dto.Name, dto.Addr.City)
	// Output: 1 Tom Beijing
}

func printJSON(t *testing.T, msg string, v any) {
	json, _ := json.Marshal(v)
	t.Log(msg, string(json))
}
