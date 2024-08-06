package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
)

const KBBI_URL = "https://kbbi.kemdikbud.go.id/entri/"

type Arti struct {
	KelasKata  string `json:"kelas_kata"`
	Keterangan string `json:"keterangan"`
}

type ResponseObj struct {
	Lema string `json:"lema"`
	Arti []Arti `json:"arti"`
}

type FakeBrowserHeadersResponse struct {
	Result []map[string]string `json:"result"`
}

func RandomHeader(headersList []map[string]string) map[string]string {
	randomIndex := rand.Intn(len(headersList))
	return headersList[randomIndex]
}

func GetHeadersList() []map[string]string {
	scrapeopsAPIKey := os.Getenv("SCRAPE_OPS")
	scrapeopsAPIEndpoint := "http://headers.scrapeops.io/v1/browser-headers?api_key=" + scrapeopsAPIKey

	req, _ := http.NewRequest("GET", scrapeopsAPIEndpoint, nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()

		var fakeBrowserHeadersResponse FakeBrowserHeadersResponse
		json.NewDecoder(resp.Body).Decode(&fakeBrowserHeadersResponse)
		return fakeBrowserHeadersResponse.Result
	}

	var emptySlice []map[string]string
	return emptySlice
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
	// var htmlTagRegex = regexp.MustCompile("<[^>]*>")
	// htmlContent, _ := s.Html()
	// cleanHtml := htmlTagRegex.ReplaceAllString(htmlContent, "")
	// decodedHtml := html.UnescapeString(cleanHtml)
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

func SearchWord(word string, optionProxy *string) ([]ResponseObj, error) {
	var dataResponse []ResponseObj
	var globalErr error

	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(1),
		colly.AllowURLRevisit(),
	)

	keyScrape := os.Getenv("SCRAPE_OPS")
	var headersList []map[string]string

	if keyScrape != "" {
		headersList = GetHeadersList()
	}

	c.OnRequest(func(r *colly.Request) {
		if keyScrape != "" {
			randomHeader := RandomHeader(headersList)
			for key, value := range randomHeader {
				r.Headers.Set(key, value)
			}
		}
	})

	c.SetRequestTimeout(time.Second * 40)

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

			responseObj := ResponseObj{
				Lema: extractKataDasar(h2),
				Arti: parseArti(list),
			}
			dataResponse = append(dataResponse, responseObj)
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		globalErr = err
	})

	urlKbbi := fmt.Sprintf("%s%s", KBBI_URL, word)
	if optionProxy != nil {
		typeProxy := *optionProxy
		if typeProxy == "residential" {
			proxyList := []string{
				fmt.Sprintf("http://scrapeops:%s@residential-proxy.scrapeops.io:8181", keyScrape),
			}

			rp, err := proxy.RoundRobinProxySwitcher(proxyList...)
			if err != nil {
				log.Fatal(err)
			}

			c.SetProxyFunc(rp)
		} else if typeProxy == "endpoint" {
			u, errUrlParse := url.Parse("https://proxy.scrapeops.io/v1/")
			if errUrlParse != nil {
				return nil, fmt.Errorf("failed to parse url: %w", errUrlParse)
			}

			q := u.Query()
			q.Set("api_key", keyScrape)
			q.Set("url", urlKbbi)
			u.RawQuery = q.Encode()

			urlKbbi = u.String()
		}
	}

	err := c.Visit(urlKbbi)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	c.Wait()

	if globalErr != nil {
		if globalErr.Error() == "limit reached" {
			PrintError("your search has reached the maximum limit in a day")
		}

		return nil, fmt.Errorf("something went wrong: %w", globalErr)
	}

	if len(dataResponse) == 0 {
		return []ResponseObj{}, nil
	}

	return dataResponse, nil
}
