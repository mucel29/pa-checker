package main

import (
	checkermodules "checker-pa/src/checker-modules"
	"checker-pa/src/display"
	"checker-pa/src/manager"
	"checker-pa/src/menu"
	"checker-pa/src/utils"
	_ "embed"
	"flag"
	"fmt"
	"strings"
)

//go:embed res/config/module_config.json
var moduleConfigStr string

//go:embed res/config/user_config.json
var defaultUserConfigStr string

var useInteractive bool

func init() {
	flag.BoolVar(&useInteractive, "i", false, "Interactive mode")
}

func main() {
	flag.Parse()

	err := utils.InitConfig(defaultUserConfigStr, moduleConfigStr)
	if err != nil {
		utils.Fatal("FATAL ERROR DETECTED! " + err.Error() + "\n ABORTING!")
	}

	m, err := manager.NewManager()

	if err != nil {
		utils.Fatal("FATAL ERROR DETECTED! " + err.Error() + "\n ABORTING!")
	}

	// TODO?: make this return an error as well?
	// m.Check()

	if useInteractive {

		utils.Log("Interactive Display")
		d := display.NewDisplay()

		go func() {
			err := m.Run()
			if err != nil {
				defer d.Stop()
				utils.Fatal("FATAL ERROR DETECTED! " + err.Error() + "\n ABORTING!")
			}
		}()

		mn := menu.Menu{Display: d, Manager: m}

		mn.Launch()

		d.Enable()

	} else {
		err = m.Run()
		if err != nil {
			utils.Fatal("FATAL ERROR DETECTED! " + err.Error() + "\n ABORTING!")
		}

		utils.Log("Basic Display")

		for _, module := range m.Modules {
			module.Dump()
		}

		summary := strings.Builder{}

		summary.WriteString("\n===== Summary =====\n")

		for _, module := range m.Modules {
			if module.GetStatus() != checkermodules.Ready {
				summary.WriteString(fmt.Sprintf("%-7s - %-8s\n", module.GetName(), module.GetStatus().String()))
			} else {
				summary.WriteString(fmt.Sprintf("%-7s - %-8s\n", module.GetName(), module.GetResult()))
			}
		}

		summary.WriteString(fmt.Sprintf("\nScore: %d\n", m.TotalScore()))

		fmt.Println(summary.String())

	}
}
