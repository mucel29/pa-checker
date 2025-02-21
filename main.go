package main

import (
	"checker-pa/src/manager"
	"checker-pa/src/utils"
	_ "embed"
	"fmt"
)

//go:embed res/module_config.json
var moduleConfigStr string

//go:embed res/user_config.json
var defaultUserConfigStr string

func main() {

	utils.InitConfig(defaultUserConfigStr, moduleConfigStr)
	// TODO: check for user config

	m, err := manager.NewManager()

	if err != nil {
		panic(err)
	}

	fmt.Println(utils.Config.UserConfig)
	fmt.Println(utils.Config.ModuleConfig)
	fmt.Println()

	m.Run()

	for _, module := range m.Modules {
		moduleOutput := module.Details()
		if moduleOutput.Error != nil {
			fmt.Println(moduleOutput.Error)
		}

		fmt.Printf("Score: %d\n\n\n", moduleOutput.Score)
	}

}
