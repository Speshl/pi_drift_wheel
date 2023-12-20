package config

type Config struct {
	AppCfg               AppConfig
	ControllerManagerCfg ControllerManagerConfig
	SbusCfg              SBusConfig
}

type AppConfig struct {
	UpdateRate int
}

type ControllerManagerConfig struct{}

type SBusConfig struct {
	SBusPath string
	SBusIn   bool
	SBusOut  bool
}
