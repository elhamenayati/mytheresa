package internal

import (
	"fmt"
	"os"
)

// /read file
func ReadData() (*os.File, error) {
	fileName := "data/products.json"

	object, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening object type:", err)
		return nil, err
	}

	return object, nil
}
