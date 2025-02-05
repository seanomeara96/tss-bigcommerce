package main

import (
	"log"
	"os"
	"tss-bigcommerce/internal"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("[ERROR] loading .env file: %v", err)
	}

	caterHireConfig := internal.GenerateFilesConfig{}
	caterHireConfig.JobType = internal.CaterHireJobType
	caterHireConfig.StoreHash = os.Getenv("CH_STORE_HASH")
	caterHireConfig.AuthToken = os.Getenv("CH_XAUTHTOKEN")
	if caterHireConfig.StoreHash == "" || caterHireConfig.AuthToken == "" {
		log.Fatal("[ERROR] missing environment variables CH_STORE_HASH or CH_XAUTHTOKEN")
	}

	if err := internal.GenerateFiles(caterHireConfig); err != nil {
		log.Fatal("failed to run GenerateFiles")
	}

	hireAllConfig := internal.GenerateFilesConfig{}
	hireAllConfig.JobType = internal.CaterHireJobType
	hireAllConfig.StoreHash = os.Getenv("HA_STORE_HASH")
	hireAllConfig.AuthToken = os.Getenv("HA_XAUTHTOKEN")
	if hireAllConfig.StoreHash == "" || hireAllConfig.AuthToken == "" {
		log.Fatal("[ERROR] missing environment variables HA_STORE_HASH or HA_XAUTHTOKEN")
	}

	if err := internal.GenerateFiles(hireAllConfig); err != nil {
		log.Fatal("failed to run GenerateFiles")
	}

}
