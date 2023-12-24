package controllers

func GetG27KeyMap() map[string]Mapping {
	keyMap := make(map[string]Mapping, 3)

	keyMap["3:0"] = Mapping{ //steer
		CodeName: "ABS_X",
		Type:     3,
		Code:     0,
		Channel:  0,
		RawInput: 0,
		MapType:  "axis_full",
		Min:      0,
		Max:      16383,
		Inverted: false,
	}

	keyMap["3:2"] = Mapping{ //throttle
		CodeName: "ABS_Z",
		Type:     3,
		Code:     2,
		Channel:  1,
		RawInput: 1,
		MapType:  "axis_top",
		Min:      0,
		Max:      255,
		Inverted: true,
	}

	keyMap["3:5"] = Mapping{ //brake
		CodeName: "ABS_RZ",
		Type:     3,
		Code:     5,
		Channel:  1,
		RawInput: 2,
		MapType:  "axis_bottom",
		Min:      0,
		Max:      255,
		Inverted: true,
	}

	keyMap["3:1"] = Mapping{ //clutch
		CodeName: "ABS_X",
		Type:     3,
		Code:     1,
		Channel:  -1,
		RawInput: 3,
		MapType:  "axis_bottom",
		Min:      0,
		Max:      255,
		Inverted: true,
	}

	keyMap["1:300"] = Mapping{ //first
		CodeName: "300",
		Type:     1,
		Code:     303,
		Channel:  -1,
		RawInput: 4,
		MapType:  "button_gear",
		Min:      0,
		Max:      1,
		Inverted: false,
	}

	keyMap["1:301"] = Mapping{ //second
		CodeName: "301",
		Type:     1,
		Code:     303,
		Channel:  -1,
		RawInput: 5,
		MapType:  "button_gear",
		Min:      0,
		Max:      1,
		Inverted: false,
	}

	keyMap["1:302"] = Mapping{ //third
		CodeName: "302",
		Channel:  -1,
		RawInput: 6,
		MapType:  "button_gear",
		Min:      0,
		Max:      1,
		Inverted: false,
	}

	keyMap["1:303"] = Mapping{ //fourth
		CodeName: "unknown",
		Type:     1,
		Code:     303,
		Channel:  -1,
		RawInput: 7,
		MapType:  "button_gear",
		Min:      0,
		Max:      1,
		Inverted: false,
	}

	keyMap["1:704"] = Mapping{ //fifth
		CodeName: "704",
		Channel:  -1,
		RawInput: 8,
		MapType:  "button_gear",
		Min:      0,
		Max:      1,
		Inverted: false,
	}

	keyMap["1:705"] = Mapping{ //sixth
		CodeName: "705",
		Channel:  -1,
		RawInput: 9,
		MapType:  "button_gear",
		Min:      0,
		Max:      1,
		Inverted: false,
	}

	keyMap["1:710"] = Mapping{ //reverse
		CodeName: "710",
		Type:     1,
		Code:     710,
		Channel:  -1,
		RawInput: 10,
		MapType:  "button_gear",
		Min:      0,
		Max:      1,
		Inverted: false,
	}

	return keyMap
}
