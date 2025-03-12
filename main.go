package main

import (
	"checker-pa/src/display"
	"checker-pa/src/manager"
	"checker-pa/src/menu"
	"checker-pa/src/utils"
	_ "embed"
	"flag"
)

//go:embed res/module_config.json
var moduleConfigStr string

//go:embed res/user_config.json
var defaultUserConfigStr string

func main() {

	utils.InitConfig(defaultUserConfigStr, moduleConfigStr)
	useInteractive := flag.Bool("i", false, "Interactive mode")

	// TODO: check for user config

	m, err := manager.NewManager()

	if err != nil {
		panic(err)
	}

	flag.Parse()

	err = m.Run()
	if err != nil {
		panic(err)
	}
	m.Check()

	if *useInteractive {

		utils.Log("Interactive Display")
		d := display.NewDisplay()

		mn := menu.Menu{Display: d, Manager: m}

		mn.Launch()

		d.Enable()

	} else {
		utils.Log("Basic Display")

		for _, module := range m.Modules {
			module.Dump()
		}

	}
}
