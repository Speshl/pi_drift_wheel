package controllers

/* Raw Input Mapping */
//Values 0-9 are expected to be axis. (0 steer, 1 throttle, 2 brake, 3 clutch (not used), 4-9 unused)
//Values 10-19 are expected to be H pattern gears. (10 first, 11 second, 12 third ... 19 reverse) only 6 gears used atm
//Values 20-21 are paddle shifts
//Values 32-64 are all other buttons
//
/* Raw Input Mapping */

func GetG27KeyMap() map[string]Mapping {
	keyMap := make(map[string]Mapping, 3)

	keyMap["3:0"] = Mapping{
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

	keyMap["3:2"] = Mapping{ //throttle
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

	keyMap["3:5"] = Mapping{ //brake
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

	keyMap["3:1"] = Mapping{ //clutch
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

	keyMap["1:300"] = Mapping{ //first
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

	keyMap["1:301"] = Mapping{ //second
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

	keyMap["1:302"] = Mapping{ //third
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

	keyMap["1:303"] = Mapping{ //fourth
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

	keyMap["1:704"] = Mapping{ //fifth
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

	keyMap["1:705"] = Mapping{ //sixth
		Label:    "6th",
		CodeName: "705",
		Type:     1,
		Code:     704,
		RawInput: 15,
		Min:      0,
		Max:      1,
		Rests:    "low",
		Inverted: false,
	}

	keyMap["1:710"] = Mapping{ //reverse
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
	// keyMap["1:710"] = Mapping{ //reverse
	// 	Label:    "upshift",
	// 	CodeName: "710",
	// 	Type:     1,
	// 	Code:     710,
	// 	RawInput: 20,
	// 	Min:      0,
	// 	Max:      1,
	// 	Rests:    "low",
	// 	Inverted: false,
	// }

	// keyMap["1:710"] = Mapping{ //reverse
	// 	Label:    "downshift",
	// 	CodeName: "710",
	// 	Type:     1,
	// 	Code:     710,
	// 	RawInput: 20,
	// 	Min:      0,
	// 	Max:      1,
	// 	Rests:    "low",
	// 	Inverted: false,
	// }

	// //left Face buttons
	// keyMap["1:710"] = Mapping{ //reverse
	// 	Label:    "top_left",
	// 	CodeName: "710",
	// 	Type:     1,
	// 	Code:     710,
	// 	RawInput: 32,
	// 	Min:      0,
	// 	Max:      1,
	// 	Rests:    "low",
	// 	Inverted: false,
	// }
	// keyMap["1:710"] = Mapping{ //reverse
	// 	Label:    "mid_left",
	// 	CodeName: "710",
	// 	Type:     1,
	// 	Code:     710,
	// 	RawInput: 33,
	// 	Min:      0,
	// 	Max:      1,
	// 	Rests:    "low",
	// 	Inverted: false,
	// }
	// keyMap["1:710"] = Mapping{ //reverse
	// 	Label:    "bot_left",
	// 	CodeName: "710",
	// 	Type:     1,
	// 	Code:     710,
	// 	RawInput: 34,
	// 	Min:      0,
	// 	Max:      1,
	// 	Rests:    "low",
	// 	Inverted: false,
	// }

	// //Right Face buttons
	// keyMap["1:710"] = Mapping{ //reverse
	// 	Label:    "top_right",
	// 	CodeName: "710",
	// 	Type:     1,
	// 	Code:     710,
	// 	RawInput: 35,
	// 	Min:      0,
	// 	Max:      1,
	// 	Rests:    "low",
	// 	Inverted: false,
	// }
	// keyMap["1:710"] = Mapping{ //reverse
	// 	Label:    "mid_right",
	// 	CodeName: "710",
	// 	Type:     1,
	// 	Code:     710,
	// 	RawInput: 36,
	// 	Min:      0,
	// 	Max:      1,
	// 	Rests:    "low",
	// 	Inverted: false,
	// }
	// keyMap["1:710"] = Mapping{ //reverse
	// 	Label:    "bot_right",
	// 	CodeName: "710",
	// 	Type:     1,
	// 	Code:     710,
	// 	RawInput: 37,
	// 	Min:      0,
	// 	Max:      1,
	// 	Rests:    "low",
	// 	Inverted: false,
	// }

	return keyMap
}
