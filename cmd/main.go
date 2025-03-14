package main

import (
	lorago "github.com/LorraineWen/lorago/lora_router"
	"net/http"
)

func main() {
	engine := lorago.Default()
	userGroup := engine.Group("user")
	userGroup.Get("/index2", func(context *lorago.Context) {
		panic("jifei")
		user := make(map[string]string)
		user["name"] = "amie"
		user["age"] = "18"
		context.JsonResponseWrite(http.StatusOK, user)
	})
	engine.Run()
}
