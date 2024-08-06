package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func LogError(message string, err error) {
	logFile, openErr := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		log.Printf("Failed to open log file: %v", openErr)
		return
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("%s: %v", message, err)
}

func readWordsFromFile(filename string) ([]string, error) {
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

func saveToDatabase(db *sqlx.DB, results []ResponseObj, searchedWord string) error {
	var lemas []Lema
	for _, result := range results {
		for _, arti := range result.Arti {
			lemas = append(lemas, Lema{
				Kata:       searchedWord,
				Lema:       result.Lema,
				KelasKata:  arti.KelasKata,
				Keterangan: arti.Keterangan,
			})
		}
	}
	return InsertLemas(db, lemas)
}

func processBatch(words []string, batchSize int, concurrency int, db *sqlx.DB, optionProxy *string) {
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

					checkExist, errCheck := ExistsLemaByKata(db, word)
					if errCheck != nil {
						PrintError("An error occurred when checking in the database\n")
						return
					}

					if checkExist {
						PrintInfo("Word '%s' data in the database already exists\n", word)
						return
					}

					PrintInfo("Processing '%s'", word)
					results, err := SearchWord(word, optionProxy)
					if err != nil {
						message := fmt.Sprintf("Error searching for '%s'\n", word)
						PrintError("%s: %v", message, err)
						LogError(message, err)
						return
					}

					if len(results) == 0 {
						message := fmt.Sprintf("No results found for '%s'\n", word)
						PrintInfo(message)
						LogError(message, nil)
						return
					}

					errInsert := saveToDatabase(db, results, word)
					if errInsert != nil {
						message := fmt.Sprintf("Error inserting '%s'\n", word)
						PrintError("%s: %v", message, errInsert)
						LogError(message, errInsert)
						return
					}

					PrintSuccess("Successfully processed word '%s'", word)
					for _, result := range results {
						PrintSuccess("Lema: %s\n", result.Lema)
						for _, arti := range result.Arti {
							PrintSuccess("  Kelas Kata: %s\n", arti.KelasKata)
							PrintSuccess("  Keterangan: %s\n", arti.Keterangan)
						}
						fmt.Println()
					}
				}(word)
			}
		}(batch)

		processed += len(batch)

		PrintCustom("[PROGRESS] %d/%d words processed", color.FgCyan, true, processed, total)
	}

	wg.Wait()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		PrintError("Error loading .env file")
		return
	}

	filename := "word.txt"
	words, err := readWordsFromFile(filename)
	if err != nil {
		PrintError(err.Error())
		return
	}

	PrintInfo("Read %d words from file", len(words))

	db, err := ConnectDB()
	if err != nil {
		PrintError("Error connecting to database: %v", err)
		return
	}
	defer CloseDB(db)

	batchSize := 100
	concurrency := 10
	optionProxy := "residential"

	start := time.Now()
	processBatch(words, batchSize, concurrency, db, &optionProxy)
	duration := time.Since(start)

	PrintInfo("Total execution time: %v", duration)
}
