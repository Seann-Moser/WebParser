package v2

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type HtmlData struct {
	Tag        string            `json:"tag"`
	Attributes map[string]string `json:"attributes"`
	TextData   string            `json:"text_data"`
	Child      []*HtmlData       `json:"children"`
	Sibling    []*HtmlData       `json:"siblings"`
}
type FlatData struct {
	Tag        string            `json:"tag"`
	Attributes map[string]string `json:"attributes"`
	TextData   string            `json:"text_data"`
}

func (f *FlatData) FindLinks(baseLink string, linkAttributes []string) (string, error) {
	for _, a := range linkAttributes {
		link, ok := f.Attributes[a]
		if !ok {
			continue
		}
		link, _ = url.QueryUnescape(link)
		if link == "" {
			continue
		}
		if strings.HasPrefix(link, "//") {
			return "https:" + link, nil
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
			return "", nil
		}
		parsedURL, err := url.Parse(baseLink)
		if err != nil {
			return "", err
		}
		output := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, link)
		return output, err
	}
	return "", nil
}

func (h *HtmlData) Flatten(tags []string) *HtmlData {
	var flatD *HtmlData
	if tags == nil || isInArray(h.Tag, tags) {
		flatD = &HtmlData{
			Tag:        h.Tag,
			Attributes: h.Attributes,
			TextData:   h.TextData,
		}
	}
	for _, c := range h.Child {
		tmp := c.Flatten(tags)
		if len(tmp.TextData) > 0 {
			if len(flatD.TextData) == 0 {
				flatD.TextData = tmp.TextData
			} else {
				flatD.TextData += "---" + tmp.TextData
			}
		}

		for k, v := range tmp.Attributes {
			if len(v) > 0 {
				if len(flatD.Attributes[k]) == 0 {
					flatD.Attributes[k] = v
				} else {
					flatD.Attributes[k] += "," + v
				}
			}

		}
	}
	for _, c := range h.Sibling {
		tmp := c.Flatten(tags)
		if len(tmp.TextData) > 0 {
			if len(flatD.TextData) == 0 {
				flatD.TextData = tmp.TextData
			} else {
				flatD.TextData += "---" + tmp.TextData
			}
		}
		for k, v := range tmp.Attributes {
			flatD.Attributes[k] += "," + v
		}
	}

	return flatD
}

func isInArray(v string, v1 []string) bool {
	for _, d := range v1 {
		if strings.EqualFold(d, v) {
			return true
		}
	}
	return false
}

func (h *HtmlData) getTags(tags []string) []*HtmlData {
	var d []*HtmlData
	if tags == nil || isInArray(h.Tag, tags) {
		d = append(d, h)
	}
	for _, c := range h.Child {
		d = append(d, c.getTags(tags)...)
	}
	for _, c := range h.Sibling {
		d = append(d, c.getTags(tags)...)
	}
	return d
}

func (h *HtmlData) getAttributes(attributes []string) map[string]string {
	return nil
}

func (h *HtmlData) getAllAttributes(attributes []string) []FlatData {
	return nil
}

func (h *HtmlData) Search(tags []string, attributes map[string]string) []*HtmlData {
	output := []*HtmlData{}
	if tags == nil || isInArray(h.Tag, tags) {
		if len(attributes) == 0 {
			output = append(output, h)
		}
		for k, v := range attributes {
			if attribute, found := h.Attributes[k]; found {
				foundValue, err := regexp.MatchString(v, attribute)
				if err != nil {
					if strings.EqualFold(v, attribute) {
						output = append(output, h)
					}
				}
				if foundValue {
					output = append(output, h)
				}
			}
		}
	}
	for _, c := range h.Child {
		output = append(output, c.Search(tags, attributes)...)
	}
	for _, s := range h.Sibling {
		output = append(output, s.Search(tags, attributes)...)
	}
	return output
}
func (h *HtmlData) FindLinks(baseLink string, linkAttributes []string) (string, error) {
	for _, a := range linkAttributes {
		link, ok := h.Attributes[a]
		if !ok {
			continue
		}
		link, _ = url.QueryUnescape(link)
		if link == "" {
			continue
		}
		if strings.HasPrefix(link, "//") {
			return "https:" + link, nil
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
			return "", nil
		}
		parsedURL, err := url.Parse(baseLink)
		if err != nil {
			return "", err
		}
		output := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, link)
		return output, err
	}
	return "", nil
}
