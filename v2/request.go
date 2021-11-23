package v2

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type HTMLSourceRequest struct {
	client    *http.Client
	tokenizer *html.Tokenizer
}

func NewHTMLSourceRequest() *HTMLSourceRequest {
	return &HTMLSourceRequest{
		client: &http.Client{
			//Timeout: 5,
		},
	}
}

func (r *HTMLSourceRequest) Request(url url.URL) error {
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("bad status code")
	}
	defer func() { _ = resp.Body.Close() }()
	respStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	respReader := strings.NewReader(string(respStr))
	r.tokenizer = html.NewTokenizer(respReader)
	return nil
}
func (r *HTMLSourceRequest) Download(url, path string, delay int) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	dir, _ := filepath.Split(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	response, err := http.Get(url)
	if err != nil {
		//p.Logger.Error(fmt.Sprintf("failed downloading file from url: %s", url), zap.Error(err))
		return err
	}
	defer func() { _ = response.Body.Close() }()

	file, err := os.Create(path)
	if err != nil {
		//	p.Logger.Error(fmt.Sprintf("failed creating file path for file %s", path), zap.Error(err))
		return err
	}
	defer func() { _ = file.Close() }()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		//p.Logger.Error("failed saving image", zap.Error(err))
		return err
	}
	//p.Logger.Debug(fmt.Sprintf("finished downloading %s/%s",filepath.Base(dir),fileName))
	time.Sleep(time.Duration(delay) * time.Second)
	return nil
}

func (r *HTMLSourceRequest) Process(depth int, currentTag string) (*HtmlData, error) {
	RootHtmlData := &HtmlData{
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
				Tag:        token.Data,
				Attributes: map[string]string{},
			}
			for _, v := range token.Attr {
				t.Attributes[v.Key] = v.Val
			}
			RootHtmlData.Sibling = append(RootHtmlData.Sibling, t)
		case html.StartTagToken:
			depth += 1
			child, err := r.Process(depth, token.Data)
			for _, v := range token.Attr {
				child.Attributes[v.Key] = v.Val
			}
			if err != nil {
				return RootHtmlData, nil
			}
			RootHtmlData.Child = append(RootHtmlData.Child, child)
		case html.EndTagToken:
			depth -= 1
			return RootHtmlData, nil
		case html.TextToken:
			RootHtmlData.TextData = strings.TrimSpace(RootHtmlData.TextData + token.Data)
		}
	}
}

func Download(url, path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	dir, _ := filepath.Split(path)
	if dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed downloading file from url: %s", url)
	}
	defer func() { _ = response.Body.Close() }()

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed creating file path for file %s", path)
	}
	defer func() { _ = file.Close() }()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("failed saving image")
	}
	return nil
}
