package kata

import (
	"kbbi-scraper/internal/common"
	"kbbi-scraper/internal/kbbi"

	"github.com/gocolly/colly/v2"
	"github.com/jmoiron/sqlx"
)

func GetWordList(db *sqlx.DB, email string, password string) error {
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	err := kbbi.LoginKBBI(c, email, password)
	if err != nil {
		return err
	}

	progress := common.LoadProgress()

	startLetter := 'A'
	if progress.CurrentLetter != "" {
		startLetter = rune(progress.CurrentLetter[0])
	}

	for letter := startLetter; letter <= 'Z'; letter++ {
		startPage := 1
		if string(letter) == progress.CurrentLetter {
			startPage = progress.CurrentPage
		}

		err = kbbi.GetWordListByAlphabet(db, c, string(letter), startPage)
		if err != nil {
			common.PrintError("Error scraping words for letter %c: %v", letter, err)
			common.SaveProgress(common.Progress{
				CurrentLetter: string(letter),
				CurrentPage:   startPage,
			})
			continue
		}
	}

	return nil
}
