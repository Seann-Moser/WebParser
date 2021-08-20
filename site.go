package WebParser

import (
	"fmt"
	"github.com/jinzhu/copier"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Site struct{
	Name string `json:"name"`
	URL string `json:"url"`
	Parser *Parser `json:"parser"`
	MetaData []*SiteMetaData
	Recursive bool `json:"recursive"`
	MaxDepth int `json:"max_depth"`
	Delay int `json:"delay"`
	Reprocess bool `json:"reprocess"`
	DownloadPath string `json:"download_path"`
}

const (
	LinkTypeImage = "IMAGE"
	LinkTypeSubPage = "PAGE"
	LinkTypeVideo = "VIDEO"
	LinkTypeLink = "LINK"
	LinkTypeText = "Text"
)


type SiteMetaData struct{
	Type string `json:"type"`
	LinkTypes []string `json:"link_types"`
	Search []HTMLSearch `json:"search"`
	FindAttributes []string `json:"find_attributes"`
	Flatten bool `json:"flatten"`
	Attributes []map[string]string `copier:"-" json:"attributes"` //Type, Data
	Reverse bool `json:"reverse"`
}

func (s *Site) Load(url string) error{
	logger, _ := zap.NewDevelopment()
	s.Parser = NewParser(logger)
	err := s.Parser.Get(url)
	if err != nil{
		return err
	}
	s.LoadMetaData()
	return nil
}

func (s *Site) LoadMetaData(){
	var err error
	for i,l := range s.MetaData{
		htmlData := s.Parser.Html.AdvanceSearches(l.Search)
		if l.Reverse{
			htmlData = Reverse(htmlData)
		}
		for _,d := range htmlData{
			if len(l.Attributes) > 0 && len(l.Attributes[len(l.Attributes)-1]) == 0{

			}else{
				l.Attributes = append(l.Attributes,map[string]string{})
			}
			if l.Flatten{
				d = d.FlattenFull()
			}

			for _,attribute := range l.FindAttributes{
				if attribute == "text"{
					if d.TextData == ""{
						continue
					}
					l.Attributes[len(l.Attributes) -1][attribute] = d.TextData
				}else{
					if d.Attributes[attribute] == ""{
						continue
					}
					if IsLink(*l,attribute){
						l.Attributes[len(l.Attributes) -1][attribute],err = d.GetLink([]string{attribute}, s.Parser)
						if err != nil{
							l.Attributes[len(l.Attributes) -1][attribute] = d.Attributes[attribute]
						}
					}else{
						l.Attributes[len(l.Attributes) -1][attribute] = d.Attributes[attribute]
					}
				}
			}
		}
		s.MetaData[i] = l
	}
}

type Links struct {
	Key string `json:"key"`
	Value string `json:"value"`
}
func (s *Site) GetLinks(linkType,nameAttribute string) ([]*Links,error){
	var testOutput []*Links
	switch linkType {
	case LinkTypeSubPage:
		for _, m := range s.MetaData {
			if m.Type == LinkTypeSubPage {
				for i, d := range m.Attributes {
					for k, v := range d {
						if IsLink(*m, k) {
							name := strconv.Itoa(i)
							if nameAttribute != "" && d[nameAttribute] != "" {
								name = d[nameAttribute]
							}
							testOutput = append(testOutput, &Links{
								Key:   name,
								Value: v,
							})
							break
						}
					}
				}
			}
		}
	case LinkTypeImage:
		for _, m := range s.MetaData {
			if m.Type == LinkTypeImage {
				for i, d := range m.Attributes {
					for k, v := range d {
						if IsLink(*m, k) {
							name := strconv.Itoa(i)
							if nameAttribute != "" && d[nameAttribute] != "" {
								name = d[nameAttribute]
							}
							if nameAttribute == ""{
								test := strings.Split(v,"/")
								name = test[len(test)-1]
							}
							testOutput = append(testOutput, &Links{
								Key:   name,
								Value: v,
							})
							break
						}
					}
				}
			}
		}
	case LinkTypeText:
		test := ""
		for _, m := range s.MetaData {
			if m.Type == LinkTypeText {
				for _, d := range m.Attributes {
					for _, v := range d {
						test += v
					}
				}
			}
		}
		testOutput = append(testOutput, &Links{
			Key:   "text",
			Value: test,
		})

	case LinkTypeVideo:
	default:
		return testOutput,fmt.Errorf("link type does not exist")
	}
	return testOutput,nil
}

func (s *Site)NeedToProcess() bool{
	if !s.Reprocess{
		p := filepath.Join(s.DownloadPath,s.Name)
		if _, err := os.Stat(p); err == nil {
			return false
		}
		return true
	}
	return true
}

func (s *Site) Download(linkType,nameAttribute,ext string) error{
	switch  linkType{
	case LinkTypeImage:
		links,err := s.GetLinks(linkType,nameAttribute)
		if err != nil{
			return err
		}
		for _,l := range links{
			err = multierr.Append(err, s.Parser.Download(l.Value,filepath.Join(s.DownloadPath,s.Name,l.Key)+ext, s.Delay))
		}
		return err
	}
	return nil
}

func IsLink(metaData SiteMetaData,attribute string) bool{
	for _,a := range metaData.LinkTypes{
		if a == attribute{
			return true
		}
	}
	return false
}


func (s *Site) GetSubPages(nameAttribute string,subSiteData Site,loadSub bool,limitLinks int) ([]*Site,error){
	var output []*Site
	links,err := s.GetLinks(LinkTypeSubPage,nameAttribute)
	if err != nil{
		return nil,err
	}
	for i,d := range links{
		newSite := &Site{
			Name:      d.Key,
			URL:       d.Value,
			MetaData: []*SiteMetaData{},
			Recursive: subSiteData.Recursive,
			MaxDepth:  subSiteData.MaxDepth,
			DownloadPath: subSiteData.DownloadPath,
			Delay: subSiteData.Delay,
			Reprocess: s.Reprocess || subSiteData.Reprocess,
		}
		if newSite.DownloadPath == ""{
			newSite.DownloadPath = s.DownloadPath
		}
		if subSiteData.MetaData != nil{
			err = copier.Copy(&newSite.MetaData,subSiteData.MetaData)
			if err != nil{
				return nil,err
			}
		}
		if loadSub && newSite.NeedToProcess(){
			time.Sleep(time.Duration(s.Delay) * time.Second)
			err := newSite.Load(d.Value)
			if err != nil{
				return nil,err
			}
		}
		output = append(output, newSite)
		if limitLinks > -1 && i >= limitLinks - 1 {
			return output,nil
		}
	}

	return output,nil
}

//func (s *Site) GetSubPagesRecursive(nameAttribute string,depth int,currentLinks map[string]string) ([]*Site,error){
	//output := []*Site{}
	//if depth >= s.MaxDepth{
	//	return output,nil
	//}
	//links,err := s.GetLinks(LinkTypeSubPage,nameAttribute)
	//if err != nil{
	//	return nil,err
	//}
	//links = getDifferentLinks(currentLinks,links)
	//for name,url := range links{
	//	newSite := &Site{
	//		Name:      name,
	//		URL:       url,
	//		MetaData: []*SiteMetaData{},
	//		Recursive: s.Recursive,
	//		MaxDepth:  s.MaxDepth,
	//		Delay: s.Delay,
	//	}
	//	if s.MetaData != nil{
	//		err = copier.Copy(&newSite.MetaData,s.MetaData)
	//		if err != nil{
	//			return nil,err
	//		}
	//	}
	//	output = append(output, newSite)
	//	if depth < s.MaxDepth - 1 {
	//		time.Sleep(time.Duration(s.Delay) * time.Second)
	//		err = newSite.Load(url)
	//		if err != nil {
	//			return nil, err
	//		}
	//		otherSites, err := newSite.GetSubPagesRecursive(nameAttribute, depth+1,merge(currentLinks,links))
	//		if err != nil {
	//			return nil, err
	//		}
	//		output = append(output, otherSites...)
	//	}
	//
	//}
//
//	return nil,nil
//}
//
//func getDifferentLinks(current,new map[string]string) map[string]string{
//	output := map[string]string{}
//	for k,v := range new{
//		if _,found := current[k];!found{
//			output[k] = v
//		}
//	}
//	return output
//}
//func merge(m1,m2 map[string]string) map[string]string{
//	output := m1
//	for k,v := range m2{
//		if _,found := output[k];!found{
//			output[k] = v
//		}
//	}
//	return output
//}