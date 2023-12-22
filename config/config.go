package config

import (
	"fmt"
	"log"
)

func GetConfig() Config {
	cfg := Config{
		AppCfg:               GetAppConfig(),
		ControllerManagerCfg: GetControllerManagerConfig(),
		SbusCfgs:             GetSBusConfigs(),
	}

	log.Printf("app config: \n%+v\n", cfg)
	return cfg
}

func GetAppConfig() AppConfig {
	return AppConfig{
		UpdateRate: AppUpdateRate, // value in milliseconds
	}
}

func GetControllerManagerConfig() ControllerManagerConfig {
	return ControllerManagerConfig{}
}

func GetSBusConfigs() []SBusConfig {
	sbusCfgs := make([]SBusConfig, 0, MaxSbus)
	for i := 0; i < MaxSbus; i++ {
		sbusCfgs = append(sbusCfgs, GetSBusConfig(i))
	}
	return sbusCfgs
}

func GetSBusConfig(portNum int) SBusConfig {
	return SBusConfig{
		SBusPath: GetStringEnv(fmt.Sprintf("%d_SBUSPATH", portNum), DefaultSBusPaths[portNum]),
		SBusType: GetStringEnv(fmt.Sprintf("%d_SBUSTYPE", portNum), DefaultSBusTypes[portNum]),
		SBusRx:   GetBoolEnv(fmt.Sprintf("%d_SBUSRX", portNum), DefaultSBusRx[portNum]),
		SBusTx:   GetBoolEnv(fmt.Sprintf("%d_SBUSTX", portNum), DefaultSBusTx[portNum]),
	}
}
