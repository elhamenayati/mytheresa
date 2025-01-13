package server

import (
	"log"
	"reflect"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     int    `mapstructure:"DB_PORT"`
	DBName     string `mapstructure:"DB_NAME"`
}

func loadConfig() (Config, error) {
	var config Config

	///load env variables from env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it.")
	}

	viper.AutomaticEnv()
	confType := reflect.TypeOf(config)
	for i := 0; i < confType.NumField(); i++ {
		field := confType.Field(i)
		tag := field.Tag.Get("mapstructure")
		viper.MustBindEnv(tag)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return config, err
	}

	return config, nil
}
