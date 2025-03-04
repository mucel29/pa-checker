package utils

import (
	"log/slog"
	"os"
)

var logFile *os.File
var logger *slog.Logger

var Config struct {
	*ModuleConfig
	*UserConfig
}

func InitConfig(defaultUserConfigStr string, moduleConfigStr string) {
	var err error

	Config.UserConfig, err = newUserConfig(defaultUserConfigStr)
	if err != nil {
		panic(err)
	}

	Config.ModuleConfig, err = newModuleConfig(moduleConfigStr)
	if err != nil {
		panic(err)
	}

	logFile, err = os.Create("./log.txt")
	if err != nil {
		panic(err)
	}

	logger = slog.New(slog.NewTextHandler(logFile, nil))

}

func Log(str string) {
	logger.Info(str)
}
