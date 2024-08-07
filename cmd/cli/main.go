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
			searchWordlist(db, "local")
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

	batchSize := 100
	concurrency := 10
	optionProxy := "residential"

	start := time.Now()
	lema.ProcessBatch(words, batchSize, concurrency, db, &optionProxy)
	duration := time.Since(start)

	common.PrintInfo("Total execution time: %v", duration)
}

func getWordlistContent(db *sqlx.DB) {
	username := common.GetInput("Enter your KBBI username: ")
	password := common.GetInput("Enter your KBBI password: ")

	err := kata.GetWordList(db, username, password)
	if err != nil {
		common.PrintError("Error getting wordlist: %v", err)
		return
	}
}
