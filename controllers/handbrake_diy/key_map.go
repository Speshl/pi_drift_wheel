package handbrake_diy

import (
	"github.com/Speshl/pi_drift_wheel/controllers/models"
)

func GetKeyMap() map[string]models.Mapping {
	keyMap := make(map[string]models.Mapping, 0)

	keyMap["3:0"] = models.Mapping{
		CodeName: "ABS_X",
		Type:     3,
		Code:     0,
		RawInput: 2, //brake
		Rests:    "low",
		Min:      -127,
		Max:      127,
		Inverted: false,
	}

	return keyMap
}
