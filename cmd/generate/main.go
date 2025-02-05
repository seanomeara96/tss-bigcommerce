package main

import (
	"log"
	"tss-bigcommerce/internal"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("[ERROR] loading .env file: %v", err)
	}

	if err := internal.GenerateFiles(); err != nil {
		log.Fatal("failed to run GenerateFiles")
	}
}
