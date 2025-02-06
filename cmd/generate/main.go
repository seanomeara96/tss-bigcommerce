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

	db, err := internal.Database(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//TODO retrieve last recorded order id from both stores and add to the config

	caterHireConfig := internal.GenerateFilesConfig{}
	caterHireConfig.DB = db
	caterHireConfig.MinOrderID = 99999
	caterHireConfig.JobType = internal.CaterHireJobType
	caterHireConfig.StoreHash = os.Getenv("CH_STORE_HASH")
	caterHireConfig.AuthToken = os.Getenv("CH_XAUTHTOKEN")
	if caterHireConfig.StoreHash == "" || caterHireConfig.AuthToken == "" {
		log.Fatal("[ERROR] missing environment variables CH_STORE_HASH or CH_XAUTHTOKEN")
	}

	if err := internal.GenerateFiles(caterHireConfig); err != nil {
		log.Fatal("failed to run GenerateFiles for caterhire")
	}

	hireAllConfig := internal.GenerateFilesConfig{}
	hireAllConfig.DB = db
	hireAllConfig.MinOrderID = 99999
	hireAllConfig.JobType = internal.HireAlljobType
	hireAllConfig.StoreHash = os.Getenv("HA_STORE_HASH")
	hireAllConfig.AuthToken = os.Getenv("HA_XAUTHTOKEN")
	if hireAllConfig.StoreHash == "" || hireAllConfig.AuthToken == "" {
		log.Fatal("[ERROR] missing environment variables HA_STORE_HASH or HA_XAUTHTOKEN")
	}

	if err := internal.GenerateFiles(hireAllConfig); err != nil {
		log.Fatal("failed to run GenerateFiles for hireall")
	}

}
