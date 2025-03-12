package utils

import (
	"encoding/json"
	"regexp"
)

var commentRegex = regexp.MustCompile("//.*")

func NewUserConfig(source string) (*UserConfig, error) {
	var m UserConfig
	newSource := commentRegex.ReplaceAllString(source, "")
	err := json.Unmarshal([]byte(newSource), &m)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

func newModuleConfig(source string) (*ModuleConfig, error) {
	var m ModuleConfig
	newSource := commentRegex.ReplaceAllString(source, "")
	err := json.Unmarshal([]byte(newSource), &m)

	if err != nil {
		return nil, err
	}

	return &m, nil
}
