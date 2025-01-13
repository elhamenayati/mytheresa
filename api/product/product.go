package product

import (
	"net/http"

	"github.com/elhamenayati/mytheresa/internal"
	"github.com/elhamenayati/mytheresa/model"
	"github.com/labstack/echo"
)

func List(c echo.Context) error {
	category := c.QueryParam("category")
	priceLessThan := c.QueryParam("priceLessThan")

	product := new(model.Product)
	products, err := product.LoadByParam(category, priceLessThan)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in loading products")
	}

	var rest []internal.Product
	for _, p := range products {
		rest = append(rest, p.Rest())
	}

	return c.JSON(http.StatusOK, rest)
}
