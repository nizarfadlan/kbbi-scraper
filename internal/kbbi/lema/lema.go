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
package lema

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"kbbi-scraper/internal/common"
	"kbbi-scraper/internal/database"
	"kbbi-scraper/internal/kbbi"

	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
)

const NORESULT_FILE = "no_result_word.json"

type NoResult struct {
	Word string `json:"word"`
	Url  string `json:"url"`
}

func ReadWordsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return words, nil
}

func saveToDatabase(db *sqlx.DB, results []kbbi.ResponseSearch, searchedWord string) error {
	var lemas []database.Lema
	for _, result := range results {
		for _, arti := range result.Arti {
			lemas = append(lemas, database.Lema{
				Kata:       searchedWord,
				Lema:       result.Lema,
				KelasKata:  arti.KelasKata,
				Keterangan: arti.Keterangan,
			})
		}
	}
	return database.InsertLemas(db, lemas)
}

func ProcessBatch(words []string, batchSize int, concurrency int, db *sqlx.DB, optionProxy string, providerProxy string) {
	total := len(words)
	processed := 0
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		batch := words[i:end]
		wg.Add(1)
		go func(batch []string) {
			defer wg.Done()
			for _, word := range batch {
				semaphore <- struct{}{}
				go func(word string) {
					defer func() { <-semaphore }()

					err := processWord(word, db, optionProxy, providerProxy)
					if err != nil {
						common.PrintError("Error processing word '%s': %v", word, err)
					}
				}(word)
			}
		}(batch)

		processed += len(batch)

		common.PrintCustom("[PROGRESS] %d/%d words processed", color.FgCyan, true, processed, total)
	}

	wg.Wait()
}

func processWord(word string, db *sqlx.DB, optionProxy string, providerProxy string) error {
	checkExist, errCheck := database.ExistsLemaByKata(db, word)
	if errCheck != nil {
		return fmt.Errorf("error checking in the database: %w", errCheck)
	}

	if checkExist {
		common.PrintInfo("word '%s' data in the database already exists", word)
		return nil
	}

	if checkWordOnNoResults(word) {
		common.PrintWarning("The word '%s' is in the list of files with no results", word)
		return nil
	}

	common.PrintInfo("Processing '%s'", word)
	results, err := kbbi.SearchWord(word, optionProxy, providerProxy)
	if err != nil {
		message := fmt.Sprintf("Error searching for '%s'\n", word)
		common.LogError(message, err)
		return fmt.Errorf("searching for '%s': %w", word, err)
	}

	if len(results) == 0 {
		url := fmt.Append([]byte(kbbi.KBBI_URL), word)
		message := fmt.Sprintf("[NO RESULT] No results found for '%s': %s\n", word, url)
		common.PrintError(message)
		common.LogInfo(message)
		addNoResult(NoResult{
			Word: word,
			Url:  string(url),
		})
		return nil
	}

	errInsert := saveToDatabase(db, results, word)
	if errInsert != nil {
		message := fmt.Sprintf("error inserting '%s'\n", word)
		common.LogError(message, errInsert)
		return fmt.Errorf("error inserting '%s': %w", word, errInsert)
	}

	common.PrintCustom("========================================", color.FgGreen, true)
	common.PrintSuccess("Successfully processed word '%s'", word)
	for iLema, result := range results {
		common.PrintCustom("Lema: %s", color.FgMagenta, true, result.Lema)
		for iArti, arti := range result.Arti {
			fmt.Printf("  Arti %d\n", iArti+1)
			common.PrintCustom("  Kelas Kata: %s", color.FgMagenta, true, arti.KelasKata)
			common.PrintCustom("  Keterangan: %s", color.FgMagenta, true, arti.Keterangan)
			if iArti < len(result.Arti)-1 {
				common.PrintCustom("  ========================================", color.FgMagenta, true)
			}
		}

		if iLema < len(results)-1 {
			common.PrintCustom("========================================", color.FgYellow, true)
		}
	}
	common.PrintCustom("========================================", color.FgGreen, true)

	return nil
}

func saveNoResults(results []NoResult) {
	data, err := json.Marshal(results)
	if err != nil {
		log.Printf("Error marshaling noresults: %v", err)
		return
	}

	err = os.WriteFile(NORESULT_FILE, data, 0644)
	if err != nil {
		log.Printf("Error saving noresults: %v", err)
	}
}

func loadNoResults() []NoResult {
	data, err := os.ReadFile(NORESULT_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			return []NoResult{}
		}
		log.Printf("Error reading noresults file: %v", err)
		return []NoResult{}
	}

	var results []NoResult
	err = json.Unmarshal(data, &results)
	if err != nil {
		log.Printf("Error unmarshaling noresults: %v", err)
		return []NoResult{}
	}

	return results
}

func checkWordOnNoResults(word string) bool {
	results := loadNoResults()
	for _, result := range results {
		if result.Word == word {
			return true
		}
	}
	return false
}

func addNoResult(newResult NoResult) {
	results := loadNoResults()
	results = append(results, newResult)
	saveNoResults(results)
}
