package main

import (
	lorago "github.com/LorraineWen/lorago"
)

func main() {
	engine := lorago.New()
	userGroup := engine.Group("user")
	userGroup.Get("/name/:id", func(context *lorago.Context) {
		context.W.Write([]byte("get amie"))
	})
	userGroup.Get("/name/*/getname", func(context *lorago.Context) {
		context.W.Write([]byte("get amie"))
	})
	userGroup.Post("/name", func(context *lorago.Context) {
		context.W.Write([]byte("post amie"))
	})
	userGroup.Delete("/name", func(context *lorago.Context) {
		context.W.Write([]byte("delete amie"))
	})
	userGroup.Put("/name", func(context *lorago.Context) {
		context.W.Write([]byte("put amie"))
	})
	engine.Run()
}
