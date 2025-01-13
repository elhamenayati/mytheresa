package server

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

func MYSQLsConnection() *gorm.DB {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	dsn := "%s:%s@tcp(%s:%d)/%s?parseTime=true"
	dsn = fmt.Sprintf(dsn,
		config.DBUser, config.DBPassword, config.DBHost,
		config.DBPort, config.DBName,
	)

	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		fmt.Println(" ******** ", err.Error())
		panic("Failed to connect to Mysql: " + dsn)
	}

	return db
}
