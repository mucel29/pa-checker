package main

import (
	"checker-pa/src/display"
	"checker-pa/src/manager"
	"checker-pa/src/menu"
	"checker-pa/src/utils"
	_ "embed"
	"flag"
	"log"
)

//go:embed res/module_config.json
var moduleConfigStr string

//go:embed res/user_config.json
var defaultUserConfigStr string

var useInteractive bool

func init() {
	flag.BoolVar(&useInteractive, "i", false, "Interactive mode")
}

func main() {
	flag.Parse()

	err := utils.InitConfig(defaultUserConfigStr, moduleConfigStr)
	if err != nil {
		log.Fatalln("FATAL ERROR DETECTED! " + err.Error() + "\n ABORTING!")
	}

	// TODO: check for user config
	m, err := manager.NewManager()

	if err != nil {
		log.Fatalln("FATAL ERROR DETECTED! " + err.Error() + "\n ABORTING!")
	}

	err = m.Run()
	if err != nil {
		log.Fatalln("FATAL ERROR DETECTED! " + err.Error() + "\n ABORTING!")
	}

	//TODO?: make this return an error as well?
	m.Check()

	if useInteractive {

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
