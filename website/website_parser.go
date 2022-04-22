package website

import (
	"fmt"
	v2 "github.com/TheBlockNinja/WebParser/v2"
	"net/http"
	"net/url"
)

type Parser struct {
	Name       string `json:"name"`
	WebsiteURL *url.URL
	SearchList []*Search
}

func NewWebParser(u string, searchData []*Search) (*Parser, error) {
	tmpUrl, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	return &Parser{
		Name:       "default",
		WebsiteURL: tmpUrl,
		SearchList: searchData,
	}, nil
}

func (wp *Parser) Parse(SourceReq *v2.HTMLSourceRequest, searchURL string) ([]*v2.HtmlData, []map[string]string, error) {
	u, err := url.Parse(searchURL)
	if err != nil {
		return nil, nil, err
	}
	if u.Host != wp.WebsiteURL.Host {
		return nil, nil, fmt.Errorf("host parser(%s) does not match search url: %s", wp.WebsiteURL.Host, u.Host)
	}
	source, err := SourceReq.GetSourceCode(searchURL, http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}
	l, maxOrder := separate(wp.SearchList)
	var output []*v2.HtmlData
	var remappedOutput []map[string]string
	for i := 0; i <= maxOrder; i++ {
		combinedData := search(l, i)
		baseSearch := source.Search(combinedData.Tags, combinedData.Attributes)
		for index, d := range baseSearch {
			if combinedData.Flatten {
				baseSearch[index] = d.Flatten(nil)
			}
			r := remap(baseSearch[index], combinedData, u)
			if len(r) > 0 {
				remappedOutput = append(remappedOutput, r)
			}

		}
		if combinedData.ForwardData && i != maxOrder {
			source.Child = baseSearch
		} else {
			output = append(output, baseSearch...)
		}
	}

	return output, remappedOutput, nil
}

func remap(d *v2.HtmlData, combinedSearch *CombinedSearch, baseUrl *url.URL) map[string]string {
	remapped := map[string]string{}
	if combinedSearch.SkipRemap {
		return remapped
	}
	for k, v := range d.Attributes {
		if remapKey, found := combinedSearch.RemapValues[k]; found {
			remapped[remapKey] = v
		} else {
			remapped[k] = v
		}
	}

	link, _ := d.FindLinks(fmt.Sprintf("%s://%s", baseUrl.Scheme, baseUrl.Host), []string{"href", "src", "data-src"})
	addKeyIfExists(combinedSearch, "link", link, remapped)
	addKeyIfExists(combinedSearch, "text", d.TextData, remapped)

	return remapped
}

func addKeyIfExists(combinedSearch *CombinedSearch, key, value string, remap map[string]string) {
	if len(value) == 0 {
		return
	}
	if remapKey, found := combinedSearch.RemapValues[key]; found {
		remap[remapKey] = value
	} else {
		remap[key] = value
	}
}
