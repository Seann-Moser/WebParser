package website

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"

	v2 "github.com/Seann-Moser/WebParser/v2"
)

// Parser an object that combines multiple search's to get more specific data
// This can also be used with QueryHelper
type Parser struct {
	ID         string    `json:"id" db:"id" join_name:"id"`
	Name       string    `json:"name" joinable:"false"`
	WebsiteURL string    `json:"website_url" where:"=" joinable:"false"`
	SearchList []*Search `json:"search_list" skip_table:"true"`
}

// NewWebParser creates a new parser from a list of search data
func NewWebParser(name, u string, searchData []*Search) (*Parser, error) {
	return &Parser{
		ID:         uuid.New().String(),
		Name:       name,
		WebsiteURL: u,
		SearchList: searchData,
	}, nil
}

// Parse will take in an htmlSourceRequest and an url to retrieve information from that site
func (wp *Parser) Parse(SourceReq *v2.HTMLSourceRequest, searchURL string) ([]*v2.HtmlData, []map[string]string, error) {
	u, err := url.Parse(searchURL)
	if err != nil {
		return nil, nil, err
	}
	tmpUrl, err := url.Parse(wp.WebsiteURL)
	if err != nil {
		return nil, nil, err
	}
	if u.Host != tmpUrl.Host {
		return nil, nil, fmt.Errorf("host parser(%s) does not match search url: %s", tmpUrl.Host, u.Host)
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
		baseSearch := source.Search(combinedData.Tags, combinedData.Attributes, nil)
		for index, d := range baseSearch {
			if combinedData.Flatten {
				baseSearch[index] = d.Flatten(nil, nil)
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

func remap(d *v2.HtmlData, combinedSearch *combinedSearch, baseUrl *url.URL) map[string]string {
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

func addKeyIfExists(combinedSearch *combinedSearch, key, value string, remap map[string]string) {
	if len(value) == 0 {
		return
	}
	if remapKey, found := combinedSearch.RemapValues[key]; found {
		remap[remapKey] = value
	} else {
		remap[key] = value
	}
}
