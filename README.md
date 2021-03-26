### WebParser
This is an HTML parser to act like a simple version pythons BS4.


### Uses

```go
    logger, err := zap.NewProduction()
    parser := NewParser(logger)
    err = parser.Get("https://github.com/TheBlockNinja/WebParser")
    if err != nil {
        parser.Logger.Error("failed load website source", zap.Error(err))
    }
    output := parser.html.FindAttribute("rel", "fluid-icon")
    for _, i := range output {
        link, err := i.GetLink([]string{"href", "src"}, parser)
        if err != nil {
            parser.Logger.Error("failed getting link", zap.Error(err))
            continue
        }
        extension := strings.Split(link, ".")
        filename := fmt.Sprintf("%s.%s", i.Attributes["title"], extension[len(extension)-1])

        fmt.Printf("Link: %s Title: %s\n", link, i.Attributes["title"])
        err = parser.Download(link, filename)
        if err != nil {
            parser.Logger.Error("failed downloading image", zap.Error(err))
            continue
        }
    }
```

### Functions

- Parser
    - Load(url string) error
    - Get(url string) error 
    - Download(url, path string) error
  
- HTMLData
    - GetLink(attribute []string, p *Parser) (string, error)
    - Flatten() *HTMLData
    - Reverse(data []*HTMLData) []*HTMLData
    - FindTag(tag string) []*HTMLData
    - FindAttribute(attribute, value string) []*HTMLData
    - Find(tag, attribute, value string) []*HTMLData

```go
type HTMLData struct {
	Tag        string
	Attributes map[string]string
	TextData   string
	Child      []*HTMLData
	Sibling    []*HTMLData
}
```