package internal

type Product struct {
	Sku      string       `json:"sku"`
	Name     string       `json:"name"`
	Category string       `json:"category"`
	Price    ProductPrice `json:"price"`
}

type ProductPrice struct {
	Original           int     `json:"original"`
	Final              int     `json:"final"`
	DiscountPercentage *string `json:"discount_percentage"`
	Currency           string  `json:"currency"`
}
