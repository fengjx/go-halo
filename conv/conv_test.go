package conv

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
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
