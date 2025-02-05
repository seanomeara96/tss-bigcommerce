package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("[ERROR] loading .env file: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`Hello world!`))
	})

	if err := http.ListenAndServe(":3000", nil); err != nil {
		panic(err)
	}
}
