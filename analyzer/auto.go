package parser

import (
	"strings"

	v2 "github.com/Seann-Moser/WebParser/v2"
)

type Auto struct {
}

func NewAuto() *Auto {
	return &Auto{}
}

func (a *Auto) Title(data *v2.HtmlData) string {
	v := data.Search([]string{}, map[string]string{"tag": "title"}, nil)
	if len(v) == 0 {
		return ""
	}
	for i := 0; i < len(v); i++ {
		current := v[i]
		if current.TextData != "" {
			return current.TextData
		}
	}
	return ""
}

func (a *Auto) Description(data *v2.HtmlData) string {
	v := data.Search([]string{}, map[string]string{"text": "description", "text_2": "summary", "text_3": "Synopsis"}, nil)
	if len(v) == 0 {
		return ""
	}
	skip := []string{}
	for i := len(v) - 1; i >= 0; i-- {
		current := v[i]
		for j := 0; j < 3; j++ {
			skip = append(skip, current.ID)
			d := current.Search([]string{"p", "span"}, map[string]string{}, skip)
			if len(d) == 0 {
				current = current.Parent
				continue
			}
			fd := (&v2.HtmlData{
				ID:         "",
				Parent:     nil,
				Tag:        "",
				Attributes: map[string]string{},
				TextData:   "",
				Child:      d,
				Sibling:    nil,
			}).Flatten(nil, skip)
			return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(fd.TextData, "---", " "), "\n", " "), "\t", " ")
		}
	}
	return ""
}

func (a *Auto) Author(data *v2.HtmlData) string {
	return ""
}

func (a *Auto) Tags(data *v2.HtmlData) []string {
	v := data.Search([]string{}, map[string]string{"text": "genre", "text_2": "tags", "text_3": "genres"}, nil)
	if len(v) == 0 {
		return nil
	}
	skip := []string{}
	for i := 0; i < len(v); i++ {
		current := v[i]
		for j := 0; j < 3; j++ {
			skip = append(skip, current.ID)
			d := current.Search([]string{"p", "span", "a"}, map[string]string{}, skip)
			if len(d) == 0 {
				current = current.Parent
				continue
			}
			output := []string{}
			dup := map[string]struct{}{}
			for _, s := range d {
				if len(s.TextData) == 0 {
					continue
				}
				if _, found := dup[s.TextData]; found {
					continue
				}
				output = append(output, s.TextData)
				dup[s.TextData] = struct{}{}
			}
			return output
		}
	}
	return nil
}

func (a *Auto) Dates(data *v2.HtmlData) []string {
	return nil
}

func (a *Auto) Chapters(data *v2.HtmlData, baseLink string) []*v2.HtmlData {
	v := data.Search(nil, map[string]string{"text": "chapter", "class": "chapter"}, nil)
	if len(v) == 0 {
		return nil
	}
	skip := []string{}
	for i := 0; i < len(v); i++ {
		current := v[i]
		for j := 0; j < 2; j++ {
			if current.Tag == "head" || current.Tag == "meta" {
				break
			}
			skip = append(skip, current.ID)

			links := current.Search([]string{"a"}, map[string]string{"href": "^/.*[0-9]+", "src": "^/.*[0-9]+"}, skip)
			if len(links) == 0 {
				current = current.Parent
				continue
			}
			for _, link := range links {
				link.AddLinkInfo(baseLink, []string{"href", "src"})
			}
			return links
			//return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(fd.TextData, "---", " "), "\n", " "), "\t", " ")
		}
	}
	return nil
}
