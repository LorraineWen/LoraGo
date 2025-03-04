package main

import (
	"fmt"
	lorago "github.com/LorraineWen/lorago/router"
	log "github.com/LorraineWen/lorago/router/log"
	"net/http"
	"os"
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
	fmt.Println(os.Getwd())
	userGroup := engine.Group("user")
	userGroup.Use(log.LogMiddleware)
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
	logger := log.NewLogger()
	logger.Level = log.LevelDebug
	userGroup.Post("/index1", func(ctx *lorago.Context) {
		user := &User{}
		err := ctx.BindXml(user)
		logger.Debug("debug")
		logger.Info("info")
		logger.Error("error")
		if err == nil {
			ctx.JsonResponseWrite(http.StatusOK, user)
		} else {
			fmt.Println(err)
		}
	})
	engine.Run()
}
