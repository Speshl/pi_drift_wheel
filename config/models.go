package config

const (
	CRSFBaudRate  = 921600
	MaxSbus       = 2
	MaxCRSF       = 2
	AppUpdateRate = 6
)

var (
	DefaultCRSFPaths = []string{
		"/dev/ttyACM0",
		"",
	}

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

	DefaultSBusChannels = []string{
		"3,4,5",
		"",
	}

	DefaultInvertOutputs = []bool{
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
	}
)

type Config struct {
	AppCfg               AppConfig
	ControllerManagerCfg ControllerManagerConfig
	SbusCfgs             []SBusConfig
	CRSFCfgs             []CRSFConfig
}

type AppConfig struct {
	UpdateRate    int
	InvertOutputs []bool
}

type ControllerManagerConfig struct {
}

type SBusConfig struct {
	SBusPath     string
	SBusType     string
	SBusRx       bool
	SBusTx       bool
	SBusChannels []int
}

type CRSFConfig struct {
	CRSFPath string
}
