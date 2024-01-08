package config

import (
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
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
	channelString := GetStringEnv(fmt.Sprintf("%d_SBUSCHANNELS", portNum), DefaultSBusChannels[portNum])
	splitChannels := strings.Split(channelString, ",")
	intChannels := make([]int, 0, 16)
	for i := range splitChannels {
		channel, err := strconv.Atoi(splitChannels[i])
		if err != nil {
			slog.Error("failed parsing sbus channels", "port", i, "entry", channelString)
			break
		}
		intChannels = append(intChannels, channel)
	}

	return SBusConfig{
		SBusPath:     GetStringEnv(fmt.Sprintf("%d_SBUSPATH", portNum), DefaultSBusPaths[portNum]),
		SBusType:     GetStringEnv(fmt.Sprintf("%d_SBUSTYPE", portNum), DefaultSBusTypes[portNum]),
		SBusRx:       GetBoolEnv(fmt.Sprintf("%d_SBUSRX", portNum), DefaultSBusRx[portNum]),
		SBusTx:       GetBoolEnv(fmt.Sprintf("%d_SBUSTX", portNum), DefaultSBusTx[portNum]),
		SBusChannels: intChannels,
	}
}
