package utils

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
}
