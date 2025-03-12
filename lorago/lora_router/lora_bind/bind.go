package lora_bind

import "net/http"

/*
*@Author: LorraineWen
*绑定器接口，后续可以支持对各类格式参数的校验
 */
type Binder interface {
	Name() string
	Bind(*http.Request, any) error
}

func validateAllParams(data any) error {
	return Validator.validateAllParams(data)
}

var JsonBinder = jsonBinder{}
var XmlBinder = xmlBinder{}
