package test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/elhamenayati/mytheresa/api/product"
	"github.com/elhamenayati/mytheresa/internal"
	"github.com/elhamenayati/mytheresa/model"
	"github.com/elhamenayati/mytheresa/server"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestConvertPrice(t *testing.T) {
	tests := []struct {
		name           string
		price          string
		expectedResult int
	}{
		{
			name:           "Invalid format",
			price:          "50",
			expectedResult: 0,
		},
		{
			name:           "Format with no cents",
			price:          "50.0",
			expectedResult: 5000,
		},
		{
			name:           "Format with trailing zeros",
			price:          "52.00",
			expectedResult: 5200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := model.ConvertPrice(tt.price)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestCalculateDiscount(t *testing.T) {
	tests := []struct {
		name     string
		discount []int
		price    int
		expected internal.ProductPrice
	}{
		{
			name:     "No discount",
			discount: []int{0},
			price:    100,
			expected: internal.ProductPrice{Original: 100, Final: 100, DiscountPercentage: nil, Currency: "EUR"},
		},
		{
			name:     "Single discount",
			discount: []int{30},
			price:    100,
			expected: internal.ProductPrice{Original: 100, Final: 70, DiscountPercentage: stringPtr("30%"), Currency: "EUR"},
		},
		{
			name:     "Multiple discounts, take max",
			discount: []int{10, 20, 30},
			price:    100,
			expected: internal.ProductPrice{Original: 100, Final: 70, DiscountPercentage: stringPtr("30%"), Currency: "EUR"},
		},
		{
			name:     "Multiple discounts, different order",
			discount: []int{30, 15, 5},
			price:    50,
			expected: internal.ProductPrice{Original: 50, Final: 35, DiscountPercentage: stringPtr("30%"), Currency: "EUR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.CalculateDiscount(tt.discount, tt.price)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProduct_Rest(t *testing.T) {
	tests := []struct {
		name     string
		product  model.Product
		expected internal.Product
	}{
		{
			name:    "Boots product",
			product: model.Product{Sku: "sku1", Name: "Boot1", Category: "boots", Price: 100},
			expected: internal.Product{Sku: "sku1", Name: "Boot1", Category: "boots", Price: internal.ProductPrice{
				Original:           100,
				Final:              70, // Assuming 30% max discount for boots
				DiscountPercentage: stringPtr("30%"),
				Currency:           "EUR",
			}},
		},
		{
			name:    "Non-boots product",
			product: model.Product{Sku: "sku2", Name: "Shirt1", Category: "shirts", Price: 50},
			expected: internal.Product{Sku: "sku2", Name: "Shirt1", Category: "shirts", Price: internal.ProductPrice{
				Original:           50,
				Final:              50, //No Discount
				DiscountPercentage: nil,
				Currency:           "EUR",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.product.Rest()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProduct_LoadByParam(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when initializing GORM with sqlmock", err)
	}

	server.DB = gormDB // Inject the mock DB into your server

	tests := []struct {
		name          string
		category      string
		priceLessThan string
		mockRows      *sqlmock.Rows
		mockError     error
		expected      []model.Product
		expectedErr   error
		prepareMock   func()
	}{
		{
			name:          "No params",
			category:      "",
			priceLessThan: "",
			mockRows:      sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("1230", "sample boot", "boots", 89000),
			expected:      []model.Product{{Sku: "1230", Name: "sample boot", Category: "boots", Price: 89000}},
			expectedErr:   nil,
			prepareMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `products`").WillReturnRows(sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("1230", "sample boot", "boots", 89000))
			},
		},
		{
			name:          "No params",
			category:      "",
			priceLessThan: "",
			mockRows:      nil,
			expected:      []model.Product{},
			expectedErr:   nil,
			prepareMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `products`").WillReturnRows(sqlmock.NewRows([]string{"sku", "name", "category", "price"}))
			},
		},
		{
			name:          "Category param",
			category:      "boots",
			priceLessThan: "",
			mockRows:      sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("000002", "BV Lean leather ankle boots", "boots", 99000),
			expected:      []model.Product{{Sku: "000002", Name: "BV Lean leather ankle boots", Category: "boots", Price: 99000}},
			expectedErr:   nil,
			prepareMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `products` WHERE \\(category = \\?\\)$").WithArgs("boots").WillReturnRows(sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("000002", "BV Lean leather ankle boots", "boots", 99000))
			},
		},
		{
			name:          "Price param",
			category:      "",
			priceLessThan: "60",
			mockRows:      sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("1255", "Product1", "shirts", 50),
			expected:      []model.Product{{Sku: "1255", Name: "Product1", Category: "shirts", Price: 50}},
			expectedErr:   nil,
			prepareMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `products` WHERE \\(price <= \\?\\)$").WithArgs("60").WillReturnRows(sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("1255", "Product1", "shirts", 50))
			},
		},
		{
			name:          "Both params",
			category:      "shirts",
			priceLessThan: "50",
			mockRows:      sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("sku1", "Product1", "shirts", 50),
			expected:      []model.Product{{Sku: "sku1", Name: "Product1", Category: "shirts", Price: 50}},
			expectedErr:   nil,
			prepareMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `products` WHERE \\(category = \\?\\) AND \\(price <= \\?\\)$").WithArgs("shirts", "50").WillReturnRows(sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("sku1", "Product1", "shirts", 50))
			},
		},
		{
			name:          "No Data Found",
			category:      "test",
			priceLessThan: "10",
			expected:      []model.Product{},
			expectedErr:   nil,
			prepareMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `products` WHERE \\(category = \\?\\) AND \\(price <= \\?\\)$").WithArgs("test", "10").WillReturnRows(sqlmock.NewRows([]string{"sku", "name", "category", "price"}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := new(model.Product)

			if tt.prepareMock != nil {
				tt.prepareMock() // Prepare mock expectations *before* calling the function
			}

			products, err := p.LoadByParam(tt.category, tt.priceLessThan)

			// Check for expected errors
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error(), "LoadByParam() error")
				assert.Nil(t, products, "LoadByParam() should return nil products on error")
			} else {
				assert.NoError(t, err, "LoadByParam() unexpected error")
				assert.Equal(t, tt.expected, products, "LoadByParam() returned unexpected products")
			}

		})
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestList(t *testing.T) {
	// Setup mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)

	if err != nil {
		t.Fatalf("an error '%s' was not expected when init gorm", err)
		return
	}

	server.DB = gormDB // Inject the mock DB into your server

	tests := []struct {
		name           string
		category       string
		priceLessThan  string
		mockRows       *sqlmock.Rows
		mockError      error
		expectedStatus int
		expectedBody   []internal.Product
		prepareMock    func()
	}{
		{
			name:           "Successful request with no params",
			category:       "",
			priceLessThan:  "",
			mockRows:       sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("sku1", "Shirt1", "shirts", 50).AddRow("sku2", "Boot2", "boots", 100),
			expectedStatus: http.StatusOK,
			expectedBody: []internal.Product{
				{Sku: "sku1", Name: "Shirt1", Category: "shirts", Price: internal.ProductPrice{Original: 50, Final: 50, DiscountPercentage: nil, Currency: "EUR"}},
				{Sku: "sku2", Name: "Boot2", Category: "boots", Price: internal.ProductPrice{Original: 100, Final: 70, DiscountPercentage: stringPtr("30%"), Currency: "EUR"}},
			},
			prepareMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `products`").WillReturnRows(sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("sku1", "Shirt1", "shirts", 50).AddRow("sku2", "Boot2", "boots", 100))

			},
		},
		{
			name:           "Successful request with category param",
			category:       "boots",
			priceLessThan:  "",
			mockRows:       sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("sku2", "Boot2", "boots", 100),
			expectedStatus: http.StatusOK,
			expectedBody: []internal.Product{
				{Sku: "sku2", Name: "Boot2", Category: "boots", Price: internal.ProductPrice{Original: 100, Final: 70, DiscountPercentage: stringPtr("30%"), Currency: "EUR"}},
			},
			prepareMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `products` WHERE \\(category = \\?\\)$").WithArgs("boots").WillReturnRows(sqlmock.NewRows([]string{"sku", "name", "category", "price"}).AddRow("sku2", "Boot2", "boots", 100))
			},
		},
		{
			name:          "Successful request with priceLessThan param",
			category:      "",
			priceLessThan: "60",
			mockRows: sqlmock.NewRows([]string{"sku", "name", "category", "price"}).
				AddRow("sku1", "Product1", "shirts", 50),
			expectedStatus: http.StatusOK,
			expectedBody:   []internal.Product{{Sku: "sku1", Name: "Product1", Category: "shirts", Price: internal.ProductPrice{Original: 50, Final: 50, DiscountPercentage: nil, Currency: "EUR"}}},
			prepareMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `products` WHERE \\(price <= \\?\\)$").WithArgs("60").WillReturnRows(sqlmock.NewRows([]string{"sku", "name", "category", "price"}).
					AddRow("sku1", "Product1", "shirts", 50))
			},
		},
		{
			name:           "DB Error",
			category:       "",
			priceLessThan:  "",
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil,
			prepareMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `products`").WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "No Data Found",
			category:       "test",
			priceLessThan:  "10",
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
			prepareMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `products` WHERE \\(category = \\?\\) AND \\(price <= \\?\\)$").WithArgs("test", "10").WillReturnRows(sqlmock.NewRows([]string{"sku", "name", "category", "price"}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/products?category="+tt.category+"&priceLessThan="+tt.priceLessThan, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.prepareMock != nil {
				tt.prepareMock()
			}

			// Call the handler
			err := product.List(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedBody != nil {
				var responseBody []internal.Product
				err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
				assert.NoError(t, err)

				// Compare the response body
				assert.Equal(t, tt.expectedBody, responseBody)
			}

			if tt.expectedStatus == http.StatusInternalServerError {
				assert.Equal(t, "\"error in loading products\"\n", rec.Body.String())
			}

		})
	}
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}
