package v2

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type HtmlData struct {
	ID         string            `json:"id"`
	Parent     *HtmlData         `json:"parent"`
	Tag        string            `json:"tag"`
	Attributes map[string]string `json:"attributes"`
	TextData   string            `json:"text_data"`
	Child      []*HtmlData       `json:"children"`
	Sibling    []*HtmlData       `json:"siblings"`
}

// Flatten is used to grab all siblings and children and flatten them into a single object
// useful for grabbing text from sites
func (h *HtmlData) Flatten(tags []string, skipId []string) *HtmlData {
	var flatD *HtmlData
	if tags == nil || isInArray(h.Tag, tags) {
		flatD = &HtmlData{
			Tag:        h.Tag,
			Attributes: h.Attributes,
			TextData:   h.TextData,
		}
	}
	for _, c := range h.Child {
		tmp := c.Flatten(tags, skipId)
		if len(tmp.TextData) > 0 && !isInArray(c.ID, skipId) {

			if len(flatD.TextData) == 0 {
				flatD.TextData = tmp.TextData
			} else {
				flatD.TextData += "---" + tmp.TextData
			}
		}

		for k, v := range tmp.Attributes {
			if len(v) > 0 && !isInArray(c.ID, skipId) {
				if len(flatD.Attributes[k]) == 0 {
					flatD.Attributes[k] = v
				} else {
					flatD.Attributes[k] += "," + v
				}
			}

		}
	}
	for _, c := range h.Sibling {
		tmp := c.Flatten(tags, skipId)
		if len(tmp.TextData) > 0 && !isInArray(c.ID, skipId) {
			if len(flatD.TextData) == 0 {
				flatD.TextData = tmp.TextData
			} else {
				flatD.TextData += "---" + tmp.TextData
			}
		}
		for k, v := range tmp.Attributes {
			if len(v) > 0 && !isInArray(c.ID, skipId) {
				if len(flatD.Attributes[k]) == 0 {
					flatD.Attributes[k] = v
				} else {
					flatD.Attributes[k] += "," + v
				}
			}
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

// GetTags will return all html elements that contain the tags provided to the method
// ie img will return a list of all the image tags from a website
func (h *HtmlData) GetTags(tags []string) []*HtmlData {
	var d []*HtmlData
	if tags == nil || isInArray(h.Tag, tags) {
		d = append(d, h)
	}
	for _, c := range h.Child {
		d = append(d, c.GetTags(tags)...)
	}
	for _, c := range h.Sibling {
		d = append(d, c.GetTags(tags)...)
	}
	return d
}

// Search will go through a site and find all tags with the attribute key-value pair
// the attributes value is a regex expression
// EX: "href":".*\.png$" - will match to all href attributes ending with .png
func (h *HtmlData) Search(tags []string, attributes map[string]string, skipId []string) []*HtmlData {
	var output []*HtmlData
	if !isInArray(h.ID, skipId) {
		if tags == nil || isInArray(h.Tag, tags) || len(tags) == 0 {
			if len(attributes) == 0 {
				output = append(output, h)
			}
			for k, v := range attributes {
				if strings.HasPrefix(k, "tag") || strings.HasPrefix(k, "*") {
					foundValue, err := regexp.MatchString(v, h.Tag)
					if err != nil {
						if strings.EqualFold(v, h.Tag) {
							output = append(output, h)
							continue
						}
					}

					if foundValue || Match(v, h.Tag) > 0.5 {
						output = append(output, h)
						continue
					}
				}
				if strings.HasPrefix(k, "text") || strings.HasPrefix(k, "*") {
					foundValue, err := regexp.MatchString(v, h.TextData)
					if err != nil {
						if strings.EqualFold(v, h.TextData) {
							output = append(output, h)
							continue
						}
					}

					if foundValue || Match(v, h.TextData) > 0.5 {
						output = append(output, h)
						continue
					}
				}
				if strings.HasPrefix(k, "*") {
					for _, av := range h.Attributes {
						foundValue, err := regexp.MatchString(v, av)
						if err != nil {
							if strings.EqualFold(v, av) {
								output = append(output, h)
								break
							}
						}

						if foundValue || Match(v, av) > 0.5 {
							output = append(output, h)
							break
						}
					}
				}
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
	}
	for _, c := range h.Child {
		output = append(output, c.Search(tags, attributes, skipId)...)
	}
	for _, s := range h.Sibling {
		output = append(output, s.Search(tags, attributes, skipId)...)
	}
	return output
}

func (h *HtmlData) AddLinkInfo(baseLink string, linkAttributes []string) {
	parsedURL, err := url.Parse(baseLink)
	if err != nil {
		return
	}
	for _, a := range linkAttributes {
		link, ok := h.Attributes[a]
		if !ok {
			continue
		}
		link, _ = url.QueryUnescape(link)
		if link == "" {
			continue
		}
		u, err := url.Parse(link)
		if err != nil {
			continue
		}
		h.Attributes[a] = (&url.URL{
			Scheme:   parsedURL.Scheme,
			Host:     parsedURL.Host,
			Path:     u.Path,
			RawPath:  u.RawPath,
			RawQuery: u.RawQuery,
			Fragment: u.Fragment,
		}).String()
	}
}

// FindLinks will search through the html data and find/build links from paths
// Useful when getting images from sites when they leave out the host and schema
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
