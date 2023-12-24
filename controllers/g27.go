package controllers

func GetG27KeyMap() map[string]Mapping {
	keyMap := make(map[string]Mapping, 3)

	keyMap["3:0"] = Mapping{ //steer
		CodeName: "ABS_X",
		Type:     3,
		Code:     0,
		RawInput: 0,
		Min:      0,
		Max:      16383,
		Rests:    "middle",
		Inverted: false,
	}

	keyMap["3:2"] = Mapping{ //throttle
		CodeName: "ABS_Z",
		Type:     3,
		Code:     2,
		RawInput: 1,
		Min:      0,
		Max:      255,
		Rests:    "low",
		Inverted: true,
	}

	keyMap["3:5"] = Mapping{ //brake
		CodeName: "ABS_RZ",
		Type:     3,
		Code:     5,
		RawInput: 2,
		Min:      0,
		Max:      255,
		Rests:    "low",
		Inverted: true,
	}

	keyMap["3:1"] = Mapping{ //clutch
		CodeName: "ABS_X",
		Type:     3,
		Code:     1,
		RawInput: 3,
		Min:      0,
		Max:      255,
		Rests:    "low",
		Inverted: true,
	}

	keyMap["1:300"] = Mapping{ //first
		CodeName: "300",
		Type:     1,
		Code:     303,
		RawInput: 4,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:301"] = Mapping{ //second
		CodeName: "301",
		Type:     1,
		Code:     303,
		RawInput: 5,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:302"] = Mapping{ //third
		CodeName: "302",
		RawInput: 6,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:303"] = Mapping{ //fourth
		CodeName: "unknown",
		Type:     1,
		Code:     303,
		RawInput: 7,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:704"] = Mapping{ //fifth
		CodeName: "704",
		Type:     1,
		Code:     704,
		RawInput: 8,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:705"] = Mapping{ //sixth
		CodeName: "705",
		Type:     1,
		Code:     704,
		RawInput: 9,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:710"] = Mapping{ //reverse
		CodeName: "710",
		Type:     1,
		Code:     710,
		RawInput: 10,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	return keyMap
}
