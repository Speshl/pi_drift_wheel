package config

type Config struct {
	AppCfg               AppConfig
	ControllerManagerCfg ControllerManagerConfig
	SbusCfg              SBusConfig
}

type AppConfig struct{}

type ControllerManagerConfig struct{}

type SBusConfig struct {
	SBusInPath string
	SBusInBaud int
}
