/*
 *  Copyright (c) 2024 Nizar Izzuddin Yatim Fadlan <hello@nizarfadlan.dev>
 * All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */
package cmd

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

func Execute() {
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

	// if !common.CheckSessionExists() {
	// 	email := common.GetInput("Enter your KBBI email: ")
	// 	password := common.GetInput("Enter your KBBI password: ")

	// 	if email == "" || password == "" {
	// 		common.PrintError("Email or password cannot be empty")
	// 		return
	// 	}

	// 	cookie, err := kbbi.LoginKBBI(email, password)
	// 	if err != nil {
	// 		common.PrintError("Error logging in to KBBI: %v", err)
	// 		return
	// 	}

	// 	if cookie != nil {
	// 		common.SaveSession(common.Session{
	// 			Email:    email,
	// 			Password: password,
	// 			Cookie:   *cookie,
	// 		})
	// 	} else {
	// 		common.PrintError("Error logging in to KBBI: cookie is nil")
	// 		return
	// 	}
	// }

	for {
		common.DisplayMenu()
		choice := common.GetUserChoice()

		switch choice {
		case "1":
			getWordlistContent(db)
			return
		case "2":
			common.PrintInfo("Default wordlist source is local file")
			typeWordList := common.GetInput("Choose wordlist source (local/db): ")
			if typeWordList != "db" && typeWordList != "local" {
				typeWordList = "local"
			}

			searchWordlist(db, typeWordList)
			return
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
		// Reverse string
		// sort.Sort(sort.Reverse(sort.StringSlice(words)))
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
	} else if withProxy == "y" {
		common.PrintInfo("Default proxy is residential")
		chooseProxy := common.GetInput("Choose proxy (residential/datacenter): ")
		if chooseProxy != "residential" && chooseProxy != "datacenter" {
			chooseProxy = "residential"
		}
		optionProxy = chooseProxy
	} else {
		common.PrintError("Invalid input")
		return
	}

	providerProxy := "scrapingant"
	if optionProxy == "datacenter" {
		common.PrintInfo("Default provider proxy is scrapingant")
		chooseProviderProxy := common.GetInput("Choose provider proxy (scrapeops/scrapingant/scraperapi/scrapingbee): ")
		if chooseProviderProxy != "scrapeops" && chooseProviderProxy != "scrapingant" && chooseProviderProxy != "scraperapi" && chooseProviderProxy != "scrapingbee" {
			chooseProviderProxy = "scrapingant"
		}
		providerProxy = chooseProviderProxy
	}

	batchSize := 100
	concurrency := 10
	if optionProxy == "datacenter" {
		if providerProxy == "scrapeops" || providerProxy == "scrapingant" {
			concurrency = 1
			batchSize = 10
		} else if providerProxy == "scraperapi" || providerProxy == "scrapingbee" {
			concurrency = 5
			batchSize = 50
		}
	}

	start := time.Now()
	lema.ProcessBatch(words, batchSize, concurrency, db, optionProxy, providerProxy)
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
