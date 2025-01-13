package server

import (
	"fmt"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	BP          = echo.New()
	Middlewares = []echo.MiddlewareFunc{
		middleware.Recover(),
		middleware.CORS(),
		middleware.GzipWithConfig(middleware.GzipConfig{
			Level: 5,
		}),
	}

	DB = MYSQLsConnection()
)

func Run() {
	setup()

	defer DB.Close()
}

func setup() {
	BP.Use(Middlewares...)
	BP.Logger.Fatal(BP.Start(fmt.Sprintf(":%d", 8080)))
}
