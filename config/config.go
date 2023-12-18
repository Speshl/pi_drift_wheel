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
	return AppConfig{}
}

func GetControllerManagerConfig() ControllerManagerConfig {
	return ControllerManagerConfig{}
}

func GetSBusConfig() SBusConfig {
	return SBusConfig{
		SBusInPath: "/dev/ttyAMA0", // /dev/ttyACM0 // /dev/ttyAMA0
		SBusInBaud: 100000,
	}
}
