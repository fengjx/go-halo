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
	ID        int               `json:"id" db:"id"`
	Name      string            `json:"name" db:"name"`
	Addr      *AddrEntity       `json:"addr" db:"addr"`
	Age       int               `json:"age" db:"age"`
	Score     float64           `json:"score" db:"score"`
	Active    bool              `json:"active" db:"active"`
	Tags      []string          `json:"tags" db:"tags"`
	Attrs     map[string]string `json:"attrs" db:"attrs"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time        `json:"updated_at" db:"updated_at"`
	Any       interface{}       `json:"any" db:"any"`
	Alias     MyInt             `json:"alias" db:"alias"`
	Numbers   []int             `json:"numbers" db:"numbers"`
	Data      []byte            `json:"data" db:"data"`
	Nested    NestedEntity      `json:"nested" db:"nested"`
}

type AddrEntity struct {
	City     string     `json:"city" db:"city"`
	PostCode int        `json:"post_code" db:"post_code"`
	Location [2]float64 `json:"location" db:"location"`
}

type UserDTO struct {
	ID        int               `json:"id"`
	Name      string            `json:"name"`
	Addr      *AddrDTO          `json:"addr"`
	Age       int               `json:"age"`
	Score     float64           `json:"score"`
	Active    bool              `json:"active"`
	Tags      []string          `json:"tags"`
	Attrs     map[string]string `json:"attrs"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt *time.Time        `json:"updated_at"`
	Any       interface{}       `json:"any"`
	Alias     int               `json:"alias"`
	Numbers   []int             `json:"numbers"`
	Data      []byte            `json:"data"`
	Nested    NestedDTO         `json:"nested"`
}

type AddrDTO struct {
	City     string     `json:"city"`
	PostCode int        `json:"post_code"`
	Location [2]float64 `json:"location"`
}

type NestedEntity struct {
	Desc string
	Val  int
}

type NestedDTO struct {
	Desc string
	Val  int
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
	now := time.Now()
	nowPtr := &now
	entity := &UserEntity{
		ID:        1,
		Name:      "Tom",
		Addr:      &AddrEntity{City: "Beijing", PostCode: 100000, Location: [2]float64{39.9042, 116.4074}},
		Age:       30,
		Score:     99.5,
		Active:    true,
		Tags:      []string{"golang", "dev"},
		Attrs:     map[string]string{"role": "admin", "team": "backend"},
		CreatedAt: now,
		UpdatedAt: nowPtr,
		Any:       map[string]any{"k": 1},
		Alias:     42,
		Numbers:   []int{1, 2, 3},
		Data:      []byte("hello world"),
		Nested:    NestedEntity{Desc: "嵌套结构体", Val: 888},
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
	if dto.Age != entity.Age {
		t.Errorf("Age 字段未正确拷贝")
	}
	if dto.Score != entity.Score {
		t.Errorf("Score 字段未正确拷贝")
	}
	if dto.Active != entity.Active {
		t.Errorf("Active 字段未正确拷贝")
	}
	if len(dto.Tags) != len(entity.Tags) || dto.Tags[0] != entity.Tags[0] {
		t.Errorf("Tags 字段未正确拷贝")
	}
	if len(dto.Attrs) != len(entity.Attrs) || dto.Attrs["role"] != entity.Attrs["role"] {
		t.Errorf("Attrs 字段未正确拷贝")
	}
	if !dto.CreatedAt.Equal(entity.CreatedAt) {
		t.Errorf("CreatedAt 字段未正确拷贝")
	}
	if dto.UpdatedAt == nil || !dto.UpdatedAt.Equal(*entity.UpdatedAt) {
		t.Errorf("UpdatedAt 字段未正确拷贝")
	}
	if dto.Alias != int(entity.Alias) {
		t.Errorf("Alias 字段未正确拷贝")
	}
	if len(dto.Numbers) != len(entity.Numbers) || dto.Numbers[1] != entity.Numbers[1] {
		t.Errorf("Numbers 字段未正确拷贝")
	}
	if string(dto.Data) != string(entity.Data) {
		t.Errorf("Data 字段未正确拷贝")
	}
	if dto.Nested.Desc != entity.Nested.Desc || dto.Nested.Val != entity.Nested.Val {
		t.Errorf("Nested 字段未正确拷贝")
	}
	if dto.Addr.PostCode != entity.Addr.PostCode {
		t.Errorf("Addr.PostCode 字段未正确拷贝")
	}
	if dto.Addr.Location != entity.Addr.Location {
		t.Errorf("Addr.Location 字段未正确拷贝")
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
	type MyEntity struct {
		Foo int
	}
	type MyDTO struct {
		Bar int
	}
	Register(func(src MyEntity) (MyDTO, error) {
		return MyDTO{Bar: 100}, nil
	})
	entity := MyEntity{Foo: 1}
	result := Convert[MyEntity, MyDTO](entity)
	dto := result.Val()
	if result.Error() != nil {
		t.Fatalf("自定义转换器失败: %v", result.Error())
	}
	if dto.Bar != 100 {
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

func BenchmarkConvert_New(b *testing.B) {
	now := time.Now()
	nowPtr := &now
	entity := UserEntity{
		ID:        1,
		Name:      "Tom",
		Addr:      &AddrEntity{City: "Beijing", PostCode: 100000, Location: [2]float64{39.9042, 116.4074}},
		Age:       30,
		Score:     99.5,
		Active:    true,
		Tags:      []string{"golang", "dev"},
		Attrs:     map[string]string{"role": "admin", "team": "backend"},
		CreatedAt: now,
		UpdatedAt: nowPtr,
		Any:       map[string]any{"k": 1},
		Alias:     42,
		Numbers:   []int{1, 2, 3},
		Data:      []byte("hello world"),
		Nested:    NestedEntity{Desc: "嵌套结构体", Val: 888},
	}
	for i := 0; i < b.N; i++ {
		dto := &UserDTO{
			ID:   entity.ID,
			Name: entity.Name,
			Addr: &AddrDTO{
				City:     entity.Addr.City,
				PostCode: entity.Addr.PostCode,
				Location: entity.Addr.Location,
			},
			Age:       entity.Age,
			Score:     entity.Score,
			Active:    entity.Active,
			Tags:      entity.Tags,
			Attrs:     entity.Attrs,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			Any:       entity.Any,
			Alias:     int(entity.Alias),
			Numbers:   entity.Numbers,
			Data:      entity.Data,
			Nested:    NestedDTO{Desc: entity.Nested.Desc, Val: entity.Nested.Val},
		}
		_ = dto
	}
}

func BenchmarkConvert_JSON(b *testing.B) {
	entity := UserEntity{
		ID:   1,
		Name: "Tom",
		Addr: &AddrEntity{City: "Beijing"},
	}
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(entity)
		if err != nil {
			b.Fatal(err)
		}
		var dto UserDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			b.Fatal(err)
		}
		_ = dto
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

func BenchmarkConvert_CustomConverter(b *testing.B) {
	type MyEntity struct {
		Foo int
	}
	type MyDTO struct {
		Bar int
	}
	Register(func(src MyEntity) (MyDTO, error) {
		return MyDTO{Bar: src.Foo + 1}, nil
	})
	entity := MyEntity{Foo: 123}
	for i := 0; i < b.N; i++ {
		result := Convert[MyEntity, MyDTO](entity)
		if result.Error() != nil {
			b.Fatal(result.Error())
		}
		_ = result.Val()
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
