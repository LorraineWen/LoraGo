package lora_bind

import (
	"fmt"
	"github.com/go-playground/validator"
	"reflect"
	"strings"
	"sync"
)

var Validator = &defaultValidator{}

// 自定义一个验证器
type LoraValidator interface {
	//结构体验证，如果错误返回对应的错误信息
	ValidateStruct(any) error
	//返回对应使用的验证器
	Engine() any
}

// 使用第三方的验证器
type defaultValidator struct {
	one      sync.Once
	validate *validator.Validate
}

// 集成一个第三方的post请求属性验证库
type SliceValidationError []error

func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
				}
			}
		}
		return b.String()
	}
}
func (d *defaultValidator) Engine(any) any {
	d.one.Do(func() {
		d.validate = validator.New()
	})
	return d.validate
}
func (d *defaultValidator) validateStruct(obj any) error {
	d.one.Do(func() {
		d.validate = validator.New()
	})
	return validator.New().Struct(obj) //第三方的validate实例
}
func (d *defaultValidator) validateAllParams(data any) error {
	if data == nil {
		return nil
	}
	value := reflect.ValueOf(data)
	switch value.Kind() {
	case reflect.Ptr:
		return d.validateAllParams(value.Elem().Interface())
	case reflect.Struct:
		return d.validateStruct(data)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := d.validateStruct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	default:
		return nil
	}
}
