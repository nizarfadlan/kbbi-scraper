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
package kata

import (
	"fmt"
	"kbbi-scraper/internal/common"
	"kbbi-scraper/internal/kbbi"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/jmoiron/sqlx"
)

func GetWordList(db *sqlx.DB, email string, password string, concurrency int) error {
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	_, err := kbbi.LoginKBBI(email, password)
	if err != nil {
		return err
	}

	progress := common.LoadProgress()
	startLetter := 'A'

	if progress.CurrentLetter != "" {
		startLetter = rune(progress.CurrentLetter[0])
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)
	errChan := make(chan error, 26)

	for letter := startLetter; letter <= 'Z'; letter++ {
		wg.Add(1)
		go func(letter rune) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			startPage := 1
			if string(letter) == progress.CurrentLetter {
				startPage = progress.CurrentPage
			}

			err := processLetter(c.Clone(), db, letter, startPage)
			if err != nil {
				errChan <- fmt.Errorf("error processing letter %c: %v", letter, err)
			}
		}(letter)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			common.PrintError("%v", err)
		}
	}

	return nil
}

func processLetter(c *colly.Collector, db *sqlx.DB, letter rune, startPage int) error {
	currentPage := startPage

	for {
		isLastPage, err := kbbi.GetWordListByAlphabet(c, db, string(letter), currentPage)
		if err != nil {
			common.SaveProgress(common.Progress{
				CurrentLetter: string(letter),
				CurrentPage:   currentPage,
			})
			return err
		}

		if isLastPage {
			break
		}

		currentPage++
	}

	return nil
}
