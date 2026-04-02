package main

import (
	"checker-pa/src/display"
	"checker-pa/src/manager"
	"checker-pa/src/menu"
	"checker-pa/src/utils"
	_ "embed"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

//go:embed res/config/user_config.json
var defaultUserConfigStr string

var useInteractive bool
var projectPath string

func init() {
	flag.BoolVar(&useInteractive, "i", false, "Interactive mode")
	flag.StringVar(&projectPath, "path", ".", "Project path")
}

func main() {
	flag.Parse()

	err := utils.InitConfig(defaultUserConfigStr, projectPath)
	if err != nil {
		utils.Fatal("FATAL ERROR DETECTED! " + err.Error() + "\n ABORTING!")
	}

	m, err := manager.NewManager()

	if err != nil {
		utils.Fatal("FATAL ERROR DETECTED! " + err.Error() + "\n ABORTING!")
	}
	defer m.CleanUp()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		<-sigChan
		utils.Log("Received interrupt/termination signal. Cleaning up...")
		if m != nil {
			m.CleanUp()
		}
		os.Exit(1)
	}()

	// TODO?: make this return an error as well?
	// m.Check()

	if useInteractive {

		utils.Log("Interactive Display")
		d := display.NewDisplay()
		d.OnStop = func() {
			m.CleanUp()
		}

		go func() {
			err := m.Run()
			if err != nil {
				d.App.Stop()
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

		m.BasicSummary("")

	}
}
