package config

const (
	MaxSbus       = 2
	AppUpdateRate = 6
)

var (
	DefaultSBusPaths = []string{
		"/dev/ttyAMA0",
		"",
	}

	DefaultSBusTypes = []string{
		"control",
		"telemetry",
	}

	DefaultSBusRx = []bool{
		true,
		false,
	}

	DefaultSBusTx = []bool{
		true,
		false,
	}
)

type Config struct {
	AppCfg               AppConfig
	ControllerManagerCfg ControllerManagerConfig
	SbusCfgs             []SBusConfig
}

type AppConfig struct {
	UpdateRate int
}

type ControllerManagerConfig struct{}

type SBusConfig struct {
	SBusPath string
	SBusType string
	SBusRx   bool
	SBusTx   bool
}
