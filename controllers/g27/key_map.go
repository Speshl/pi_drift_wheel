package g27

import (
	"github.com/Speshl/pi_drift_wheel/controllers/models"
)

/* Raw Input Mapping */
//Values 0-9 are expected to be axis. (0 steer, 1 throttle, 2 brake, 3 clutch (not used), 4-9 unused)
//Values 10-19 are expected to be H pattern gears. (10 first, 11 second, 12 third ... 19 reverse) only 6 gears used atm
//Values 20-21 are paddle shifts
//Values 32-64 are all other buttons
//
/* Raw Input Mapping */

func GetKeyMap() map[string]models.Mapping {
	keyMap := make(map[string]models.Mapping, 3)

	keyMap["3:0"] = models.Mapping{
		Label:    "steer",
		CodeName: "ABS_X",
		Type:     3,
		Code:     0,
		RawInput: 0,
		Min:      0,
		Max:      16383,
		Rests:    "middle",
		Inverted: false,
	}

	keyMap["3:2"] = models.Mapping{
		Label:    "throttle",
		CodeName: "ABS_Z",
		Type:     3,
		Code:     2,
		RawInput: 1,
		Min:      0,
		Max:      255,
		Rests:    "low",
		Inverted: true,
	}

	keyMap["3:5"] = models.Mapping{
		Label:    "brake",
		CodeName: "ABS_RZ",
		Type:     3,
		Code:     5,
		RawInput: 2,
		Min:      0,
		Max:      255,
		Rests:    "low",
		Inverted: true,
	}

	keyMap["3:1"] = models.Mapping{
		Label:    "clutch",
		CodeName: "ABS_X",
		Type:     3,
		Code:     1,
		RawInput: 3,
		Min:      0,
		Max:      255,
		Rests:    "low",
		Inverted: true,
	}

	keyMap["1:300"] = models.Mapping{
		Label:    "1st",
		CodeName: "300",
		Type:     1,
		Code:     300,
		RawInput: 10,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:301"] = models.Mapping{
		Label:    "2nd",
		CodeName: "301",
		Type:     1,
		Code:     301,
		RawInput: 11,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:302"] = models.Mapping{
		Label:    "3rd",
		CodeName: "302",
		Type:     1,
		Code:     302,
		RawInput: 12,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:303"] = models.Mapping{
		Label:    "4th",
		CodeName: "303",
		Type:     1,
		Code:     303,
		RawInput: 13,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:704"] = models.Mapping{
		Label:    "5th",
		CodeName: "704",
		Type:     1,
		Code:     704,
		RawInput: 14,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:705"] = models.Mapping{
		Label:    "6th",
		CodeName: "705",
		Type:     1,
		Code:     705,
		RawInput: 15,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:710"] = models.Mapping{
		Label:    "R",
		CodeName: "710",
		Type:     1,
		Code:     710,
		RawInput: 19,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	//TODO Complete mapping
	keyMap["1:293"] = models.Mapping{
		Label:    "upshift",
		CodeName: "BTN_PINKIE",
		Type:     1,
		Code:     293,
		RawInput: 20,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:294"] = models.Mapping{
		Label:    "downshift",
		CodeName: "710",
		Type:     1,
		Code:     294,
		RawInput: 20,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	//left Face buttons
	keyMap["1:295"] = models.Mapping{
		Label:    "top_left",
		CodeName: "BTN_BASE2",
		Type:     1,
		Code:     295,
		RawInput: 32,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}
	keyMap["1:708"] = models.Mapping{
		Label:    "mid_left",
		CodeName: "BTN_TRIGGER_HAPPY5",
		Type:     1,
		Code:     708,
		RawInput: 33,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}
	keyMap["1:709"] = models.Mapping{
		Label:    "bot_left",
		CodeName: "BTN_TRIGGER_HAPPY6",
		Type:     1,
		Code:     709,
		RawInput: 34,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	//Right Face buttons
	keyMap["1:294"] = models.Mapping{
		Label:    "top_right",
		CodeName: "BTN_BASE",
		Type:     1,
		Code:     294,
		RawInput: 35,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}
	keyMap["1:706"] = models.Mapping{
		Label:    "mid_right",
		CodeName: "BTN_TRIGGER_HAPPY3",
		Type:     1,
		Code:     706,
		RawInput: 36,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}
	keyMap["1:707"] = models.Mapping{
		Label:    "bot_right",
		CodeName: "BTN_TRIGGER_HAPPY4",
		Type:     1,
		Code:     707,
		RawInput: 37,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	//Red Row
	keyMap["1:299"] = models.Mapping{
		Label:    "red1",
		CodeName: "BTN_BASE6",
		Type:     1,
		Code:     299,
		RawInput: 38,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:296"] = models.Mapping{
		Label:    "red2",
		CodeName: "BTN_BASE3",
		Type:     1,
		Code:     296,
		RawInput: 39,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:297"] = models.Mapping{
		Label:    "red3",
		CodeName: "BTN_BASE4",
		Type:     1,
		Code:     297,
		RawInput: 40,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:298"] = models.Mapping{
		Label:    "red4",
		CodeName: "BTN_BASE5",
		Type:     1,
		Code:     298,
		RawInput: 41,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	//D Pad
	keyMap["3:17"] = models.Mapping{
		Label:    "up/down",
		CodeName: "ABS_HAT0Y",
		Type:     3,
		Code:     17,
		RawInput: 42,
		Min:      -1,
		Max:      1,
		Rests:    "mid",
		Inverted: true,
	}

	keyMap["3:16"] = models.Mapping{
		Label:    "left/right",
		CodeName: "ABS_HAT0X",
		Type:     3,
		Code:     16,
		RawInput: 43,
		Min:      -1,
		Max:      1,
		Rests:    "mid",
		Inverted: false,
	}

	//Diamond
	keyMap["1:291"] = models.Mapping{
		Label:    "y/triangle",
		CodeName: "BTN_TOP",
		Type:     1,
		Code:     291,
		RawInput: 44,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:290"] = models.Mapping{
		Label:    "b/circle",
		CodeName: "BTN_THUMB2",
		Type:     1,
		Code:     290,
		RawInput: 45,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:288"] = models.Mapping{
		Label:    "a/x",
		CodeName: "BTN_JOYSTICK/BTN_TRIGGER",
		Type:     1,
		Code:     288,
		RawInput: 45,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:289"] = models.Mapping{
		Label:    "x/square",
		CodeName: "BTN_THUMB",
		Type:     1,
		Code:     289,
		RawInput: 46,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	return keyMap
}
