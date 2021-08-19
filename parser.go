package WebParser

import (
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Parser struct {
	Logger *zap.Logger
	URL    string
	Body   string
	Tokens *html.Tokenizer
	Html   *HTMLData
}

func NewParser(Logger *zap.Logger) *Parser {
	return &Parser{
		Logger: Logger,
	}
}

func (p *Parser) load(url string) error {
	response, err := http.Get(url)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed to load site %s", url), zap.Error(err))
		return err
	}
	defer func() { _ = response.Body.Close() }()
	byteBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed to load site body for %s", url), zap.Error(err))
		return err
	}
	p.URL = url
	p.Body = string(byteBody)
	return nil
}

func (p *Parser) Get(url string) error {
	if p.URL != url {
		err := p.load(url)
		if err != nil {
			p.Logger.Error(fmt.Sprintf("failed getting url source: %s", url), zap.Error(err))
			return err
		}
	}
	myReader := strings.NewReader(p.Body)
	p.Tokens = html.NewTokenizer(myReader)
	data, err := p.getHtmlDataR(0, "main")
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed parsing url data: %s", url), zap.Error(err))
		return err
	}
	p.Html = data
	return nil
}

func (p *Parser) Download(url, path string,delay int) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	dir,fileName := filepath.Split(path)
	if dir != ""{
		err := os.MkdirAll(dir,0755)
		if err != nil{
			return err
		}
	}
	response, err := http.Get(url)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed downloading file from url: %s", url), zap.Error(err))
		return err
	}
	defer func() { _ = response.Body.Close() }()

	file, err := os.Create(path)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed creating file path for file %s", path), zap.Error(err))
		return err
	}
	defer func() { _ = file.Close() }()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		p.Logger.Error("failed saving image", zap.Error(err))
		return err
	}
	p.Logger.Debug(fmt.Sprintf("finished downloading %s/%s",filepath.Base(dir),fileName))
	time.Sleep(time.Duration(delay) * time.Second)
	return nil
}
