package main

import (
	"fmt"
	"time"

	"kbbi-scraper/internal/common"
	"kbbi-scraper/internal/database"
	"kbbi-scraper/internal/kbbi/kata"
	"kbbi-scraper/internal/kbbi/lema"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		common.PrintError("Error loading .env file")
		return
	}

	db, err := database.ConnectDB()
	if err != nil {
		common.PrintError("Error connecting to database: %v", err)
		return
	}
	defer database.CloseDB(db)

	for {
		common.DisplayMenu()
		choice := common.GetUserChoice()

		switch choice {
		case "1":
			getWordlistContent(db)
		case "2":
			common.PrintInfo("Default wordlist source is local file")
			typeWordList := common.GetInput("Choose wordlist source (local/db): ")
			if typeWordList != "db" && typeWordList != "local" {
				typeWordList = "local"
			} else {
				common.PrintError("Invalid input")
				return
			}

			searchWordlist(db, typeWordList)
		case "3":
			common.PrintInfo("Thank you for using this program. See you soon!")
			return
		default:
			common.PrintError("Choice is invalid. Please try again.")
		}

		fmt.Println()
	}
}

func searchWordlist(db *sqlx.DB, typeWordList string) {
	var words []string
	if typeWordList == "local" {
		filename := "word.txt"
		wordsFile, err := lema.ReadWordsFromFile(filename)
		if err != nil {
			common.PrintError("Error getting words from database: %v", err)
			return
		}

		words = wordsFile
	} else if typeWordList == "db" {
		wordsDB, err := database.GetWords(db)
		if err != nil {
			common.PrintError("Error getting words from database: %v", err)
			return
		}

		wordsLocal := make([]string, len(wordsDB))
		for i, word := range wordsDB {
			wordsLocal[i] = word.Kata
		}
		words = wordsLocal
	} else {
		common.PrintError("Invalid type word list")
		return
	}

	common.PrintInfo("Read %d words from file", len(words))

	withProxy := common.GetInput("Do you want to use proxy? (y/n): ")
	var optionProxy string
	if withProxy == "n" {
		optionProxy = ""
	} else if withProxy != "y" {
		common.PrintInfo("Default proxy is residential")
		chooseProxy := common.GetInput("Choose proxy (residential/datacenter): ")
		if chooseProxy == "" {
			chooseProxy = "residential"
		} else if chooseProxy != "residential" && chooseProxy != "datacenter" {
			common.PrintError("Invalid proxy type")
			return
		}
		optionProxy = chooseProxy
	} else {
		common.PrintError("Invalid input")
		return
	}

	batchSize := 100
	concurrency := 10

	start := time.Now()
	lema.ProcessBatch(words, batchSize, concurrency, db, &optionProxy)
	duration := time.Since(start)

	common.PrintInfo("Total execution time: %v", duration)
}

func getWordlistContent(db *sqlx.DB) {
	email := common.GetInput("Enter your KBBI email: ")
	password := common.GetInput("Enter your KBBI password: ")

	concurrency := 10

	start := time.Now()
	err := kata.GetWordList(db, email, password, concurrency)
	if err != nil {
		common.PrintError("Error getting wordlist: %v", err)
		return
	}
	duration := time.Since(start)

	common.PrintInfo("Total execution time: %v", duration)
}
