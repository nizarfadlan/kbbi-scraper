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
package kbbi

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
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
	KBBI_BANNED_URL   = "https://kbbi.kemdikbud.go.id/Account/Banned"
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

type LoginResult struct {
	Cookie   string
	IsBanned bool
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
	s.Find("small").Remove()
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

func parsePrakategorial(e *goquery.Selection) (string, string) {
	var kelasKata, contoh string

	processPrakategorial := func(s *goquery.Selection) {
		s.Contents().Each(func(_ int, node *goquery.Selection) {
			if node.Is("i") && strings.Contains(node.Text(), "prakategorial") {
				kelasKata = "prakategorial[kata tidak dipakai dalam bentuk dasarnya]"
			} else if node.Is("font") || (!node.Is("a") && !node.Is("span") && node.Text() != "") {
				text := strings.TrimSpace(node.Text())
				if text != "" && contoh == "" {
					contoh = text
				}
			}
		})
	}

	processPrakategorial(e)

	if kelasKata == "" {
		e.Siblings().Each(func(_ int, sibling *goquery.Selection) {
			processPrakategorial(sibling)
			if kelasKata != "" {
				return
			}
		})
	}

	return kelasKata, contoh
}

func parseArtiTypePrakategorial(e *goquery.Selection) []Arti {
	kelasKata, contoh := parsePrakategorial(e)

	if kelasKata == "" {
		return nil
	}

	return []Arti{{
		KelasKata:  kelasKata,
		Keterangan: contoh,
	}}
}

func extractKataTidakBaku(e *goquery.Selection) string {
	kataTidakBaku := e.Find("small")
	if kataTidakBaku.Length() > 0 {
		kataTidakBaku.Find("sup").Remove()
		text := strings.TrimSpace(kataTidakBaku.Text())
		parts := strings.Split(text, ":")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

func LoginKBBI(email, password string) (*LoginResult, error) {
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	loginResult := &LoginResult{}
	var token string

	c.OnHTML("form input[name=__RequestVerificationToken]", func(e *colly.HTMLElement) {
		token = e.Attr("value")
	})

	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != http.StatusOK {
			return
		}

		for _, cookie := range c.Cookies(r.Request.URL.String()) {
			if cookie.Name == ".AspNet.ApplicationCookie" {
				loginResult.Cookie = cookie.Value
				break
			}
		}

		loginResult.IsBanned = r.Request.URL.String() == KBBI_BANNED_URL
	})

	if err := c.Visit(KBBI_LOGIN_URL); err != nil {
		return nil, fmt.Errorf("failed to visit login page: %w", err)
	}

	if token == "" {
		return nil, fmt.Errorf("could not find CSRF token")
	}

	err := c.Post(KBBI_LOGIN_URL, map[string]string{
		"__RequestVerificationToken": token,
		"Posel":                      email,
		"KataSandi":                  password,
		"IngatSaya":                  "false",
	})

	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}

	if loginResult.Cookie == "" {
		return nil, fmt.Errorf("could not find login cookie")
	}

	if loginResult.IsBanned {
		return nil, fmt.Errorf("account is banned")
	}

	return loginResult, nil
}

func GetWordListByAlphabet(c *colly.Collector, db *sqlx.DB, letter string, startPage int) (bool, error) {
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

	fmt.Println(baseURL.String())

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

func SearchWord(word string, optionProxy string, providerProxy string) ([]ResponseSearch, error) {
	var dataResponse []ResponseSearch
	var globalErr error
	maxRetries := 3
	retryDelay := time.Second * 5

	for retry := 0; retry < maxRetries; retry++ {
		c := colly.NewCollector(
			colly.Async(true),
			colly.MaxDepth(2),
			colly.AllowURLRevisit(),
		)

		setHeaders(c)
		c.SetRequestTimeout(time.Second * 60)
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Parallelism: 10,
			RandomDelay: 5 * time.Second,
		})

		c.WithTransport(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		})

		c.OnHTML(".body-content", func(e *colly.HTMLElement) {
			e.DOM.Find("h4:contains('Pesan')").NextAll().Remove()
			e.DOM.Find("form#searchForm").PrevAll().Remove()
			e.DOM.Find("h4:contains('Pesan')").Remove()
			e.DOM.Find("form#searchForm").Remove()

			if checkBatasHarian(e.DOM) {
				globalErr = fmt.Errorf("limit reached")
				return
			}
			if checkFrasaNotFound(e.DOM) {
				return
			}

			e.DOM.Find("h2").Each(func(_ int, h2 *goquery.Selection) {
				lemma := extractKataDasar(h2)

				prakategorial := parseArtiTypePrakategorial(h2)
				if len(prakategorial) > 0 {
					bentukTidakBaku := extractKataTidakBaku(h2)
					responseObj := ResponseSearch{
						Lema: lemma,
						Arti: prakategorial,
					}

					common.LogInfo(fmt.Sprintf("Response Search Prakategorial '%s': \nKata tidak baku: %s\nresponse: %s", word, bentukTidakBaku, responseObj))
					// dataResponse = append(dataResponse, responseObj)
				}

				list := h2.NextUntil("h2").Filter("ul, ol")
				if list.Length() > 0 {
					responseObj := ResponseSearch{
						Lema: lemma,
						Arti: parseArti(list),
					}
					dataResponse = append(dataResponse, responseObj)
				}
			})
		})

		c.OnError(func(r *colly.Response, err error) {
			globalErr = err
			if r.StatusCode >= 400 && r.StatusCode < 600 {
				var errorResponse map[string]interface{}
				jsonErr := json.Unmarshal(r.Body, &errorResponse)
				if jsonErr == nil {
					if detail, ok := errorResponse["detail"]; ok {
						globalErr = fmt.Errorf("%s: %s", r.Request.URL, detail)
						return
					}
				}
				globalErr = fmt.Errorf("%s: %s", r.Request.URL, r.Body)
			}
		})

		c.OnResponse(func(r *colly.Response) {
			if r.StatusCode == 200 && len(r.Body) == 0 {
				fmt.Printf("Received %s empty body with 200 status code", word)
			}
		})

		urlKbbi, errProxy := setProxy(c, word, optionProxy, providerProxy)
		if errProxy != nil {
			return nil, fmt.Errorf("\nfailed to set proxy: %w", errProxy)
		}

		err := c.Visit(urlKbbi)
		if err != nil {
			if isEOF(err) {
				common.PrintError(fmt.Sprintf("EOF error occurred. Retrying in %v... (Attempt %d/%d)", retryDelay, retry+1, maxRetries))
				time.Sleep(retryDelay)
				continue
			}

			return nil, fmt.Errorf("\nsearch failed: %w", err)
		}

		c.Wait()

		if globalErr != nil {
			if globalErr.Error() == "limit reached" {
				common.PrintError("your search has reached the maximum limit in a day")
				return nil, fmt.Errorf("limit reached")
			} else if isEOF(globalErr) {
				if retry == maxRetries-1 {
					return nil, fmt.Errorf("\nEOF error occurred: %w", globalErr)
				}
				common.PrintError(fmt.Sprintf("EOF error occurred. Retrying in %v... (Attempt %d/%d)", retryDelay, retry+1, maxRetries))
				time.Sleep(retryDelay)
				continue
			} else {
				return nil, fmt.Errorf("\nsomething went wrong: %w", globalErr)
			}
		} else {
			break
		}
	}

	if len(dataResponse) == 0 {
		return []ResponseSearch{}, nil
	}

	return dataResponse, nil
}

func isEOF(err error) bool {
	if err == nil {
		return false
	}

	errString := err.Error()
	if err == io.EOF ||
		strings.Contains(errString, "EOF") ||
		strings.Contains(errString, "connection reset by peer") ||
		strings.Contains(errString, "broken pipe") ||
		strings.Contains(errString, "use of closed network connection") {
		return true
	}

	if netErr, ok := err.(*net.OpError); ok {
		if netErr.Err != nil && strings.Contains(netErr.Err.Error(), "EOF") {
			return true
		}
	}

	return false
}

func setHeaders(c *colly.Collector) {
	scrapeopsAPIKey := os.Getenv("SCRAPE_OPS")
	var headersList []map[string]string

	if scrapeopsAPIKey != "" {
		headersList = common.GetHeadersList()
	}

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
		r.Headers.Set("DNT", "1")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		if scrapeopsAPIKey != "" {
			randomHeader := common.RandomHeader(headersList)
			for key, value := range randomHeader {
				r.Headers.Set(key, value)
			}
		}
	})
}

func setProxy(c *colly.Collector, word string, optionProxy string, providerProxy string) (string, error) {
	urlKbbi := fmt.Sprintf("%s%s", KBBI_URL, word)
	if optionProxy != "" {
		if optionProxy == "residential" {
			proxyResidential := common.GetProxyResidential()

			rp, err := proxy.RoundRobinProxySwitcher(proxyResidential...)
			if err != nil {
				return "", fmt.Errorf("\nfailed to create proxy switcher: %w", err)
			}

			c.SetProxyFunc(rp)
		} else if optionProxy == "datacenter" {
			pu, err := common.GetProxyDataCenter(urlKbbi, providerProxy)
			if err != nil {
				return "", fmt.Errorf("\nfailed to get proxy endpoint: %w", err)
			}

			urlKbbi = pu
		}
	}

	return urlKbbi, nil
}
