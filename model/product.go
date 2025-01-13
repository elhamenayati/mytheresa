package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/elhamenayati/mytheresa/internal"
	"github.com/elhamenayati/mytheresa/server"
)

type Product struct {
	Sku      string `json:"sku" gorm:"primary_key"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Price    int    `json:"price"`
}

func (Product) TableName() string {
	return "products"
}

func (p *Product) Save() error {
	return server.DB.Table(p.TableName()).Save(&p).Error
}

func (p *Product) LoadByParam(category, price string) (products []Product, err error) {
	query := server.DB.Table(p.TableName())

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if price != "" {
		query = query.Where("price <= ?", price)
	}

	err = query.Find(&products).Error

	return products, err
}

func (p *Product) Rest() internal.Product {
	product := internal.Product{}
	if p.Category == "boots" {
		price := CalculateDiscount([]int{30, 15}, p.Price)

		product = internal.Product{
			Sku:      p.Sku,
			Name:     p.Name,
			Category: p.Category,
			Price:    price,
		}
	} else if p.Category != "boots" {
		price := CalculateDiscount([]int{0}, p.Price)

		product = internal.Product{
			Sku:      p.Sku,
			Name:     p.Name,
			Category: p.Category,
			Price:    price,
		}
	}

	return product
}

func CalculateDiscount(discount []int, price int) internal.ProductPrice {
	///find maximum discount
	maxDiscount := 0
	for _, dis := range discount {
		if dis >= maxDiscount {
			maxDiscount = dis
		}
	}

	var percenatge *string
	if maxDiscount != 0 {
		p := fmt.Sprintf("%d", maxDiscount) + "%"
		percenatge = &p
	} else {
		percenatge = nil
	}

	return internal.ProductPrice{
		Original:           price,
		Final:              price - (price * maxDiscount / 100),
		DiscountPercentage: percenatge,
		Currency:           "EUR",
	}
}

func validatePrice(price string) error {
	if !strings.Contains(price, ".") {
		return fmt.Errorf("invalid price format")
	}

	parts := strings.Split(price, ".")
	if len(parts) != 2 {
		return errors.New("invalid price format")
	}
	if _, err := strconv.Atoi(parts[0]); err != nil {
		return err
	}
	if _, err := strconv.Atoi(parts[1]); err != nil {
		return err
	}

	return nil
}

func ConvertPrice(price string) (int, error) {
	if err := validatePrice(price); err != nil {
		return 0, err
	}

	parts := strings.Split(price, ".")
	cents, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	euros, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	return euros*100 + cents, nil
}
