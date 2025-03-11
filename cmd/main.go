package main

import (
	"fmt"
	lorago "github.com/LorraineWen/lorago/router"
	"github.com/LorraineWen/lorago/router/lora_log"
	"github.com/LorraineWen/lorago/router/lora_pool"
	"net/http"
	"time"
)

type User struct {
	Name string `json:"name" validate:"required" binding:"required" xml:"name"`
	Age  int    `json:"age" validate:"required,max=50,min=18" xml:"age"`
}
type RespError struct {
	Code  int    `json:"code"`
	Error string `json:"lora_error"`
}
type BlogError struct {
	Success bool
	Code    int64
	Data    any
	Msg     string
}

type BlogNoDataError struct {
	Success bool
	Code    int64
	Msg     string
}

func (b *BlogError) Error() string {
	return b.Msg
}

func (b *BlogError) Fail(code int64, msg string) {
	b.Success = false
	b.Code = code
	b.Msg = msg
}

func (b *BlogError) Response() any {
	if b.Data == nil {
		return &BlogNoDataError{
			Success: b.Success,
			Code:    b.Code,
			Msg:     b.Msg,
		}
	}
	return b
}
func main() {
	engine := lorago.New()
	engine.Logger.Level = lora_log.LevelDebug
	engine.Logger.Formatter = lora_log.TextFormatter{}
	userGroup := engine.Group("user")
	userGroup.Use(lorago.LogMiddleware)
	userGroup.Post("/index2", func(context *lorago.Context) {
		user := []User{}
		context.ValidateAnother = true
		context.DisallowUnknownFields = true
		context.Validate = true
		err := context.BindJson(&user)
		if err != nil {
			context.JsonResponseWrite(http.StatusOK, &RespError{Error: err.Error()})
			return
		}
		context.JsonResponseWrite(http.StatusOK, user)
	})
	var u1 *User
	userGroup.Post("/index1", func(ctx *lorago.Context) {
		ctx.Logger.Debug("hello")
		ctx.Logger.Info("hello1")
		u1.Age = 1
		user := &User{}
		err := ctx.BindXml(user)
		if err == nil {
			ctx.JsonResponseWrite(http.StatusOK, user)
		} else {
			fmt.Println(err)
		}
	})
	pool, err := lora_pool.NewPool(100)
	if err != nil {
		fmt.Println(err)
	}
	pool.Submit(func() {
		fmt.Println("1")
		panic("my panic")
		time.Sleep(2 * time.Second)
	})
	time.Sleep(2 * time.Second)
	pool.Submit(func() {
		fmt.Println("2")
		time.Sleep(2 * time.Second)
	})
	time.Sleep(2 * time.Second)
	pool.Submit(func() {
		fmt.Println("3")
		time.Sleep(2 * time.Second)
	})
	pool.Submit(func() {
		fmt.Println("4")
		time.Sleep(2 * time.Second)
	})
	pool.Submit(func() {
		fmt.Println("5")
		time.Sleep(2 * time.Second)
	})
	pool.Submit(func() {
		fmt.Println("6")
		time.Sleep(2 * time.Second)
	})
	engine.Run()
}
