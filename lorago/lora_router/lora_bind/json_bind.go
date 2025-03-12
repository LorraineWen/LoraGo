package lora_bind

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

/*
*@Author: LorraineWen
*定义json格式绑定器
*支持对结构体属性的校验和请求参数属性的校验
 */
type jsonBinder struct {
	DisallowUnknownFields bool
	IsValidate            bool
	IsValidateAnother     bool
}

func (this jsonBinder) Name() string {
	return "json"
}
func (this jsonBinder) Bind(req *http.Request, data any) error {
	if req == nil || req.Body == nil {
		return errors.New("请求错误")
	}
	return this.decodeJson(req.Body, data)
}
func (this jsonBinder) decodeJson(body io.ReadCloser, data any) error {
	decoder := json.NewDecoder(body)
	//如果启用了请求参数属性检测，那么就会检查请求参数中的属性，在相应结构体中是否存在
	if this.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	//如果启用了请求参数属性检测，那么就会检查结构体中的属性，在相应请求参数中是否存在
	//自定义校验方式
	if this.IsValidate && !this.IsValidateAnother {
		if data == nil {
			return nil
		}
		valueOf := reflect.ValueOf(data)
		//判断传入进来的是否是结构体指针，因为只有指针才传值
		if valueOf.Kind() != reflect.Pointer {
			return errors.New("bind data must be a pointer")
		}
		t := valueOf.Elem().Interface()
		of := reflect.ValueOf(t)
		switch of.Kind() {
		//结构体类型的反射
		case reflect.Struct:
			return validateStructParams(of, data, decoder)
		//切片类型的反射
		case reflect.Slice, reflect.Array:
			elem := of.Type().Elem()
			elemType := elem.Kind()
			if elemType == reflect.Struct {
				return validateSliceParams(elem, data, decoder)
			}
		}
	}
	err := decoder.Decode(data)
	if err != nil {
		return err
	}
	//第三方，可以在结构体上面加上更多的标签，比如数值属性可以设置max和min作为取值范围
	if this.IsValidateAnother {
		err = validateAllParams(data)
		if err != nil {
			return err
		}
	}
	return nil
}
func validateSliceParams(elem reflect.Type, data any, decoder *json.Decoder) error {
	mapData := make([]map[string]interface{}, 0)
	_ = decoder.Decode(&mapData)
	if len(mapData) <= 0 {
		return nil
	}
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		tag := field.Tag.Get("json")
		value := mapData[0][tag]
		if value == nil && field.Tag.Get("binding") == "required" {
			return errors.New(fmt.Sprintf("filed [%s] is required", tag))
		}
	}
	if data != nil {
		marshal, _ := json.Marshal(mapData)
		err := json.Unmarshal(marshal, data)
		if err != nil {
			return err
		}
	}
	return nil
}
func validateStructParams(of reflect.Value, data any, decoder *json.Decoder) error {
	mapData := make(map[string]interface{})
	err := decoder.Decode(&mapData)
	if err != nil {
		return err
	}
	for i := 0; i < of.NumField(); i++ {
		field := of.Type().Field(i)
		tag := field.Tag.Get("json") //获取结构体标签对应的值
		value := mapData[tag]
		if value == nil && field.Tag.Get("binding") == "required" { //如果该属性在请求中没有，并且是必须属性就报错
			return errors.New(fmt.Sprintf("filed [%s] is not exist", tag))
		}
	}
	marshal, err := json.Marshal(mapData)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshal, data)
	return err
}
