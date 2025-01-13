package inits

import (
	"encoding/json"
	"io"

	"github.com/elhamenayati/mytheresa/internal"
	"github.com/elhamenayati/mytheresa/model"
	"github.com/elhamenayati/mytheresa/server"
	"github.com/jinzhu/gorm"
	"github.com/labstack/gommon/log"
)

func init() {
	autoMigrate(server.DB)

	///initialize products in database which reads sample data from products.json file
	data, err := internal.ReadData()
	if err != nil {
		log.Errorf("error in read products data from json file:%v", err)
	}
	defer data.Close()

	fileBytes, err := io.ReadAll(data)
	if err != nil {
		log.Errorf("error read object type into memory:%v", err)
	}

	var products []model.Product
	err = json.Unmarshal(fileBytes, &products)
	if err != nil {
		log.Errorf("error unmarshalling JSON:%v", err)
	}

	for i, p := range products {
		if err := p.Save(); err != nil {
			log.Errorf("error in adding product[%d] to database:%v", i, err)
		}
	}
}

func Init() {}

func autoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		&model.Product{},
	)
}
