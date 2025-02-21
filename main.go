package main

import (
	"checker-pa/src/manager"
	"checker-pa/src/utils"
	_ "embed"
	"fmt"

	"github.com/fatih/color"
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

<<<<<<< HEAD
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

		if moduleOutput.Score < 0 {
			color.New(color.FgRed).Printf("MODULE FAILURE: %s\nScore: N/A\n\n\n", module.GetName())
			color.Unset()
			continue
		} else {
			// TODO: if threshold is set, check if score is below threshold then print red, else print green
			fmt.Printf("Score: %d\n\n\n", moduleOutput.Score)
		}
	}
=======
	moduleConfig, err := utils.NewModuleConfig(moduleConfigStr)
	if err != nil {
		panic(err)
	}
>>>>>>> 44b5a3e (infra: modified utils and added concrete types for module config and user config)

	fmt.Println(userConfig)
	fmt.Println(moduleConfig)
}
