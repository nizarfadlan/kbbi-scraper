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

// func GetContentWords(
// 	words []string,
// 	email string,
// 	password string,
// 	batchSize int,
// 	concurrency int,
// 	db *sqlx.DB,
// 	optionProxy *string,
// ) error {
// 	c := colly.NewCollector(
// 		colly.AllowURLRevisit(),
// 		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
// 	)

// 	err := kbbi.LoginKBBI(c, email, password)
// 	if err != nil {
// 		return err
// 	}

// 	ProcessBatch(words, batchSize, concurrency, db, optionProxy)

// 	return nil
// }

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

					err := processWord(word, db, optionProxy)
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

func processWord(word string, db *sqlx.DB, optionProxy *string) error {
	checkExist, errCheck := database.ExistsLemaByKata(db, word)
	if errCheck != nil {
		return fmt.Errorf("error checking in the database: %w", errCheck)
	}

	if checkExist {
		common.PrintInfo("word '%s' data in the database already exists", word)
		return nil
	}

	common.PrintInfo("Processing '%s'", word)
	results, err := kbbi.SearchWord(word, optionProxy)
	if err != nil {
		message := fmt.Sprintf("Error searching for '%s'\n", word)
		common.LogError(message, err)
		return fmt.Errorf("searching for '%s': %w", word, err)
	}

	if len(results) == 0 {
		message := fmt.Sprintf("No results found for '%s': %s\n", word, fmt.Append([]byte(kbbi.KBBI_URL), word))
		common.PrintWarning(message)
		common.LogInfo(message)
		return nil
	}

	errInsert := saveToDatabase(db, results, word)
	if errInsert != nil {
		message := fmt.Sprintf("error inserting '%s'\n", word)
		common.LogError(message, errInsert)
		return fmt.Errorf("error inserting '%s': %w", word, errInsert)
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

	return nil
}
