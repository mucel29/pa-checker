package utils

import (
	"encoding/json"
	"log/slog"
	"os"
)

const (
	UserConfigPath = "./config.json"
)

var logFile *os.File
var logger *slog.Logger

var Config struct {
	*ModuleConfig
	*UserConfig

	DefaultUserConfig string
}

func InitConfig(defaultUserConfigStr string, moduleConfigStr string) {
	var err error

	Config.UserConfig, err = NewUserConfig(defaultUserConfigStr)
	if err != nil {
		panic(err)
	}

	Config.ModuleConfig, err = newModuleConfig(moduleConfigStr)
	if err != nil {
		panic(err)
	}

	Config.DefaultUserConfig = defaultUserConfigStr

	logFile, err = os.Create("./log.txt")
	if err != nil {
		panic(err)
	}

	logger = slog.New(slog.NewTextHandler(logFile, nil))

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
