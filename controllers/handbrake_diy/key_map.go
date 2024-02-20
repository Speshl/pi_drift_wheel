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
		Label:    "handbrake",
	}

	keyMap["3:2"] = models.Mapping{
		Label:    "empty",
		CodeName: "ABS_Z",
		Type:     3,
		Code:     2,
		RawInput: 1,
		Min:      0,
		Max:      255,
		Rests:    "low",
		Inverted: true,
	}

	return keyMap
}
