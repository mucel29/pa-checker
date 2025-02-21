package main

import (
	"checker-pa/src/utils"
	_ "embed"
	"fmt"
)

//go:embed res/module_config.json
var moduleConfigStr string

//go:embed res/user_config.json
var defaultUserConfigStr string

func main() {
	userConfig, err := utils.NewUserConfig(defaultUserConfigStr)
	if err != nil {
		panic(err)
	}

	moduleConfig, err := utils.NewModuleConfig(moduleConfigStr)
	if err != nil {
		panic(err)
	}

	fmt.Println(userConfig)
	fmt.Println(moduleConfig)
}
