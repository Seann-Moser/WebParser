package webParser

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/net/html"
)

type HTMLData struct {
	Tag        string            `json:"tag"`
	Attributes map[string]string `json:"attributes"`
	TextData   string            `json:"text_data"`
	Child      []*HTMLData       `json:"children"`
	Sibling    []*HTMLData       `json:"siblings"`
}

func (p *Parser) getHtmlDataR(depth int, currentTag string) (*HTMLData, error) {
	RootHtmlData := &HTMLData{
		Tag:        currentTag,
		Attributes: map[string]string{},
		Sibling:    []*HTMLData{},
	}
	for {
		nextToken := p.Tokens.Next()
		token := p.Tokens.Token()
		err := p.Tokens.Err()
		if err == io.EOF {
			return RootHtmlData, nil
		}
		switch nextToken {
		case html.ErrorToken:
			p.Logger.Error("failed getting token", zap.Error(err))
			if err != nil {
				return RootHtmlData, nil
			}
		case html.SelfClosingTagToken:
			t := &HTMLData{
				Tag:        token.Data,
				Attributes: map[string]string{},
			}
			for _, v := range token.Attr {
				t.Attributes[v.Key] = v.Val
			}
			RootHtmlData.Sibling = append(RootHtmlData.Sibling, t)
		case html.StartTagToken:
			depth += 1

			child, err := p.getHtmlDataR(depth, token.Data)
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
func (h *HTMLData) GetLink(attribute []string, p *Parser) (string, error) {
	var link string
	var err error
	var ok bool
	for _, a := range attribute {
		link, ok = h.Attributes[a]
		if ok {
			break
		}
	}
	link, _ = url.QueryUnescape(link)
	if link == "" {
		return "", errors.New(fmt.Sprintf("failed to find attribute %s", attribute))
	}
	if strings.Contains(link, "http") {
		links := "http" + strings.Split(link, "http")[1]
		if strings.Contains(links, "?") {
			return links, nil
		}
		if strings.Contains(links, "&") {
			links = strings.Split(links, "&")[0]
			return links, nil
		}
		return links, nil
	}
	if !strings.HasPrefix(link, "/") {
		return "", errors.New(fmt.Sprintf("link does not have the prefix / :%s", link))
	}
	parsedURL, err := url.Parse(p.URL)
	if err != nil {
		return "", err
	}
	output := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, link)
	return output, err
}

func (h *HTMLData) Flatten() *HTMLData {
	output := &HTMLData{
		Tag:        h.Tag,
		Attributes: h.Attributes,
		TextData:   h.TextData,
	}

	for _, c := range h.Child {
		for k, v := range c.Attributes {
			if _, ok := output.Attributes[k]; !ok {
				output.Attributes[k] = v
			}
		}
		output.TextData += c.TextData
	}
	for _, c := range h.Sibling {
		for k, v := range c.Attributes {
			if _, ok := output.Attributes[k]; !ok {
				output.Attributes[k] = v
			}
		}
		output.TextData += c.TextData
	}
	return output
}

func (h *HTMLData) FindTag(tag string) []*HTMLData {
	var output []*HTMLData
	if h.Tag == tag {
		output = append(output, h)
	}
	for _, c := range h.Child {
		output = append(output, c.FindTag(tag)...)
	}
	for _, s := range h.Sibling {
		output = append(output, s.FindTag(tag)...)
	}
	return output
}

func Reverse(data []*HTMLData) []*HTMLData {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	return data
}

func (h *HTMLData) FindAttribute(attribute, value string) []*HTMLData {
	var output []*HTMLData
	if found, ok := h.Attributes[attribute]; ok {
		if value == "" {
			output = append(output, h)
		} else {
			if strings.Contains(found, value) {
				output = append(output, h)
			}
		}
	}

	for _, c := range h.Child {
		output = append(output, c.FindAttribute(attribute, value)...)
	}
	for _, s := range h.Sibling {
		output = append(output, s.FindAttribute(attribute, value)...)
	}
	return output
}

func (h *HTMLData) Find(tag, attribute, value string) []*HTMLData {
	var output []*HTMLData
	if h.Tag == tag {
		if found, ok := h.Attributes[attribute]; ok {
			if value == "" {
				output = append(output, h)
			} else {
				if strings.Contains(found, value) {
					output = append(output, h)
				}
			}
		}
	}

	for _, c := range h.Child {
		output = append(output, c.Find(tag, attribute, value)...)
	}
	for _, s := range h.Sibling {
		output = append(output, s.Find(tag, attribute, value)...)
	}
	return output
}
