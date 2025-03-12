package manager

import (
	"bytes"
	"checker-pa/src/checker-modules"
	"checker-pa/src/utils"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// SystemCapabilities TODO: make it a map in the future
type SystemCapabilities int

const (
	valgrind SystemCapabilities = iota
	cppcheck
	end
)

type Manager struct {
	Modules []checkermodules.CheckerModule

	capabilities []bool
}

func (m *Manager) checkCapabilities() {

	// Check for Valgrind

	utils.Log("Checking capabilities...")

	if _, err := exec.LookPath("valgrind"); err != nil {
		utils.Log("[ERR] valgrind")
	} else {
		utils.Log("[OK] valgrind")
		m.capabilities[valgrind] = true
	}

	// Check for cppcheck
	if _, err := exec.LookPath("cppcheck"); err != nil {
		utils.Log("[ERR] cppcheck")
	} else {
		utils.Log("[OK] cppcheck")
		m.capabilities[cppcheck] = true
	}

}

func NewManager() (*Manager, error) {
	var m Manager

	m.capabilities = make([]bool, end)

	m.checkCapabilities()
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
			utils.Log("Running " + deferredModule.GetName() + " module")
			deferredModule.Run()
			finished[deferredModule.GetName()] = true
		}
	}
}

func updateMacros() {
	// Output path
	outPath, err := filepath.Abs(utils.Config.OutputPath)
	if err == nil {
		utils.ConfigMacros["OUT_DIR"] = outPath
	}

	// Make sure input path exists
	inPath, err := filepath.Abs(utils.Config.InputPath)
	if err == nil {
		if _, err := os.Stat(inPath); err == nil {
			utils.ConfigMacros["IN_DIR"] = inPath
		}
	}

	srcPath, err := filepath.Abs(utils.Config.SourcePath)
	if err == nil {
		if _, err := os.Stat(srcPath); err == nil {
			utils.ConfigMacros["SRC_DIR"] = srcPath
		}
	}

	// Load module config macros
	for k, v := range utils.Config.Macros {
		utils.ConfigMacros[k] = v
	}

}

func (m *Manager) RetrieveConfig() {
	defer updateMacros()

	if _, err := os.Stat(utils.UserConfigPath); err == nil {
		// Read the config from there
		data, err := os.ReadFile(utils.UserConfigPath)
		if err != nil {
			panic(err)
		}

		// Bug: if fields are not present, they get changed to ""
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

func forwardBytes(bytes bytes.Buffer, filename string) error {

	absForward, err := filepath.Abs(utils.Config.ForwardPath)
	if err != nil {
		return err
	}
	if err := os.Mkdir(absForward, 0777); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	f, err := os.Create(fmt.Sprintf("%s/%s", absForward, filename))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(bytes.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) Run() error {

	if _, err := exec.LookPath(utils.Config.ExecutablePath); err != nil {
		return fmt.Errorf("executable not found: %s", utils.Config.ExecutablePath)
	}

	start := time.Now()

	// Make sure temp path exists
	tempPath, err := filepath.Abs(utils.Config.TempPath)
	if err != nil {
		return err
	}

	if err := os.Mkdir(tempPath, 0777); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	// Make sure output path exists
	if err := os.Mkdir(utils.ConfigMacros["OUT_DIR"], 0777); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	wg := sync.WaitGroup{}

	for _, test := range utils.Config.Tests {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Create Context macros
			contextMacros := map[string]string{
				"FILE": test.File,
				"IN":   fmt.Sprintf("%s/%s.in", utils.ConfigMacros["IN_DIR"], test.File),
				"OUT":  fmt.Sprintf("%s/%s.out", utils.ConfigMacros["OUT_DIR"], test.File),
			}

			var processedArgs []string

			// Process args
			for _, arg := range test.Args {
				processedArgs = append(processedArgs, utils.ExpandMacros(arg, contextMacros))
			}

			var cmd *exec.Cmd

			if m.capabilities[valgrind] && utils.Config.RunValgrind {

				xmlPath := filepath.Join(tempPath, fmt.Sprintf("%s.xml", test.File))

				execPath, err := filepath.Abs(utils.Config.ExecutablePath)
				if err != nil {
					return // err
				}

				valgrindArgs := []string{
					"--leak-check=yes",
					"--xml=yes",
					fmt.Sprintf("--xml-file=%s", xmlPath),
				}

				cmd = exec.Command("valgrind", append(append(valgrindArgs, execPath), processedArgs...)...) //nolint:gosec
				// fmt.Println("running: valgrind " + strings.Join(append(append(valgrindArgs, execPath), processedArgs...), " "))
			} else {
				cmd = exec.Command(utils.Config.ExecutablePath, processedArgs...) //nolint:gosec
			}

			// fmt.Printf("%d: %s %s\n\n", i+1, utils.Config.ExecutablePath, strings.Join(processedArgs, " "))

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			start = time.Now()

			if err := cmd.Run(); err != nil {
				utils.Log("Error running " + test.File)
			}

			// Forward stdout
			if err := forwardBytes(stdout, fmt.Sprintf("%s.stdout", test.File)); err != nil {
				return // err
			}

			// Forward stderr
			if err := forwardBytes(stderr, fmt.Sprintf("%s.stderr", test.File)); err != nil {
				return // err
			}

			utils.Log(fmt.Sprintf("[%s] %s", time.Since(start).String(), test.File))

		}()
	}

	wg.Wait()
	return nil
}

func (m *Manager) Check() {
	var finished = make(map[string]bool)
	var deferred []checkermodules.CheckerModule

	for _, module := range m.Modules {
		// If the current module doesn't need to wait for another, just run it
		if checkDependencies(module, finished) {
			utils.Log("Running " + module.GetName() + " module")
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
		utils.Log("Module dependency cycle detected")
		panic(fmt.Errorf("module dependency cycle detected"))
	}

}

func (m *Manager) TotalScore() int {
	var total int
	for _, module := range m.Modules {
		total += module.Score()
	}

	return total
}
