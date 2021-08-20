package WebParser

import (
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestGithub(t *testing.T) {
	logger, err := zap.NewProduction()
	parser := NewParser(logger)
	err = parser.Get("https://github.com/TheBlockNinja/WebParser")
	if err != nil {
		parser.Logger.Error("failed load website source", zap.Error(err))
	}
	output := parser.Html.FindAttribute("rel", "fluid-icon")
	for _, i := range output {
		link, err := i.GetLink([]string{"href", "src"}, parser)
		if err != nil {
			parser.Logger.Error("failed getting link", zap.Error(err))
			continue
		}
		extension := strings.Split(link, ".")

		fmt.Printf("Link: %s Title: %s", link, i.Attributes["title"])
		filename := fmt.Sprintf("%s.%s", i.Attributes["title"], extension[len(extension)-1])
		err = parser.Download(link, filename,1)
		if err != nil {
			parser.Logger.Error("failed downloading image", zap.Error(err))
			continue
		}
	}
}

func TestSite(t *testing.T) {
	logger, _ := zap.NewProduction()
	parser := NewParser(logger)
	newSite := Site{
		Name:         "Github",
		URL:          "https://github.com/TheBlockNinja/WebParser",
		Parser:       parser,
		MetaData:     []*SiteMetaData{
			{
				Type:           LinkTypeImage,
				LinkTypes:      []string{"href"},
				Search:         []HTMLSearch{
					{
						Tag:            "",
						Attribute:      "rel",
						Value:          "fluid-icon",
						KeepParentData: false,
					},
				},
				FindAttributes: []string{"title","href"},
				Flatten:        false,
				Attributes: [	]map[string]string{},
				Reverse:        false,
			},
		},
		Recursive:    false,
		MaxDepth:     0,
		Delay:        0,
		Reprocess:    false,
		DownloadPath: "github",
	}
	newSite.Load("https://github.com/TheBlockNinja/WebParser")
	newSite.Download(LinkTypeImage,"href","")
}

func TestSiteLarge(t *testing.T) {
	logger, _ := zap.NewProduction()
	parser := NewParser(logger)
	newSite := Site{
		Name:         "Wiki",
		URL:          "https://en.wikipedia.org/wiki/Wiki",
		Parser:       parser,
		MetaData:     []*SiteMetaData{
			{
				Type:           LinkTypeImage,
				LinkTypes:      []string{"src"},
				Search:         []HTMLSearch{
					{
						Tag:            "img",
						Attribute:      "src",
						Value:          "png",
						KeepParentData: false,
					},
				},
				FindAttributes: []string{"alt","src"},
				Flatten:        false,
				Attributes: 	[]map[string]string{},
				Reverse:        false,
			},
		},
		Recursive:    false,
		MaxDepth:     1,
		Delay:        1,
		Reprocess:    false,
		DownloadPath: "Wiki",
	}
	newSite.Load("https://en.wikipedia.org/wiki/Wiki")
	newSite.LoadMetaData()
	newSite.Download(LinkTypeImage,"alt",".png")
	print()
}