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

func (h *HtmlData) Flatten(tags []string) []FlatData {
	var flatD []FlatData
	if tags == nil || isInArray(h.Tag, tags) {
		flatD = append(flatD, FlatData{
			Tag:        h.Tag,
			Attributes: h.Attributes,
			TextData:   h.TextData,
		})
	}
	for _, c := range h.Child {
		flatD = append(flatD, c.Flatten(tags)...)
	}
	for _, c := range h.Sibling {
		flatD = append(flatD, c.Flatten(tags)...)
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

func (h *HtmlData) getTags(tags []string) {

}
func (h *HtmlData) getAllTags(tags []string) []FlatData {
	return nil
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
