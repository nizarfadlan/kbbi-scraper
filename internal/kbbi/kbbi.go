package kbbi

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"kbbi-scraper/internal/common"
	"kbbi-scraper/internal/database"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
	"github.com/jmoiron/sqlx"
)

const (
	KBBI_URL          = "https://kbbi.kemdikbud.go.id/entri/"
	KBBI_LOGIN_URL    = "https://kbbi.kemdikbud.go.id/Account/Login"
	KBBI_WORDLIST_URL = "https://kbbi.kemdikbud.go.id/Cari/Alphabet"
)

type ResponseSearch struct {
	Lema string `json:"lema"`
	Arti []Arti `json:"arti"`
}

type Arti struct {
	KelasKata  string `json:"kelas_kata"`
	Keterangan string `json:"keterangan"`
}

func checkFrasaNotFound(e *goquery.Selection) bool {
	html, _ := e.Find("h4:contains('tidak ditemukan')").Html()
	return strings.Contains(html, "tidak ditemukan")
}

func checkBatasHarian(e *goquery.Selection) bool {
	html, _ := e.Find("h1:contains('Batas Sehari')").Html()
	return strings.Contains(html, "Batas Sehari")
}

func extractKataDasar(s *goquery.Selection) string {
	s.Find("sup").Remove()
	s.Find("span.rootword").Remove()
	return strings.TrimSpace(s.Text())
}

func parseKelasKata(s *goquery.Selection) string {
	var kelasKata []string
	s.Find("span").Each(func(_ int, span *goquery.Selection) {
		title, _ := span.Attr("title")
		kelasKata = append(kelasKata, fmt.Sprintf("%s[%s]", span.Text(), title))
	})
	return strings.TrimSpace(strings.Join(kelasKata, " "))
}

func parseKeterangan(s *goquery.Selection) string {
	s.Find("span").Remove()
	return strings.TrimSpace(strings.ReplaceAll(s.Text(), "\n", ""))
}

func parseArti(s *goquery.Selection) []Arti {
	var artiList []Arti
	s.Find("li").Each(func(_ int, li *goquery.Selection) {
		arti := Arti{
			KelasKata:  parseKelasKata(li),
			Keterangan: parseKeterangan(li),
		}
		artiList = append(artiList, arti)
	})
	return artiList
}

func LoginKBBI(c *colly.Collector, email, password string) error {
	var loginErr error
	var token string

	c.OnHTML("form input[name=__RequestVerificationToken]", func(e *colly.HTMLElement) {
		token = e.Attr("value")
	})

	err := c.Visit(KBBI_LOGIN_URL)
	if err != nil {
		return err
	}

	if token == "" {
		return fmt.Errorf("could not find CSRF token")
	}

	err = c.Post(KBBI_LOGIN_URL, map[string]string{
		"Posel":                      email,
		"KataSandi":                  password,
		"__RequestVerificationToken": token,
		"IngatSaya":                  "true",
	})

	if err != nil {
		return err
	}

	return loginErr
}

func GetWordListByAlphabet(db *sqlx.DB, c *colly.Collector, letter string, startPage int) (bool, error) {
	var globalErr error
	baseURL, err := url.Parse(KBBI_WORDLIST_URL)
	if err != nil {
		return false, err
	}

	params := url.Values{}
	params.Add("masukan", letter)
	params.Add("masukanLengkap", letter)
	params.Add("page", strconv.Itoa(startPage))

	baseURL.RawQuery = params.Encode()

	var totalPages int
	var words []string
	currentPage := startPage
	isLastPage := false

	c.OnHTML("#currentPageId", func(e *colly.HTMLElement) {
		parts := strings.Split(e.Text, "/")
		if len(parts) == 2 {
			totalPages, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			currentPage, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
			isLastPage = currentPage == totalPages
		}
	})

	c.OnHTML(".row .col-md-3", func(e *colly.HTMLElement) {
		wordHTML := e.DOM.Find("a")
		wordHTML.Find("sup").Remove()
		word := strings.TrimSpace(wordHTML.Text())
		words = append(words, word)
		fmt.Printf("Letter %s, Page %d: %s\n", letter, currentPage, word)
	})

	c.OnHTML(".row", func(e *colly.HTMLElement) {
		nextPage := e.ChildAttr("a[title='Ke halaman berikutnya']", "href")
		if nextPage != "" {
			currentPage++
			nextURL := e.Request.AbsoluteURL(nextPage)
			fmt.Printf("Moving to next page for letter %s: %s\n", letter, nextURL)
			common.SaveProgress(common.Progress{
				CurrentLetter: letter,
				CurrentPage:   currentPage,
			})
			e.Request.Visit(nextURL)
		} else {
			isLastPage = true
		}
	})

	err = c.Visit(baseURL.String())
	if err != nil {
		return false, err
	}

	c.OnScraped(func(r *colly.Response) {
		globalErr = database.InsertWords(db, words)
		fmt.Printf("Finished scraping words for letter %s. Total pages: %d\n", letter, totalPages)
	})

	if globalErr != nil {
		return isLastPage, globalErr
	}

	return isLastPage, nil
}

func SearchWord(word string, optionProxy *string) ([]ResponseSearch, error) {
	var dataResponse []ResponseSearch
	var globalErr error

	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(1),
		colly.AllowURLRevisit(),
	)

	setHeaders(c)
	c.SetRequestTimeout(time.Second * 60)

	c.OnHTML(".body-content", func(e *colly.HTMLElement) {
		e.DOM.Find("h4:contains('Pesan')").NextAll().Remove()
		if checkBatasHarian(e.DOM) {
			globalErr = fmt.Errorf("limit reached")
			return
		}
		if checkFrasaNotFound(e.DOM) {
			return
		}

		e.DOM.Find("h2").Each(func(_ int, h2 *goquery.Selection) {
			list := h2.NextUntil("h2").Filter("ul, ol")

			if list.Length() == 0 {
				return
			}

			responseObj := ResponseSearch{
				Lema: extractKataDasar(h2),
				Arti: parseArti(list),
			}
			dataResponse = append(dataResponse, responseObj)
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		globalErr = err
	})

	urlKbbi, errProxy := setProxy(c, word, optionProxy)
	if errProxy != nil {
		return nil, fmt.Errorf("failed to set proxy: %w", errProxy)
	}

	err := c.Visit(*urlKbbi)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	c.Wait()

	if globalErr != nil {
		if globalErr.Error() == "limit reached" {
			common.PrintError("your search has reached the maximum limit in a day")
		}

		return nil, fmt.Errorf("something went wrong: %w", globalErr)
	}

	if len(dataResponse) == 0 {
		return []ResponseSearch{}, nil
	}

	return dataResponse, nil
}

func setHeaders(c *colly.Collector) {
	scrapeopsAPIKey := os.Getenv("SCRAPE_OPS")
	var headersList []map[string]string

	if scrapeopsAPIKey != "" {
		headersList = common.GetHeadersList()
	}

	c.OnRequest(func(r *colly.Request) {
		if scrapeopsAPIKey != "" {
			randomHeader := common.RandomHeader(headersList)
			for key, value := range randomHeader {
				r.Headers.Set(key, value)
			}
		}
	})
}

func setProxy(c *colly.Collector, word string, optionProxy *string) (*string, error) {
	urlKbbi := fmt.Sprintf("%s%s", KBBI_URL, word)
	if optionProxy != nil {
		typeProxy := *optionProxy
		if typeProxy == "residential" {
			proxyList := common.GetProxyList()

			rp, err := proxy.RoundRobinProxySwitcher(proxyList...)
			if err != nil {
				return nil, fmt.Errorf("failed to create proxy switcher: %w", err)
			}

			c.SetProxyFunc(rp)
		} else if typeProxy == "datacenter" {
			pu, err := common.GetProxyDataCenter(urlKbbi)
			if err != nil {
				return nil, fmt.Errorf("failed to get proxy endpoint: %w", err)
			}

			urlKbbi = *pu
		}
	}

	return &urlKbbi, nil
}
