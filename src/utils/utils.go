package utils

import (
	"encoding/json"
	"regexp"
)

var commentRegex = regexp.MustCompile("//.*")

type JsonObject map[string]interface{}

func ParseCommentedJSON(source string) (JsonObject, error) {
	var root JsonObject
	uncommentedSource := commentRegex.ReplaceAllString(source, "")

	err := json.Unmarshal([]byte(uncommentedSource), &root)

	return root, err
}

func (root JsonObject) Stringify() string {
	bytes, err := json.MarshalIndent(root, "", "  ")

	if err != nil {
		panic(err)
	}

	return string(bytes[:])
}
