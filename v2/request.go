package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"

	"golang.org/x/net/html"
)

type HTMLSourceRequest struct {
	client       *http.Client
	tokenizer    *html.Tokenizer
	Cache        *cache.Cache
	SleepTimeMax int
}

// NewHTMLSourceRequest creates a new source request with a http client
func NewHTMLSourceRequest() *HTMLSourceRequest {
	return &HTMLSourceRequest{
		client: &http.Client{
			//Timeout: 5,
		},
	}
}

// NewHTMLSourceRequestWithSleep creates a new source request with a http client with a sleep timeout
func NewHTMLSourceRequestWithSleep(sleep int) *HTMLSourceRequest {
	return &HTMLSourceRequest{
		client: &http.Client{
			//Timeout: 5,
		},
		SleepTimeMax: sleep,
	}
}

// GetSourceCode get source code from webpage
func (r *HTMLSourceRequest) GetSourceCode(searchURL string, method string, body []byte) (*HtmlData, error) {
	if r.Cache == nil {
		r.Cache = cache.New(5*time.Minute, 10*time.Minute)
	}

	bytes, found := r.Cache.Get(searchURL)
	if found {
		switch b := bytes.(type) {
		case *HtmlData:
			d := HtmlData{}
			data, err := json.Marshal(b)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(data, &d)
			if err != nil {
				return nil, err
			}
			return &d, nil
		}
	}
	httpRequestHandler := NewHTMLSourceRequest()
	u, err := url.Parse(searchURL)
	if err != nil {
		return nil, err
	}
	err = httpRequestHandler.fullRequest(u, method, body)
	if err != nil {
		return nil, err
	}
	pageSource, err := httpRequestHandler.process(0, "", nil)
	if err != nil {
		return nil, err
	}
	r.Cache.Set(searchURL, pageSource, cache.DefaultExpiration)
	r.wait()
	return pageSource, nil
}
func (r *HTMLSourceRequest) wait() {
	if r.SleepTimeMax > 0 {
		rand.Seed(time.Now().Unix())
		min := rand.Intn(r.SleepTimeMax)
		time.Sleep(time.Duration(min) * time.Second)
	}
}
func (r *HTMLSourceRequest) ProcessSourceCode(sourceCode string) (*HtmlData, error) {
	respReader := strings.NewReader(sourceCode)
	r.tokenizer = html.NewTokenizer(respReader)
	pageSource, err := r.process(0, "", nil)
	if err != nil {
		return nil, err
	}
	return pageSource, err
}

// fullRequest sets up tokenizer
func (r *HTMLSourceRequest) fullRequest(url *url.URL, method string, body []byte) error {
	req, err := http.NewRequest(method, url.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		defer func() { _ = resp.Body.Close() }()
		respStr, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		println(string(respStr))
		return errors.New("bad status code")
	}
	defer func() { _ = resp.Body.Close() }()
	respStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	respReader := strings.NewReader(string(respStr))
	r.tokenizer = html.NewTokenizer(respReader)
	r.wait()
	return nil
}

// Download will download a file given a url to a given path
func (r *HTMLSourceRequest) Download(u, path string) (string, error) {
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return "", nil
	}
	endpoint, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	if path == "" {
		sp := strings.Split(endpoint.Path, "/")
		path = sp[len(sp)-1]
	}
	dir, _ := filepath.Split(path)
	if dir != "" {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return "", err
		}
	}

	root := &url.URL{
		Scheme: endpoint.Scheme,
		Opaque: endpoint.Opaque,
		Host:   endpoint.Host,
	}
	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Referer", root.String())
	response, err := r.client.Do(req)
	if err != nil {
		//p.Logger.Error(fmt.Sprintf("failed downloading file from url: %s", url), zap.Error(err))
		return "", nil
	}
	defer func() { _ = response.Body.Close() }()
	if strings.HasSuffix(path, "/") {
		cd := response.Header.Get("content-disposition")
		if cd != "" {
			i := strings.Index(cd, "filename")
			fname := strings.Split(cd[i:], "=")
			filename := strings.ReplaceAll(fname[1], `"`, "")
			reg, err := regexp.Compile("[/?%*:|\"<>]")
			if err == nil {
				filename = reg.ReplaceAllString(filename, "")
			}
			path += filename
		} else if cd = response.Header.Get("filename"); cd != "" {
			path += cd
		}
	}

	file, err := os.Create(path)
	if err != nil && err != os.ErrExist {
		//	p.Logger.Error(fmt.Sprintf("failed creating file path for file %s", path), zap.Error(err))
		return "", nil
	}
	defer func() { _ = file.Close() }()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		//p.Logger.Error("failed saving image", zap.Error(err))
		return "", nil
	}
	r.wait()
	return path, nil
}

func (r *HTMLSourceRequest) process(depth int, currentTag string, parent *HtmlData) (*HtmlData, error) {
	RootHtmlData := &HtmlData{
		ID:         uuid.New().String(),
		Parent:     parent,
		Tag:        currentTag,
		Attributes: map[string]string{},
		Sibling:    []*HtmlData{},
	}
	for {
		nextToken := r.tokenizer.Next()
		token := r.tokenizer.Token()
		err := r.tokenizer.Err()
		if err == io.EOF {
			return RootHtmlData, nil
		}
		switch nextToken {
		case html.ErrorToken:
			if err != nil {
				return RootHtmlData, nil
			}
		case html.SelfClosingTagToken:
			t := &HtmlData{
				Parent:     parent,
				Tag:        token.Data,
				Attributes: map[string]string{},
			}
			for _, v := range token.Attr {
				t.Attributes[v.Key] = v.Val
			}
			RootHtmlData.Sibling = append(RootHtmlData.Sibling, t)
		case html.StartTagToken:
			depth += 1
			child, err := r.process(depth, token.Data, RootHtmlData)
			for _, v := range token.Attr {
				child.Attributes[v.Key] = v.Val
			}
			if err != nil {
				return RootHtmlData, nil
			}
			RootHtmlData.Child = append(RootHtmlData.Child, child)
		case html.EndTagToken:
			return RootHtmlData, nil
		case html.TextToken:
			RootHtmlData.TextData = strings.TrimSpace(RootHtmlData.TextData + token.Data)
		}
	}
}
