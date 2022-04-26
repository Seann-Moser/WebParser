package website

import "sort"

const (
	TypeTag       = "tag"
	TypeAttribute = "attribute"
)

// Search is used to build a parser that gets data from a website
// is an object that can be loaded into the db using QueryHelper
type Search struct {
	SiteID          string `json:"site_id" db:"site_id" table:"primary" where:"=" join_name:"id"`
	Type            string `json:"type" db:"type" table:"primary" can_update:"true"`
	Tag             string `json:"tag" db:"tag" table:"primary" can_update:"true"`
	TagValue        string `json:"tag_value" db:"tag_value" can_update:"true"`
	Order           int    `json:"search_order" db:"search_order" table:"primary" can_update:"true"`
	InternalTagName string `json:"internal_tag_name" db:"internal_tag_name" can_update:"true"`
	ForwardData     bool   `json:"forward_data" db:"forward_data" can_update:"true"`
	UseChildData    bool   `json:"use_child_data" db:"use_child_data" can_update:"true"`
	Flatten         bool   `json:"flatten" db:"flatten" can_update:"true"`
	SkipRemap       bool   `json:"skip_remap" db:"skip_remap" can_update:"true"`
	OnlyRemap       bool   `json:"only_remap" db:"only_remap" can_update:"true"`
}

type combinedSearch struct {
	Tags         []string
	Attributes   map[string]string
	RemapValues  map[string]string
	ForwardData  bool `json:"forward_data"`
	UseChildData bool `json:"use_child_data" db:"use_child_data"`
	Flatten      bool `json:"flatten" db:"flatten"`
	SkipRemap    bool `json:"skip_remap"`
}

func search(m map[string][]*Search, currentOrder int) *combinedSearch {
	combinedSearch := &combinedSearch{
		Tags:         []string{},
		Attributes:   map[string]string{},
		RemapValues:  map[string]string{},
		ForwardData:  false,
		UseChildData: false,
	}
	tagList := m[TypeTag]
	attributeList := m[TypeAttribute]

	for _, t := range tagList {
		if t.Order == currentOrder {
			t.updateCombinedSearch(combinedSearch)
		}
	}
	for _, t := range attributeList {
		if t.Order == currentOrder {
			t.updateCombinedSearch(combinedSearch)
		}
	}

	return combinedSearch
}

func (s *Search) updateCombinedSearch(cs *combinedSearch) {
	if s.OnlyRemap {
		if len(s.InternalTagName) > 0 {
			cs.RemapValues[s.Tag] = s.InternalTagName
		}
		return
	}
	switch s.Type {
	case TypeAttribute:
		cs.Attributes[s.Tag] = s.TagValue
	case TypeTag:
		cs.Tags = append(cs.Tags, s.Tag)
	}

	if s.UseChildData {
		cs.UseChildData = s.UseChildData
	}
	if s.ForwardData {
		cs.ForwardData = s.ForwardData
	}
	if s.SkipRemap {
		cs.SkipRemap = s.SkipRemap
	}
	if s.Flatten {
		cs.Flatten = s.Flatten
	}
	if len(s.InternalTagName) > 0 {
		cs.RemapValues[s.Tag] = s.InternalTagName
	}
}

func separate(searchList []*Search) (map[string][]*Search, int) {
	sort.Slice(searchList, func(i, j int) bool {
		return searchList[i].Order < searchList[j].Order
	})
	output := map[string][]*Search{}
	maxOrder := 0
	for _, item := range searchList {
		if v, found := output[item.Type]; found {
			output[item.Type] = append(v, item)
		} else {
			output[item.Type] = []*Search{item}
		}
		if item.Order > maxOrder {
			maxOrder = item.Order
		}
	}
	return output, maxOrder
}
