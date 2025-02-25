package main

import (
	web_frame "github.com/LorraineWen/WebFrame"
)

func main() {
	engine := web_frame.New()
	userGroup := engine.Group("user")
	userGroup.Get("/name", func(context *web_frame.Context) {
		context.W.Write([]byte("get amie"))
	})
	userGroup.Post("/name", func(context *web_frame.Context) {
		context.W.Write([]byte("post amie"))
	})
	engine.Run()
}
