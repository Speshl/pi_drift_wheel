package controllers

import (
	"log/slog"

	sbus "github.com/Speshl/go-sbus"
)

func WheelMixer(inputs []Input, mixState MixState, opts ControllerOptions) (sbus.Frame, MixState) {
	frame := sbus.NewFrame()

	if mixState.IsEmpty() {
		mixState = NewMixState()
		mixState.esc = "forward"
	}

	//ESC Value
	currentState := mixState.esc
	currentGear := 0
	if opts.UseHPattern {
		for i := 4; i <= 10; i++ {
			if inputs[i].Value > inputs[i].Min {
				if i == 10 {
					currentGear = -1
				} else {
					currentGear = i - 3
				}
				break
			}
		}

		slog.Debug("building frame", "inputs", inputs, "gear", currentGear, "esc_state", currentState, "options", opts)

		if getInputChangeAmount(inputs[1]) > getInputChangeAmount(inputs[2]) { //throttle is pressed more than brake
			if currentGear == 0 { //Neutral so keep esc at center value
				frame.Ch[0] = uint16(sbus.MidValue)
			} else if currentGear == -1 { //Reverse
				switch currentState {
				case "forward": //Going to reverse from forward needs to turn on brakes first then reverse to get esc into reverse mode
					value := MapToRange( //Map throttle to bottom half of esc channel
						inputs[1].Value,
						inputs[1].Min,
						inputs[1].Max,
						sbus.MinValue,
						sbus.MidValue,
					)
					frame.Ch[0] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
					if frame.Ch[0] < uint16(sbus.MidValue) {
						mixState.esc = "brake"
					}
				case "brake":
					frame.Ch[0] = uint16(sbus.MidValue) //set mid to get the esc out of brake
					mixState.esc = "reverse"
					slog.Debug("setting esc center to get out of brake and prepare for reverse")
				case "reverse":
					value := MapToRange( //Map throttle to bottom half of esc channel
						inputs[1].Value,
						inputs[1].Min,
						inputs[1].Max,
						sbus.MinValue,
						sbus.MidValue,
					)
					frame.Ch[0] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
				}

			} else if currentGear > 0 && currentGear <= 6 {
				value := int(float64(inputs[1].Value) / float64(6) * float64(currentGear)) //Scale throttle to gear
				if currentGear == 6 {
					value = inputs[1].Value //let top gear have full range without rounding issues
				}

				value = MapToRange(
					value,
					inputs[1].Min,
					inputs[1].Max,
					sbus.MidValue,
					sbus.MaxValue,
				)
				frame.Ch[0] = uint16(value)
				if frame.Ch[0] > uint16(sbus.MidValue) {
					mixState.esc = "forward"
				}
			} else {
				slog.Warn("gear out of bounds")
			}
		} else { //Brake is pressed more than throttle
			switch currentState {
			case "forward":
				value := MapToRangeWithDeadzoneLow(
					inputs[2].Value,
					inputs[2].Min,
					inputs[2].Max,
					sbus.MinValue,
					sbus.MidValue,
					2,
				)
				frame.Ch[0] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
				if frame.Ch[0] < uint16(sbus.MidValue) {
					mixState.esc = "brake"
				}

			case "brake":
				value := MapToRangeWithDeadzoneLow(
					inputs[2].Value,
					inputs[2].Min,
					inputs[2].Max,
					sbus.MinValue,
					sbus.MidValue,
					2,
				)
				frame.Ch[0] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
				if frame.Ch[0] > uint16(sbus.MidValue-10) {                 //brakes not/or barely pushed
					frame.Ch[0] = uint16(sbus.MidValue + 10) //set enough forward keep esc out of reverse
					mixState.esc = "forward"
					slog.Debug("keeping brakes from going to reverse, by setting slightly forward")
				}

			case "reverse":
				frame.Ch[0] = uint16(sbus.MidValue + 10) //set enough forward to get the esc out of reverse
				mixState.esc = "forward"
				slog.Debug("getting esc out of reverse before pressing the brakes")
			}

		}
	} else { //map without using gear selections
		if getInputChangeAmount(inputs[1]) > getInputChangeAmount(inputs[2]) { //throttle is pressed more than brake
			frame.Ch[0] = uint16(MapToRangeWithDeadzoneLow(
				inputs[1].Value,
				inputs[1].Min,
				inputs[1].Max,
				sbus.MidValue,
				sbus.MaxValue,
				2,
			))
		} else { //brake pressed more or equal to throttle
			value := MapToRangeWithDeadzoneLow(
				inputs[2].Value,
				inputs[2].Min,
				inputs[2].Max,
				sbus.MinValue,
				sbus.MidValue,
				2,
			)
			frame.Ch[0] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
		}
	}

	//Handle ESC State

	//Steer Value
	frame.Ch[1] = uint16(MapToRangeWithDeadzoneMid(
		inputs[0].Value,
		inputs[0].Min,
		inputs[0].Max,
		sbus.MinValue,
		sbus.MaxValue,
		2,
	))

	slog.Debug("mixed frame", "gear", currentGear, "esc_state", currentState, "esc", frame.Ch[0], "steer", frame.Ch[1])

	return frame, mixState
}
