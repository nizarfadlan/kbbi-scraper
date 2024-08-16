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
package common

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

type FakeBrowserHeadersResponse struct {
	Result []map[string]string `json:"result"`
}

const SCRAPEOPS_FAKE_BROWSER_ENDPOINT = "http://headers.scrapeops.io/v1/browser-headers?api_key="

var proxyEndpoints = map[string]string{
	"scrapeops":   "https://proxy.scrapeops.io/v1/",
	"scrapingant": "https://api.scrapingant.com/v2/general",
	"scraperapi":  "http://api.scraperapi.com",
	"scrapingbee": "https://app.scrapingbee.com/api/v1/",
}

func RandomHeader(headersList []map[string]string) map[string]string {
	if len(headersList) == 0 {
		return map[string]string{}
	}

	randomIndex := rand.Intn(len(headersList))
	return headersList[randomIndex]
}

func GetHeadersList() []map[string]string {
	scrapeopsAPIKey := os.Getenv("SCRAPE_OPS")
	scrapeopsAPIEndpoint := fmt.Sprintf("%s%s", SCRAPEOPS_FAKE_BROWSER_ENDPOINT, scrapeopsAPIKey)

	req, _ := http.NewRequest("GET", scrapeopsAPIEndpoint, nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req.Close = true

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()

		var fakeBrowserHeadersResponse FakeBrowserHeadersResponse
		json.NewDecoder(resp.Body).Decode(&fakeBrowserHeadersResponse)
		return fakeBrowserHeadersResponse.Result
	}

	var emptySlice []map[string]string
	return emptySlice
}

func GetProxyResidential() []string {
	scrapeopsAPIKey := os.Getenv("SCRAPE_OPS")
	proxyList := []string{
		fmt.Sprintf("http://scrapeops.country=jp:%s@residential-proxy.scrapeops.io:8181", scrapeopsAPIKey),
	}

	return proxyList
}

func GetProxyDataCenter(urlKbbi, providerProxy string) (string, error) {
	endpoint, ok := proxyEndpoints[providerProxy]
	if !ok {
		return "", fmt.Errorf("unknown provider: %s", providerProxy)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}
	q := u.Query()
	q.Set("url", urlKbbi)

	switch providerProxy {
	case "scrapeops":
		q.Set("country", "jp")
		q.Set("api_key", os.Getenv("SCRAPE_OPS"))
	case "scrapingant":
		q.Set("browser", "false")
		q.Set("proxy_country", "ID")
		q.Set("x-api-key", os.Getenv("SCRAPING_ANT"))
	case "scraperapi":
		q.Set("api_key", os.Getenv("SCRAPER_API"))
	case "scrapingbee":
		q.Set("render_js", "false")
		q.Set("api_key", os.Getenv("SCRAPING_BEE"))
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}
