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