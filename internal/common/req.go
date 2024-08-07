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

const (
	SCRAPEOPS_FAKE_BROWSER_ENDPOINT = "http://headers.scrapeops.io/v1/browser-headers?api_key="
	SCRAPEOPS_PROXY_ENDPOINT        = "https://proxy.scrapeops.io/v1/"
)

var scrapeopsAPIKey = os.Getenv("SCRAPE_OPS")

func RandomHeader(headersList []map[string]string) map[string]string {
	if len(headersList) == 0 {
		return map[string]string{}
	}

	randomIndex := rand.Intn(len(headersList))
	return headersList[randomIndex]
}

func GetHeadersList() []map[string]string {
	scrapeopsAPIEndpoint := SCRAPEOPS_FAKE_BROWSER_ENDPOINT + scrapeopsAPIKey

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

func GetProxyList() []string {
	proxyList := []string{
		fmt.Sprintf("http://scrapeops:%s@residential-proxy.scrapeops.io:8181", scrapeopsAPIKey),
	}

	return proxyList
}

func GetProxyDataCenter(urlKbbi string) (*string, error) {
	u, errUrlParse := url.Parse(SCRAPEOPS_PROXY_ENDPOINT)
	if errUrlParse != nil {
		return nil, fmt.Errorf("failed to parse url: %w", errUrlParse)
	}

	q := u.Query()
	q.Set("api_key", scrapeopsAPIKey)
	q.Set("url", urlKbbi)
	u.RawQuery = q.Encode()
	url := u.String()

	return &url, nil
}
