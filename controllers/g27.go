package controllers

import (
	"log/slog"

	"github.com/Speshl/pi_drift_wheel/sbus"
)

func G27Mixer(inputs []int, mixState map[string]string, opts ControllerOptions) (sbus.Frame, map[string]string) {
	frame := sbus.NewFrame()

	if mixState == nil {
		mixState = make(map[string]string, 1)
		mixState["esc_state"] = "forward"
	}

	//ESC Value

	if opts.useHPattern {
		currentGear := 0
		for i := 4; i <= 10; i++ {
			if inputs[i] > sbus.MidValue {
				if i == 10 {
					currentGear = -1
				} else {
					currentGear = i - 3
				}
				break
			}
		}

		slog.Info("building frame", "inputs", inputs, "state", mixState, "options", opts, "gear", currentGear)
		if currentGear == 0 { //Neutral so keep esc at center value
			frame.Ch[0] = uint16(sbus.MidValue)
		} else if currentGear == -1 { //Reverse
			value := MapToRange( //Map throttle to bottom half of esc channel
				inputs[1],
				sbus.MinValue,
				sbus.MaxValue,
				sbus.MinValue,
				sbus.MaxValue,
			)
			frame.Ch[0] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
		} else if currentGear > 0 && currentGear <= 6 {
			value := int(float64(inputs[1]) / float64(6) * float64(currentGear)) //Scale throttle to gear
			if currentGear == 6 {
				value = inputs[6] //let top gear have full range without rounding issues
			}

			value = MapToRange( //Map throttle to bottom half of esc channel
				value,
				sbus.MinValue,
				sbus.MaxValue,
				sbus.MidValue,
				sbus.MaxValue,
			)
			frame.Ch[0] = uint16(value)
		} else {
			slog.Warn("gear out of bounds")
		}

	} else { //map without using gear selections
		if inputs[1] > inputs[2] { //throttle is pressed more than brake
			frame.Ch[0] = uint16(MapToRange(
				inputs[1],
				sbus.MinValue,
				sbus.MaxValue,
				sbus.MidValue,
				sbus.MaxValue,
			))
		} else { //brake pressed more or equal to throttle
			value := MapToRange(
				inputs[2],
				sbus.MinValue,
				sbus.MaxValue,
				sbus.MinValue,
				sbus.MidValue,
			)
			frame.Ch[0] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
		}
	}

	//Handle ESC State

	//Steer Value
	frame.Ch[1] = uint16(inputs[0])

	return frame, mixState
}

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
