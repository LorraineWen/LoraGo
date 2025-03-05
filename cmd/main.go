package main

import (
	"fmt"
	lorago "github.com/LorraineWen/lorago/router"
	"github.com/LorraineWen/lorago/router/log"
	"net/http"
)

type User struct {
	Name string `json:"name" validate:"required" binding:"required" xml:"name"`
	Age  int    `json:"age" validate:"required,max=50,min=18" xml:"age"`
}
type RespError struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func main() {
	engine := lorago.New()
	engine.Logger.Level = log.LevelDebug
	engine.Logger.Formatter = log.TextFormatter{}
	userGroup := engine.Group("user")
	userGroup.Use(lorago.LogMiddleware)
	userGroup.Post("/index2", func(context *lorago.Context) {
		user := []User{}
		context.ValidateAnother = true
		context.DisallowUnknownFields = true
		context.Validate = true
		err := context.BindJson(&user)
		if err != nil {
			context.JsonResponseWrite(http.StatusInternalServerError, &RespError{Code: http.StatusInternalServerError, Error: err.Error()})
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
	engine.Run()
}
