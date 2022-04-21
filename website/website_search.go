package website

import "sort"

const (
	TypeTag       = "tag"
	TypeAttribute = "attribute"
)

type Search struct {
	Type            string `json:"type" db:"type"`
	Tag             string `json:"tag" db:"tag"`
	TagValue        string `json:"tag_value" db:"tag_value"`
	Order           int    `json:"order" db:"order"`
	InternalTagName string `json:"internal_tag_name" db:"internal_tag_name"`
	ForwardData     bool   `json:"forward_data"`
	UseChildData    bool   `json:"use_child_data" db:"use_child_data"`
	Flatten         bool   `json:"flatten" db:"flatten"`
	SkipRemap       bool   `json:"skip_remap" db:"skip_remap"`
	OnlyRemap       bool   `json:"only_remap" db:"only_remap"`
}

type CombinedSearch struct {
	Tags         []string
	Attributes   map[string]string
	RemapValues  map[string]string
	ForwardData  bool `json:"forward_data"`
	UseChildData bool `json:"use_child_data" db:"use_child_data"`
	Flatten      bool `json:"flatten" db:"flatten"`
	SkipRemap    bool `json:"skip_remap"`
}

func search(m map[string][]*Search, currentOrder int) *CombinedSearch {
	combinedSearch := &CombinedSearch{
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
			if t.OnlyRemap {
				if len(t.InternalTagName) > 0 {
					combinedSearch.RemapValues[t.Tag] = t.InternalTagName
				}
				continue
			}
			combinedSearch.Tags = append(combinedSearch.Tags, t.Tag)
			if t.UseChildData {
				combinedSearch.UseChildData = t.UseChildData
			}
			if t.ForwardData {
				combinedSearch.ForwardData = t.ForwardData
			}
			if t.SkipRemap {
				combinedSearch.SkipRemap = t.SkipRemap
			}
			if t.Flatten {
				combinedSearch.Flatten = t.Flatten
			}
		}

	}
	for _, t := range attributeList {

		if t.Order == currentOrder {
			if t.OnlyRemap {
				if len(t.InternalTagName) > 0 {
					combinedSearch.RemapValues[t.Tag] = t.InternalTagName
				}
				continue
			}
			combinedSearch.Attributes[t.Tag] = t.TagValue
			if t.UseChildData {
				combinedSearch.UseChildData = t.UseChildData
			}
			if t.ForwardData {
				combinedSearch.ForwardData = t.ForwardData
			}
			if t.SkipRemap {
				combinedSearch.SkipRemap = t.SkipRemap
			}
			if t.Flatten {
				combinedSearch.Flatten = t.Flatten
			}
			if len(t.InternalTagName) > 0 {
				combinedSearch.RemapValues[t.Tag] = t.InternalTagName
			}

		}
	}

	return combinedSearch
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
