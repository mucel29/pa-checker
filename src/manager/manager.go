package manager

import (
	"checker-pa/src/checker-modules"
	"checker-pa/src/utils"
	"errors"
	"log"
)

type Manager struct {
	Modules []checker_modules.CheckerModule
}

func NewManager() (*Manager, error) {
	var m Manager

	err := m.registerModules()
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Manager) register(module checker_modules.CheckerModule) {
	m.Modules = append(m.Modules, module)
}

func (m *Manager) registerModules() error {
	if utils.Config.ModuleConfig.RefChecker != nil && checker_modules.AvailableModules["ref_checker"] == nil {
		return errors.New("ref_checker not available")
	}

	if utils.Config.ModuleConfig.MemoryChecker != nil && checker_modules.AvailableModules["memory_checker"] == nil {
		return errors.New("memory_checker not available")
	}

	if utils.Config.ModuleConfig.StyleChecker != nil && checker_modules.AvailableModules["style_checker"] == nil {
		return errors.New("style_checker not available")
	}

	if utils.Config.ModuleConfig.CommitChecker != nil && checker_modules.AvailableModules["commit_checker"] == nil {
		return errors.New("commit_checker not available")
	}

	for _, module := range checker_modules.AvailableModules {
		m.register(module)
	}

	return nil
}

func checkDependencies(module checker_modules.CheckerModule, finished map[string]bool) bool {
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

func runDeferred(deferred []checker_modules.CheckerModule, finished map[string]bool) {
	for i, deferredModule := range deferred {
		if checkDependencies(deferredModule, finished) {
			deferred = append(deferred[:i], deferred[i+1:]...)
			deferredModule.Run()
			finished[deferredModule.GetName()] = true
		}
	}
}

func (m *Manager) Run() {
	var finished = make(map[string]bool)
	var deferred []checker_modules.CheckerModule

	for _, module := range m.Modules {
		// If the current module doesn't need to wait for another, just run it
		if checkDependencies(module, finished) {
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
