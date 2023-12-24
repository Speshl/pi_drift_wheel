package controllers

func GetG27KeyMap() map[string]Mapping {
	keyMap := make(map[string]Mapping, 3)

	keyMap["ABS_X"] = Mapping{ //steer
		CodeName: "ABS_X",
		Channel:  0,
		RawInput: 0,
		Type:     "axis_full",
		Min:      0,
		Max:      16383,
		Inverted: false,
	}

	keyMap["ABS_Z"] = Mapping{ //throttle
		CodeName: "ABS_Z",
		Channel:  1,
		RawInput: 1,
		Type:     "axis_top",
		Min:      0,
		Max:      255,
		Inverted: true,
	}

	keyMap["ABS_RZ"] = Mapping{ //brake
		CodeName: "ABS_RZ",
		Channel:  1,
		RawInput: 2,
		Type:     "axis_bottom",
		Min:      0,
		Max:      255,
		Inverted: true,
	}

	// keyMap["ABS_RZ"] = Mapping{ //clutch
	// 	CodeName: "ABS_RZ",
	// 	Channel:  -1,
	// 	RawInput: 3,
	// 	Type:     "axis_bottom",
	// 	Min:      0,
	// 	Max:      255,
	// 	Inverted: true,
	// }

	// keyMap["ABS_RZ"] = Mapping{ //first
	// 	CodeName: "ABS_RZ",
	// 	Channel:  -1,
	// 	RawInput: 4,
	// 	Type:     "button_gear",
	// 	Min:      0,
	// 	Max:      1,
	// 	Inverted: false,
	// }

	// keyMap["ABS_RZ"] = Mapping{ //second
	// 	CodeName: "ABS_RZ",
	// 	Channel:  -1,
	// 	RawInput: 5,
	// 	Type:     "button_gear",
	// 	Min:      0,
	// 	Max:      1,
	// 	Inverted: false,
	// }

	// keyMap["ABS_RZ"] = Mapping{ //third
	// 	CodeName: "ABS_RZ",
	// 	Channel:  -1,
	// 	RawInput: 6,
	// 	Type:     "button_gear",
	// 	Min:      0,
	// 	Max:      1,
	// 	Inverted: false,
	// }

	// keyMap["ABS_RZ"] = Mapping{ //fourth
	// 	CodeName: "ABS_RZ",
	// 	Channel:  -1,
	// 	RawInput: 7,
	// 	Type:     "button_gear",
	// 	Min:      0,
	// 	Max:      1,
	// 	Inverted: false,
	// }

	// keyMap["ABS_RZ"] = Mapping{ //fifth
	// 	CodeName: "ABS_RZ",
	// 	Channel:  -1,
	// 	RawInput: 8,
	// 	Type:     "button_gear",
	// 	Min:      0,
	// 	Max:      1,
	// 	Inverted: false,
	// }

	// keyMap["ABS_RZ"] = Mapping{ //sixth
	// 	CodeName: "ABS_RZ",
	// 	Channel:  -1,
	// 	RawInput: 9,
	// 	Type:     "button_gear",
	// 	Min:      0,
	// 	Max:      1,
	// 	Inverted: false,
	// }

	// keyMap["ABS_RZ"] = Mapping{ //reverse
	// 	CodeName: "ABS_RZ",
	// 	Channel:  -1,
	// 	RawInput: 10,
	// 	Type:     "button_gear",
	// 	Min:      0,
	// 	Max:      1,
	// 	Inverted: false,
	// }

	return keyMap
}
