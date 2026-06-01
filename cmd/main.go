package main

import (
	"net/http"
	"rate-limiter/api"
)

func main() {
	handler := http.NewServeMux()

	app := api.NewApi(handler)

	if err := app.Run(); err != nil {
		panic(err)
	}

}
