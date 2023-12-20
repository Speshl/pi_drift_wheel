package config

import (
	"log"
)

func GetConfig() Config {
	cfg := Config{
		AppCfg:               GetAppConfig(),
		ControllerManagerCfg: GetControllerManagerConfig(),
		SbusCfg:              GetSBusConfig(),
	}

	log.Printf("app config: \n%+v\n", cfg)
	return cfg
}

func GetAppConfig() AppConfig {
	return AppConfig{
		UpdateRate: 5, // value in milliseconds
	}
}

func GetControllerManagerConfig() ControllerManagerConfig {
	return ControllerManagerConfig{}
}

func GetSBusConfig() SBusConfig {
	return SBusConfig{
		SBusPath: "/dev/ttyAMA0",
		SBusIn:   true,
		SBusOut:  true,
	}
}
