package main

import (
	"fmt"
	lorago "github.com/LorraineWen/lorago/router"
)

func main() {
	engine := lorago.New()
	userGroup := engine.Group("user")
	userGroup.Get("/name", func(context *lorago.Context) {
		context.W.Write([]byte("get amie"))
	})
	userGroup.PreHandleMiddleware(func(next lorago.HandleFunc) lorago.HandleFunc {
		return func(ctx *lorago.Context) {
			fmt.Fprintln(ctx.W, "Pre Handle:")
			next(ctx)
		}
	})
	userGroup.PostHandleMiddleware(func(next lorago.HandleFunc) lorago.HandleFunc {
		return func(ctx *lorago.Context) {
			fmt.Fprintln(ctx.W, "Post Handle:")
			next(ctx)
		}
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
