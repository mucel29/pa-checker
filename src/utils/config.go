package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var UserConfigPath = "config.json"
var logFile *os.File
var logger *slog.Logger

var ProjectPath string

var Config struct {
	*ModuleConfig
	*UserConfig

	DefaultUserConfig string
}

func InitConfig(defaultUserConfigStr string, projectPath string) error {
	var err error

	ProjectPath = projectPath
	absPath, err := filepath.Abs(ProjectPath)
	if err == nil {
		ProjectPath = absPath
	}

	UserConfigPath = filepath.Join(ProjectPath, "config.json")

	logFile, err = os.Create(filepath.Join(ProjectPath, "checker_log.txt"))
	if err != nil {
		return err
	}

	logger = slog.New(slog.NewTextHandler(logFile, nil))

	Config.UserConfig, err = NewUserConfig(defaultUserConfigStr)
	if err != nil {
		return err
	}

	moduleConfigBytes, err := os.ReadFile(Abs("tests.json"))
	if err != nil {
		return fmt.Errorf("failed to read tests.json: %w", err)
	}

	Config.ModuleConfig, err = newModuleConfig(string(moduleConfigBytes))
	if err != nil {
		return fmt.Errorf("failed to parse tests.json: %w", err)
	}

	Config.DefaultUserConfig = defaultUserConfigStr

	return nil
}

func SaveUserConfig() {
	f, err := os.Create(UserConfigPath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	newData, err := json.MarshalIndent(Config.UserConfig, "", "	")
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(string(newData))
	if err != nil {
		panic(err)
	}
}

func Log(str string) {
	logger.Info(str)
}

func Err(str string) {
	logger.Error(str)
}

func Fatal(str string) {
	fmt.Println(str)
	logger.Error(str)
	os.Exit(1)
}

var ConfigMacros = make(map[string]string)

func convertMacros(srcStr string, contextMacros map[string]string) string {
	// Replace implicit macros
	for k, v := range ConfigMacros {
		srcStr = strings.ReplaceAll(srcStr, fmt.Sprintf("$%s", k), v)
	}

	// Replace context macros
	for k, v := range contextMacros {
		srcStr = strings.ReplaceAll(srcStr, fmt.Sprintf("$%s", k), v)
	}

	return srcStr
}

// ExpandMacros No cyclic macros please!
func ExpandMacros(str string, contextMacros map[string]string) string {
	lastStr := ""

	for lastStr != str {
		lastStr = str
		str = convertMacros(str, contextMacros)
	}

	return str
}

func Abs(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	absPath, _ := filepath.Abs(filepath.Join(ProjectPath, rel))

	return absPath
}
