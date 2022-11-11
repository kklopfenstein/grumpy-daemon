package utils

import (
	"fmt"
	"regexp"
)

func ContainsSearch(content string, search string) bool {
	reg, err := regexp.Compile(fmt.Sprintf("\\b(%s)\\b", regexp.QuoteMeta(search)))
	if err != nil {
		return false
	}
	return reg.Match([]byte(content))
}
