package parser

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	browser "github.com/EDDYCJY/fake-useragent"

	v2 "github.com/Seann-Moser/WebParser/v2"
)

type SiteSource struct {
	client   *http.Client
	ProxyUrl string
	MaxDelay int `json:"max_delay"`
	MinDelay int `json:"min_delay"`
}

func NewSiteSource(proxyUrl string, minDelay, maxDelay int) *SiteSource {
	return &SiteSource{
		client:   &http.Client{Timeout: 5 * time.Second},
		ProxyUrl: proxyUrl,
		MaxDelay: minDelay,
		MinDelay: maxDelay,
	}
}

func (s *SiteSource) GetSourceCode(u string) (*v2.HtmlData, error) {
	_, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	pageSource, err := s.getRawBody(u)
	if err != nil && s.ProxyUrl != "" {
		pageSource, err = s.getProxyData(u)
	} else if err != nil {
		return nil, err
	}
	return pageSource, err
}

func (s *SiteSource) getRawBody(searchURL string) (*v2.HtmlData, error) {
	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	random := browser.Random()
	req.Header.Set("User-Agent", random)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	pageSource, err := v2.NewHTMLSourceRequest().ProcessSourceCode(string(b))
	s.wait()
	return pageSource, nil
}

func (s *SiteSource) getProxyData(searchURL string) (*v2.HtmlData, error) {
	v := url.Values{}
	v.Set("u", searchURL)
	println("hitting proxy:" + searchURL)
	proxyURL, err := url.Parse(s.ProxyUrl)
	if err != nil {
		return nil, err
	}

	proxyURL.RawQuery = v.Encode()
	req, err := http.NewRequest(http.MethodGet, proxyURL.String(), nil)
	if err != nil {
		return nil, err
	}
	random := browser.Random()
	req.Header.Set("User-Agent", random)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	httpRequestHandler := v2.NewHTMLSourceRequest()
	return httpRequestHandler.ProcessSourceCode(string(b))
}

func (s *SiteSource) wait() {
	if s.MaxDelay <= 0 {
		return
	}
	rand.Seed(time.Now().UnixNano())
	if s.MaxDelay-s.MinDelay == 0 {
		time.Sleep(time.Duration(s.MaxDelay) * time.Second)
		return
	}
	number1 := rand.Intn(int(math.Abs(float64(s.MaxDelay - s.MinDelay))))
	number1 += s.MinDelay + 1
	time.Sleep(time.Duration(number1) * time.Second)
}
