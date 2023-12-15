package config

import (
	"log"
)

func GetConfig() Config {
	cfg := Config{
		AppCfg:               GetAppConfig(),
		ControllerManagerCfg: GetControllerManagerConfig(),
	}

	log.Printf("app config: \n%+v\n", cfg)
	return cfg
}

func GetAppConfig() AppConfig {
	return AppConfig{}
}

func GetControllerManagerConfig() ControllerManagerConfig {
	return ControllerManagerConfig{}
}
