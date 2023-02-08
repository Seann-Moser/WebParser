package analyzer

import (
	"regexp"
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

func (a *Auto) Tags(data *v2.HtmlData, parents int) []string {
	v := data.Search([]string{}, map[string]string{"text": "genre", "text_2": "tags", "text_3": "genres", "*": "tag"}, nil)
	if len(v) == 0 {
		return nil
	}
	skip := []string{}
	for i := 0; i < len(v); i++ {
		current := v[i]
		for j := 0; j < parents; j++ {
			skip = append(skip, current.ID)
			d := current.Search([]string{"p", "span", "a"}, map[string]string{}, skip)
			if len(d) == 0 {
				current = current.Parent
				continue
			}
			output := []string{}
			dup := map[string]struct{}{}
			numReg, _ := regexp.Compile("^[0-9]+")
			for _, s := range d {
				if len(s.TextData) == 0 {
					continue
				}
				if _, found := dup[s.TextData]; found {
					continue
				}
				if numReg.MatchString(s.TextData) {
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

func (a *Auto) Language(data *v2.HtmlData) []string {
	return nil
}

func (a *Auto) Chapters(data *v2.HtmlData, baseLink string) []*v2.HtmlData {
	v := data.Search(nil, map[string]string{"text": "chapter", "class": "chapter", "*": "thumbnail"}, nil)
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

func (a *Auto) Images(data *v2.HtmlData, baseLink string) []string {
	v := data.Search([]string{"script"}, map[string]string{"text": "jpg", "text_1": "jpeg", "text_2": "png"}, nil)
	if len(v) > 0 {
		urlRegex, err := regexp.Compile("((http|https)://)(www.)?[a-zA-Z0-9@:%._\\+~#?&//=]{2,256}\\.[a-z]{2,6}\\b([-a-zA-Z0-9@:%._\\+~#?&//=]*)")
		if err != nil {
			return nil
		}
		reg, err := regexp.Compile("[, \n\"]")
		if err != nil {
			return nil
		}
		var output []string
		splitStrings := reg.Split(v[0].TextData, -1)
		for _, s := range splitStrings {
			if len(strings.TrimSpace(s)) == 0 {
				continue
			}
			s = strings.ReplaceAll(s, "\\/", "/")
			if urlRegex.MatchString(s) {
				output = append(output, s)
			}
		}
		return output
	}

	v = data.Search([]string{"img", "a", "data-src"}, map[string]string{"*": "jpg", "*_1": "jpeg", "*_2": "png"}, nil)
	if len(v) == 0 {
		return nil
	}
	skip := []string{}
	output := []string{}
	for i := 0; i < len(v); i++ {
		current := v[i]
		for j := 0; j < 2; j++ {
			if current.Tag == "head" || current.Tag == "meta" {
				break
			}

			links := current.Search([]string{"a", "img"}, map[string]string{"href": "jpg|png|jpeg", "src": "jpg|png|jpeg", "data-src": "jpg|png|jpeg"}, skip)
			skip = append(skip, current.ID)
			if len(links) == 0 {
				current = current.Parent
				continue
			}

			for _, link := range links {
				link.AddLinkInfo(baseLink, []string{"href", "src", "data-src"})
				l, err := link.FindLinks(baseLink, []string{"href", "data-src", "src"})
				if err != nil || l == "" {
					continue
				}
				output = append(output, l)
			}
		}
	}
	return output
}

func (a *Auto) Pages(data *v2.HtmlData, baseLink string) []string {
	v := data.Search([]string{}, map[string]string{"*": "page"}, nil)
	if len(v) == 0 {
		return nil
	}
	skip := []string{}
	output := []string{}
	dedup := map[string]struct{}{}
	for i := 0; i < len(v); i++ {
		current := v[i]
		for j := 0; j < 2; j++ {
			if current.Tag == "head" || current.Tag == "meta" {
				break
			}

			links := current.Search([]string{"a", "option"}, map[string]string{"href": "^/.*[0-9]+", "src": "^/.*[0-9]+", "value": "^/.*[0-9]+"}, skip)
			skip = append(skip, current.ID)
			if len(links) == 0 {
				current = current.Parent
				continue
			}

			for _, link := range links {
				link.AddLinkInfo(baseLink, []string{"href", "src", "value"})
				l, err := link.FindLinks(baseLink, []string{"href", "src", "value"})
				if err != nil || l == "" {
					continue
				}
				if l == baseLink || !strings.Contains(l, baseLink) {
					continue
				}
				if _, found := dedup[l]; found {
					continue
				}
				output = append(output, l)
				dedup[l] = struct{}{}
			}
		}
	}
	return output
}
