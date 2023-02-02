package v2

import (
	"regexp"
	"strings"
)

func Match(baseName, name string) float64 {
	splitReg, err := regexp.Compile("[ ]")
	if err != nil {
		return 0
	}

	name = strings.ToLower(name)
	baseName = strings.ToLower(baseName)
	shortString := splitReg.Split(strings.ToLower(name), -1)
	longString := splitReg.Split(strings.ToLower(baseName), -1)
	if len(name) > len(baseName) {
		longString = splitReg.Split(name, -1)
		shortString = splitReg.Split(baseName, -1)
	}
	if len(shortString) == 0 {
		return 0
	}
	if len(longString) == 0 {
		return 0
	}
	//maxLen := len(longString)
	type MatchIndex struct {
		Value   string
		Offset  int
		Percent float64
		Index   int
	}
	total := []*MatchIndex{}
	for i, v := range longString {
		currentMatch := MatchIndex{Offset: (len(longString) + len(shortString)) / 2, Percent: 0.0}
		for shortI, shortValue := range shortString {
			if strings.EqualFold(v, shortValue) {
				currentMatch.Offset = i - shortI
				currentMatch.Value = v
				currentMatch.Percent = 1
				currentMatch.Index = shortI
				break
			}
			s := v
			l := shortValue
			if len(s) > len(l) {
				s = shortValue
				l = v
			}
			diffLen := len(strings.ReplaceAll(l, s, ""))
			currentPercent := 1 - (float64(diffLen) / float64(len(l)))
			if strings.Contains(l, s) && currentPercent > currentMatch.Percent {
				currentMatch.Offset = i - shortI
				currentMatch.Value = v
				currentMatch.Percent = currentPercent
				currentMatch.Index = shortI
			}
		}
		total = append(total, &currentMatch)
	}
	sum := 0.0
	for _, v := range total {
		sum += v.Percent
	}
	return sum / float64(len(total))
}
