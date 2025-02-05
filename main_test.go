package main

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/seanomeara96/go-bigcommerce"
)

func TestGetOrder(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("[ERROR] loading .env file: %v", err)
	}

	storeHash := os.Getenv("CH_STORE_HASH")
	authToken := os.Getenv("CH_XAUTHTOKEN")
	if storeHash == "" || authToken == "" {
		t.Fatalf("[ERROR] missing environment variables CH_STORE_HASH or CH_XAUTHTOKEN")
	}

	client := bigcommerce.NewClient(storeHash, authToken, nil, log.Default())

	orderSortParams := bigcommerce.OrderSortQuery{
		Field:     bigcommerce.OrderSortFieldID,
		Direction: bigcommerce.OrderSortDirectionDesc,
	}

	orderQueryParams := bigcommerce.OrderQueryParams{
		Limit: 1,
		Sort:  orderSortParams.String(),
		MinID: 4077,
	}

	_, _, err := client.V2.GetOrders(orderQueryParams)
	if err != nil {
		t.Fatalf("[ERROR] getting orders: %v", err)
	}

}
