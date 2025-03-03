package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LorraineWen/lorago/router/render"
	"github.com/LorraineWen/lorago/util"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
)

/*
*@Author: LorraineWen
*@Date: 2025/2/26
*该文件主要提供上下文有关的接口
*由于http.ResponseWriter存放于Context中，所以Context应该提供接口进行html模板渲染
*还需要支持json，xml等格式的返回
*支持下载文件的需求，可以自定义下载的文件的名称
*支持json格式，切片格式的请求参数解析
 */
const defaultMultipartMemory = 30 << 20 //30MB大小用来加载post表单里面的参数到内存

type Context struct {
	W                     http.ResponseWriter
	R                     *http.Request
	e                     *Engine    //用于获取模板渲染函数
	queryCache            url.Values //用于获取请求路径中的参数，实际上就是map[string[]string
	formCache             url.Values //用于获取post请求中的表单数据
	DisallowUnknownFields bool       //设置参数属性检查，json参数中有的属性，如果绑定的结构体没有就报错
	Validate              bool       //设置结构体属性检查，如果json参数中没有该结构体的相应属性，那么就会报错
}

// 一个多态函数，htmlRender等结构体实现了Render函数，因此可以传入htmlRender等接口体，调用它们自己的Render函数，编码html等响应格式
func (c *Context) Render(status int, r render.Render) error {
	//由于Render函数的底层是w.Write等函数，这些函数写完之后就自动将Header里面的状态码置为200了
	//所以不需要重复进行写入，只要将非200的状态码写入就行了，w.Write检测到已经写入过其他状态码了，就不会再写入200了
	if status != http.StatusOK {
		c.W.WriteHeader(status)
	}
	err := r.Render(c.W)
	return err
}

// 支持html格式响应
func (ctx *Context) HtmlResponseWrite(status int, data any) error {
	err := ctx.Render(status, &render.HtmlRender{IsTemplate: false, Data: data})
	return err
}

// 支持json格式响应
// 调用方式context.JsonResponseWrite(http.StatusOK, &User{Name: "amie"})
func (ctx *Context) JsonResponseWrite(status int, data any) error {
	err := ctx.Render(status, &render.JsonRender{Data: data})
	return err
}

// 支持xml格式响应
// 调用方式context.Xml(http.StatusOK, &User{Name: "amie"})
func (ctx *Context) XmlResponseWrite(status int, data any) error {
	err := ctx.Render(status, &render.XmlRender{Data: data})
	return err
}

// 支持格式化String格式响应
// 调用方式context.StringResponseWrite(http.StatusOK, "你好 %s", "amie")
func (ctx *Context) StringResponseWrite(status int, format string, data ...any) (err error) {
	err = ctx.Render(status, &render.StringRender{
		Format: format,
		Data:   data,
	})
	return
}

// 配合e.LoadTemplate函数使用，e.LoadTemplate初始化 ctx.e.htmlRender.Template，Template只需要传递该模板需要的数据
// 调用方式:
// engine.LoadTemplate("../test/template/*.html")
//
//	userGroup.Get("/index", func(context *lorago.Context) {
//		context.TemplateResponseWrite(http.StatusOK, "index.html", &User{Name: "amie"})
//	})
//
// name是../test/template目录下的具体html文件的名称
func (ctx *Context) TemplateResponseWrite(status int, name string, data any) error {
	err := ctx.Render(status, &render.HtmlRender{
		IsTemplate: true,
		Name:       name,
		Data:       data,
		Template:   ctx.e.htmlRender.Template,
	})
	return err
}

// 支持文件下载
func (ctx *Context) FileResponseWrite(filePath string) {
	http.ServeFile(ctx.W, ctx.R, filePath)
}

// 支持自定义文件名称下载，下载好的文件名称自动变为filename
func (ctx *Context) FileAttachmentResponseWrite(filepath, filename string) {
	if util.IsASCII(filename) {
		ctx.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		ctx.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(ctx.W, ctx.R, filepath)
}

// 从本地文件系统下载文件，fileSystem实际上就是一个本地的目录http.Dir("../test/template")
// filePath就是以template作为根目录了
func (ctx *Context) FileFromFileSystemResponseWrite(filePath string, fileSystem http.FileSystem) {
	defer func(old string) {
		ctx.R.URL.Path = old
	}(ctx.R.URL.Path)

	ctx.R.URL.Path = filePath

	http.FileServer(fileSystem).ServeHTTP(ctx.W, ctx.R)
}

// 路由重定向支持
func (ctx *Context) Redirect(status int, location string) {
	//由于http.Redirect的重定向只对部分状态码有效果，因此需要对状态码进行判断
	if (status < http.StatusMultipleChoices || status > http.StatusPermanentRedirect) && status != http.StatusCreated {
		panic(fmt.Sprintf("在该状态下无法进行重定向 %d", status))
	}
	http.Redirect(ctx.W, ctx.R, location, status)
}

// 将请求路径中的参数，按照map[string][]string的格式存储到c.queryCache中
func (ctx *Context) initQueryCache() {
	if ctx.R != nil {
		ctx.queryCache = ctx.R.URL.Query() //底层方法就是这个
	} else {
		ctx.queryCache = url.Values{}
	}
}

// 获取请求参数/adduser?id=1&name=amie
// 调用方式:GetQuery("id")
func (ctx *Context) GetQuery(key string) string {
	ctx.initQueryCache() //加载路径参数
	return ctx.queryCache.Get(key)
}

// 获取数组类型的请求参数，一个key对应多个value
// 一个key多个value:/adduser?id=1&id=2&id=3
// 调用方式:GetQueryArray("id")
func (ctx *Context) GetQueryArray(key string) (values []string, ok bool) {
	ctx.initQueryCache() //加载路径参数
	values, ok = ctx.queryCache[key]
	return
}

// 获取map格式的请求参数/adduser?user[id]=1&user[name]=amie
// 目前只能获取string:string类型的，可以进行数据转换支持更多类型
// 调用方式:GetQueryMap("user")
func (ctx *Context) GetQueryMap(key string) (map[string]string, bool) {
	ctx.initQueryCache() //加载路径参数
	res := make(map[string]string)
	exist := false
	for requestKey, value := range ctx.queryCache { //requestKey="user[id]"而不是"id"
		if index1 := strings.IndexByte(requestKey, '['); index1 >= 1 { //大于等于1，是为了避免[出现在user前面
			if index2 := strings.IndexByte(requestKey, ']'); index2 >= index1 {
				if requestKey[0:index1] == key { //key="user"而不是"id"
					exist = true
					res[requestKey[index1+1:index2]] = value[len(value)-1] //一个key对应多个value的情况下，只能获取最后一个值，只是map本身的特性，后面的覆盖前面的
				}
			}
		}
	}
	return res, exist
}

// 为请求参数设置一个默认值
// 调用方式DefaultQuery("id","1")
func (ctx *Context) DefaultQuery(key, defaultValue string) string {
	array, ok := ctx.GetQueryArray(key)
	if !ok {
		return defaultValue
	}
	return array[0]
}

func (ctx *Context) initFormCache() {
	if ctx.formCache == nil {
		ctx.formCache = make(url.Values)
		req := ctx.R
		if err := req.ParseMultipartForm(defaultMultipartMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println(err)
			}
		}
		ctx.formCache = ctx.R.PostForm
	}
}

// 获取post请求中的参数，由于调用的是ParseMultipartForm，所以只能下面这种格式的post请求
// ### POST test post router
// POST http://localhost:8080/user/index
// Content-Type: application/x-www-form-urlencoded
//
// id=1
// 调用方式:GetQuery("id")
func (ctx *Context) GetFormQuery(key string) string {
	ctx.initFormCache() //加载路径参数
	return ctx.formCache.Get(key)
}

// ### POST test post router
// POST http://localhost:8080/user/index2
// Content-Type: application/x-www-form-urlencoded
//
// id=1&id=2
// 调用方式:GetQueryArray("id")
func (ctx *Context) GetFormQueryArray(key string) (values []string, ok bool) {
	ctx.initFormCache() //加载路径参数
	values, ok = ctx.formCache[key]
	return
}

// POST http://localhost:8080/user/index3
// Content-Type: application/x-www-form-urlencoded
//
// user[id]=1&user[id]=2&user[name]="amie"
// 调用方式:GetQueryMap("user")
func (ctx *Context) GetFormQueryMap(key string) (map[string]string, bool) {
	ctx.initFormCache() //加载路径参数
	res := make(map[string]string)
	exist := false
	for requestKey, value := range ctx.formCache { //requestKey="user[id]"而不是"id"
		if index1 := strings.IndexByte(requestKey, '['); index1 >= 1 { //大于等于1，是为了避免[出现在user前面
			if index2 := strings.IndexByte(requestKey, ']'); index2 >= index1 {
				if requestKey[0:index1] == key { //key="user"而不是"id"
					exist = true
					res[requestKey[index1+1:index2]] = value[len(value)-1] //一个key对应多个value的情况下，只能获取最后一个值，只是map本身的特性，后面的覆盖前面的
				}
			}
		}
	}
	return res, exist
}

// 调用方式DefaultQuery("id","1")
func (ctx *Context) DefaultFormQuery(key, defaultValue string) string {
	array, ok := ctx.GetFormQueryArray(key)
	if !ok {
		return defaultValue
	}
	return array[0]
}
func (ctx *Context) FormFile(name string) (*multipart.FileHeader, error) {
	req := ctx.R
	if err := req.ParseMultipartForm(defaultMultipartMemory); err != nil {
		return nil, err
	}
	file, header, err := req.FormFile(name)
	if err != nil {
		return nil, err
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	return header, nil
}
func (ctx *Context) MultipartForm() (*multipart.Form, error) {
	err := ctx.R.ParseMultipartForm(defaultMultipartMemory)
	return ctx.R.MultipartForm, err
}
func (ctx *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

// 解析post请求中的json格式数据
// 如果要解析属性校验，需要在注册路由的时候，将Validate两个bool值设置为true
func (ctx *Context) BindJson(data any) error {
	body := ctx.R.Body
	if ctx.R == nil || body == nil {
		return errors.New("请求错误")
	}
	decoder := json.NewDecoder(body)
	//如果启用了请求参数属性检测，那么就会检查请求参数中的属性，在相应结构体中是否存在
	if ctx.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	//如果启用了请求参数属性检测，那么就会检查结构体中的属性，在相应请求参数中是否存在
	if ctx.Validate {
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
		default:
			err := decoder.Decode(data)
			if err != nil {
				return err
			}
		}
	}
	return decoder.Decode(data)
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
