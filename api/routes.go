package api

import (
	"github.com/elhamenayati/mytheresa/api/product"
	"github.com/elhamenayati/mytheresa/server"
)

var BaseURL = "/api"

func productRoutes() {
	gp := server.BP.Group(BaseURL + "/product")
	gp.GET("", product.List)
}

func Serve() {
	productRoutes()
}
