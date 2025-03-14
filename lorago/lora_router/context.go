package lora_router

import (
	"errors"
	"fmt"
	"github.com/LorraineWen/lorago/lora_bind"
	"github.com/LorraineWen/lorago/lora_log"
	"github.com/LorraineWen/lorago/lora_render"
	"github.com/LorraineWen/lorago/lora_util"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

/*
*@Author: LorraineWen
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
	engine                *Engine          //用于获取模板渲染函数
	StatusCode            int              //存放响应结果
	queryCache            url.Values       //用于获取请求路径中的参数，实际上就是map[string[]string
	formCache             url.Values       //用于获取post请求中的表单数据
	DisallowUnknownFields bool             //设置参数属性检查，json参数中有的属性，如果绑定的结构体没有就报错
	Validate              bool             //设置结构体属性检查，如果json参数中没有该结构体的相应属性，那么就会报错
	ValidateAnother       bool             //启用第三方的校验
	Logger                *lora_log.Logger //日志模块
	basicKeys             map[string]any   //用于basic身份验证，实际上是通过中间件实现basic验证
	rwMutex               sync.RWMutex     //用于basic身份验证的读写锁
	sameSite              http.SameSite    //用于jwt验证的安全验证
}

// 一个多态函数，htmlRender等结构体实现了Render函数，因此可以传入htmlRender等接口体，调用它们自己的Render函数，编码html等响应格式
func (ctx *Context) Render(status int, r lora_render.Render) error {
	err := r.Render(ctx.W, status)
	ctx.StatusCode = status
	return err
}

// 支持html格式响应
func (ctx *Context) HtmlResponseWrite(status int, data any) error {
	err := ctx.Render(status, &lora_render.HtmlRender{IsTemplate: false, Data: data})
	return err
}

// 支持json格式响应
// 调用方式context.JsonResponseWrite(http.StatusOK, &User{Name: "amie"})
func (ctx *Context) JsonResponseWrite(status int, data any) error {
	err := ctx.Render(status, &lora_render.JsonRender{Data: data})
	return err
}

// 支持xml格式响应
// 调用方式context.Xml(http.StatusOK, &User{Name: "amie"})
func (ctx *Context) XmlResponseWrite(status int, data any) error {
	err := ctx.Render(status, &lora_render.XmlRender{Data: data})
	return err
}

// 支持格式化String格式响应
// 调用方式context.StringResponseWrite(http.StatusOK, "你好 %s", "amie")
func (ctx *Context) StringResponseWrite(status int, format string, data ...any) (err error) {
	err = ctx.Render(status, &lora_render.StringRender{
		Format: format,
		Data:   data,
	})
	return
}

// 配合e.LoadTemplate函数使用，engine.LoadTemplate初始化 ctx.engine.htmlRender.Template，Template只需要传递该模板需要的数据
// 调用方式:
// engine.LoadTemplate("../test/template/*.html")
//
//	userGroup.Get("/index", func(context *lorago.Context) {
//		context.TemplateResponseWrite(http.StatusOK, "index.html", &User{Name: "amie"})
//	})
//
// name是../test/template目录下的具体html文件的名称
func (ctx *Context) TemplateResponseWrite(status int, name string, data any) error {
	err := ctx.Render(status, &lora_render.HtmlRender{
		IsTemplate: true,
		Name:       name,
		Data:       data,
		Template:   ctx.engine.htmlRender.Template,
	})
	return err
}

// 支持文件下载
func (ctx *Context) FileResponseWrite(filePath string) {
	http.ServeFile(ctx.W, ctx.R, filePath)
}

// 支持自定义文件名称下载，下载好的文件名称自动变为filename
func (ctx *Context) FileAttachmentResponseWrite(filepath, filename string) {
	if lora_util.IsASCII(filename) {
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
				fmt.Println(err)
			}
		}
		ctx.formCache = ctx.R.PostForm
	}
}

// 获取post请求中的参数，由于调用的是ParseMultipartForm，所以只能下面这种格式的post请求
// ### POST test post lora_router
// POST http://localhost:8080/user/index
// Content-Type: application/x-www-form-urlencoded
//
// id=1
// 调用方式:GetQuery("id")
func (ctx *Context) GetFormQuery(key string) string {
	ctx.initFormCache() //加载路径参数
	return ctx.formCache.Get(key)
}

// ### POST test post lora_router
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
	jsonBinder := lora_bind.JsonBinder
	jsonBinder.DisallowUnknownFields = ctx.DisallowUnknownFields
	jsonBinder.IsValidate = ctx.Validate
	jsonBinder.IsValidateAnother = ctx.ValidateAnother
	return ctx.MustBindWith(data, jsonBinder) //多态底层调用json格式校验
}

// 支持xml格式校验
func (ctx *Context) BindXml(obj any) error {
	return ctx.MustBindWith(obj, lora_bind.XmlBinder)
}
func (ctx *Context) MustBindWith(obj any, b lora_bind.Binder) error {
	//如果发生错误，返回400状态码 参数错误
	if err := ctx.ShouldBindWith(obj, b); err != nil {
		return err
	}
	return nil
}
func (ctx *Context) ShouldBindWith(obj any, b lora_bind.Binder) error {
	return b.Bind(ctx.R, obj)
}
func (ctx *Context) Fail(status int, msg string) {
	ctx.StringResponseWrite(status, msg)
}

// 支持httpcode的设置，可以在Header中设置状态码和code
func (ctx *Context) ErrorHandle(err error) {
	code, data := ctx.engine.errHandler(err)
	ctx.JsonResponseWrite(code, data)
}

func (ctx *Context) HandlerWithError(code int, obj any, err error) {
	if err != nil {
		statusCode, data := ctx.engine.errHandler(err)
		ctx.JsonResponseWrite(statusCode, data)
		return
	}
	ctx.JsonResponseWrite(code, obj)
}
func (ctx *Context) BasicSet(key string, value any) {
	ctx.rwMutex.Lock()
	defer ctx.rwMutex.Unlock()
	if ctx.basicKeys == nil {
		ctx.basicKeys = make(map[string]interface{})
	}
	ctx.basicKeys[key] = value
}
func (ctx *Context) BasicGet(key string) (value any, exist bool) {
	ctx.rwMutex.Lock()
	defer ctx.rwMutex.Unlock()
	value, exist = ctx.basicKeys[key]
	return
}

// basic验证特有的"Authorization: Basic ${basic}"验证格式
func (c *Context) SetBasicAuth(username, password string) {
	c.R.Header.Set("Authorization", "Basic "+BasicAuth(username, password))
}

// 设置Cookie
func (ctx *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(ctx.W, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: ctx.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}
func (c *Context) GetCookie(name string) string {
	cookie, err := c.R.Cookie(name)
	if err != nil {
		return ""
	}
	if cookie != nil {
		return cookie.Value
	}
	return ""
}

// 用于设置SameSet
func (ctx *Context) SetSameSet(sameSet http.SameSite) {
	ctx.sameSite = sameSet
}
