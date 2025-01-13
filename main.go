package main

import (
	"github.com/elhamenayati/mytheresa/api"
	"github.com/elhamenayati/mytheresa/inits"
	"github.com/elhamenayati/mytheresa/server"
)

func main() {
	inits.Init()
	api.Serve()
	server.Run()
}
