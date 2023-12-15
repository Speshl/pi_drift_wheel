package controllers

func GetG27KeyMap() map[string]Mapping {
	keyMap := make(map[string]Mapping, 3)

	keyMap["ABS_X"] = Mapping{
		CodeName: "ABS_X",
		Channel:  0,
		Type:     "axis_full",
		Min:      0,
		Max:      16383,
		Inverted: false,
	}

	keyMap["ABS_Z"] = Mapping{
		CodeName: "ABS_Z",
		Channel:  1,
		Type:     "axis_top",
		Min:      0,
		Max:      255,
		Inverted: true,
	}

	keyMap["ABS_RZ"] = Mapping{
		CodeName: "ABS_RZ",
		Channel:  1,
		Type:     "axis_bottom",
		Min:      0,
		Max:      255,
		Inverted: true,
	}

	return keyMap
}

func GetDIYHandBrakeKeyMap() map[string]Mapping {
	keyMap := make(map[string]Mapping, 0)
	return keyMap
}
