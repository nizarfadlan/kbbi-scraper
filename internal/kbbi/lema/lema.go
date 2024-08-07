package lema

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"kbbi-scraper/internal/common"
	"kbbi-scraper/internal/database"
	"kbbi-scraper/internal/kbbi"

	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
)

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

func ProcessBatch(words []string, batchSize int, concurrency int, db *sqlx.DB, optionProxy *string) {
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

					checkExist, errCheck := database.ExistsLemaByKata(db, word)
					if errCheck != nil {
						common.PrintError("An error occurred when checking in the database\n")
						return
					}

					if checkExist {
						common.PrintInfo("Word '%s' data in the database already exists\n", word)
						return
					}

					common.PrintInfo("Processing '%s'", word)
					results, err := kbbi.SearchWord(word, optionProxy)
					if err != nil {
						message := fmt.Sprintf("Error searching for '%s'\n", word)
						common.PrintError("%s: %v", message, err)
						common.LogError(message, err)
						return
					}

					if len(results) == 0 {
						message := fmt.Sprintf("No results found for '%s': %s\n", word, fmt.Append([]byte(kbbi.KBBI_URL), word))
						common.PrintInfo(message)
						common.LogInfo(message)
						return
					}

					errInsert := saveToDatabase(db, results, word)
					if errInsert != nil {
						message := fmt.Sprintf("Error inserting '%s'\n", word)
						common.PrintError("%s: %v", message, errInsert)
						common.LogError(message, errInsert)
						return
					}

					common.PrintSuccess("Successfully processed word '%s'", word)
					for _, result := range results {
						common.PrintSuccess("Lema: %s\n", result.Lema)
						for _, arti := range result.Arti {
							common.PrintSuccess("  Kelas Kata: %s\n", arti.KelasKata)
							common.PrintSuccess("  Keterangan: %s\n", arti.Keterangan)
						}
						fmt.Println()
					}
				}(word)
			}
		}(batch)

		processed += len(batch)

		common.PrintCustom("[PROGRESS] %d/%d words processed", color.FgCyan, true, processed, total)
	}

	wg.Wait()
}
