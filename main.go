package main

import (
	"checker-pa/src/utils"
	_ "embed"
	"fmt"
)

//go:embed res/module_config.json
var moduleConfigStr string

//go:embed res/user_config.json
var defaultUerConfigStr string

func main() {

	userConfig, _ := utils.ParseCommentedJSON(defaultUerConfigStr)

	fmt.Println(userConfig.Stringify())

}
