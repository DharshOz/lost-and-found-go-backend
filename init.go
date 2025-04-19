// init.go
package main

import (
	"log"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ Couldn't load .env file in init.go")
	}
}
