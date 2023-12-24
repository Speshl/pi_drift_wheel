package controllers

func GetDIYHandBrakeKeyMap() map[string]Mapping {
	keyMap := make(map[string]Mapping, 0)

	keyMap["3:0"] = Mapping{
		CodeName: "ABS_X",
		Type:     3,
		Code:     0,
		RawInput: 2,
		Rests:    "low",
		Min:      -127,
		Max:      127,
		Inverted: false,
	}

	return keyMap
}
