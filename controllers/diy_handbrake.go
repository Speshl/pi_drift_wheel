package controllers

func GetDIYHandBrakeKeyMap() map[string]Mapping {
	keyMap := make(map[string]Mapping, 0)

	keyMap["ABS_X"] = Mapping{
		CodeName: "ABS_X",
		Channel:  1,
		RawInput: 0,
		Type:     "axis_bottom",
		Min:      -127,
		Max:      127,
		Inverted: false,
	}

	return keyMap
}
