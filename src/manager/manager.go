package manager

import (
	"checker-pa/src/checker-modules"
	"checker-pa/src/utils"
	"errors"
	"fmt"
	"log"
	"os"
)

type Manager struct {
	Modules []checkermodules.CheckerModule
}

func NewManager() (*Manager, error) {
	var m Manager

	err := m.registerModules()
	if err != nil {
		return nil, err
	}

	m.RetrieveConfig()

	return &m, nil
}

func (m *Manager) register(module checkermodules.CheckerModule) {
	m.Modules = append(m.Modules, module)
}

func (m *Manager) registerModules() error {
	if utils.Config.ModuleConfig.RefChecker != nil && checkermodules.AvailableModules["ref_checker"] == nil {
		return errors.New("ref_checker not available")
	}

	if utils.Config.ModuleConfig.MemoryChecker != nil && checkermodules.AvailableModules["memory_checker"] == nil {
		return errors.New("memory_checker not available")
	}

	if utils.Config.ModuleConfig.StyleChecker != nil && checkermodules.AvailableModules["style_checker"] == nil {
		return errors.New("style_checker not available")
	}

	if utils.Config.ModuleConfig.CommitChecker != nil && checkermodules.AvailableModules["commit_checker"] == nil {
		return errors.New("commit_checker not available")
	}

	for _, module := range checkermodules.AvailableModules {
		m.register(module)
	}

	return nil
}

func checkDependencies(module checkermodules.CheckerModule, finished map[string]bool) bool {
	if len(module.WaitingFor()) == 0 {
		return true
	}

	for _, dependency := range module.WaitingFor() {
		if !finished[dependency] {
			return false
		}
	}

	return true
}

func runDeferred(deferred []checkermodules.CheckerModule, finished map[string]bool) {
	for i, deferredModule := range deferred {
		if checkDependencies(deferredModule, finished) {
			deferred = append(deferred[:i], deferred[i+1:]...)
			fmt.Println("Running " + deferredModule.GetName() + " module")
			deferredModule.Run()
			finished[deferredModule.GetName()] = true
		}
	}
}

func (m *Manager) RetrieveConfig() {
	if _, err := os.Stat(utils.UserConfigPath); err == nil {
		// Read the config from there
		data, err := os.ReadFile(utils.UserConfigPath)
		if err != nil {
			panic(err)
		}

		utils.Config.UserConfig, err = utils.NewUserConfig(string(data))
		if err != nil {
			panic(err)
		}
	} else {
		f, err := os.Create(utils.UserConfigPath)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		_, err = f.WriteString(utils.Config.DefaultUserConfig)
		if err != nil {
			panic(err)
		}

	}
}

func (m *Manager) Run() {
	var finished = make(map[string]bool)
	var deferred []checkermodules.CheckerModule

	for _, module := range m.Modules {
		// If the current module doesn't need to wait for another, just run it
		if checkDependencies(module, finished) {
			fmt.Println("Running " + module.GetName() + " module")
			module.Run()
			finished[module.GetName()] = true
		} else {
			deferred = append(deferred, module)
		}

		// Search for deferred Modules to run
		runDeferred(deferred, finished)
	}

	if len(deferred) == 0 {
		return
	}

	// Check for remaining deferred (lazy cycle check)
	var maxIterations = len(deferred)
	cycle := true

	for i := 0; i < maxIterations; i++ {
		if len(deferred) == 0 {
			cycle = false
			break
		}
		// Search for deferred Modules to run
		runDeferred(deferred, finished)
	}

	if cycle {
		log.Fatal("Module dependency cycle detected")
	}

}

func (m *Manager) TotalScore() int {
	var total int
	for _, module := range m.Modules {
		total += module.Score()
	}

	return total
}
