package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"tss-bigcommerce/internal"

	"github.com/joho/godotenv"
)

func run() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("[ERROR] loading .env file: %v", err)
	}

	fileDestination := os.Getenv("FILE_PATH")
	if fileDestination == "" {
		return fmt.Errorf("file destination cannot be empty")
	}

	db, err := internal.Database(nil)
	if err != nil {
		return fmt.Errorf("error conneting to the database %w", err)
	}
	defer db.Close()

	stmt, err := db.Prepare(`SELECT order_id FROM orders WHERE website = ? LIMIT 1`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	//TODO retrieve last recorded order id from both stores and add to the config

	caterHireConfig := internal.GenerateFilesConfig{}
	if err := stmt.QueryRow(internal.CATERHIRE).Scan(&caterHireConfig.MinOrderID); err != nil {
		if err == sql.ErrNoRows {
			caterHireConfig.MinOrderID = 4126
		} else {
			return err
		}
	}
	fmt.Println("caterHireConfig.MinOrderID", caterHireConfig.MinOrderID)
	caterHireConfig.JobType = internal.CaterHireJobType
	caterHireConfig.StoreHash = os.Getenv("CH_STORE_HASH")
	caterHireConfig.AuthToken = os.Getenv("CH_XAUTHTOKEN")
	if caterHireConfig.StoreHash == "" || caterHireConfig.AuthToken == "" {
		return fmt.Errorf("[ERROR] missing environment variables CH_STORE_HASH or CH_XAUTHTOKEN")
	}
	if err := internal.GenerateFiles(db, fileDestination, caterHireConfig); err != nil {
		return fmt.Errorf("failed to run GenerateFiles for caterhire")
	}

	// TODO remove
	if true {
		return nil
	}

	hireAllConfig := internal.GenerateFilesConfig{}
	if err := stmt.QueryRow(internal.HIREALL).Scan(&hireAllConfig.MinOrderID); err != nil {
		return err
	}
	hireAllConfig.JobType = internal.HireAlljobType
	hireAllConfig.StoreHash = os.Getenv("HA_STORE_HASH")
	hireAllConfig.AuthToken = os.Getenv("HA_XAUTHTOKEN")
	if hireAllConfig.StoreHash == "" || hireAllConfig.AuthToken == "" {
		return fmt.Errorf("[ERROR] missing environment variables HA_STORE_HASH or HA_XAUTHTOKEN")
	}
	if err := internal.GenerateFiles(db, fileDestination, hireAllConfig); err != nil {
		return fmt.Errorf("failed to run GenerateFiles for hireall")
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Failed to run script %v", err)
	}
}
