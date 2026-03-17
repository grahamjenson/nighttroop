package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

type FlagFileList []string

type FlagStringFileName struct {
	FileName string
	VarName  string
}

func (i *FlagFileList) String() string {
	return fmt.Sprintf("%v", *i)
}
func (i *FlagFileList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *FlagFileList) GetFileVars() []FlagStringFileName {
	var names []FlagStringFileName
	for _, v := range *i {
		names = append(names, FlagStringFileName{
			FileName: v,
			VarName:  SnakeToCapitalCamelCase(FileToName(v)),
		})
	}
	return names
}

func SnakeToCapitalCamelCase(s string) string {
	words := strings.Split(s, "_")
	for i := range words {
		if len(words[i]) > 0 {
			runes := []rune(words[i])
			// Capitalize the first letter
			runes[0] = unicode.ToUpper(runes[0])
			// Lowercase the rest of the letters (optional, depends on specific need)
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			words[i] = string(runes)
		}
	}
	return strings.Join(words, "")
}

// from gds.sql to gds_sql
func FileToName(path string) string {
	// 1. Get "gds.sql" from "asd/gds.sql"
	base := filepath.Base(path)

	// 2. Replace all dots with underscores in "gds.sql" to get "gds_sql"
	name := strings.ReplaceAll(base, ".", "_")

	return name
}
