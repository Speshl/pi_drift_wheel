package controllers

import (
	"log/slog"

	sbus "github.com/Speshl/pi_drift_wheel/sbus"
)

//Mapping - 0 steer, 1 esc, 2 gyro gain

func WheelMixer(inputs []Input, mixState MixState, opts ControllerOptions) (sbus.Frame, MixState) {
	frame := sbus.NewFrame()

	if mixState.IsEmpty() {
		mixState = NewMixState()
		mixState.esc = "forward"
		mixState.gear = 0
	}

	//Check for button state changes
	for i := range inputs {
		if inputs[i].Value == mixState.buttons[inputs[i].Label] {
			if inputs[i].Label == "top_right" {
				slog.Error("top right got here", "input", inputs[i])
			}
			continue
		}

		mixState.buttons[inputs[i].Label] = inputs[i].Value

		if inputs[i].Value != inputs[i].Max { //button presses are considered to be a value equal to the max possible value
			continue //TODO update this to properly handle hats where -1 can be a press also or if something is resting high
		}

		switch inputs[i].Label {
		case "upshift":
			if mixState.gear >= -1 {
				mixState.gear++
			}
		case "downshift":
			if mixState.gear <= 6 {
				mixState.gear--
			}
		case "top_left":
			if mixState.trims["gyro_gain"] > -100 {
				mixState.trims["gyro_gain"]--
			}
		case "top_right":
			if mixState.trims["gyro_gain"] < 100 {
				mixState.trims["gyro_gain"]++
			}
		}
	}

	//Build frame values based on current state/buttons

	//Steer Value
	frame.Ch[0] = uint16(MapToRangeWithDeadzoneMid(
		inputs[0].Value,
		inputs[0].Min,
		inputs[0].Max,
		sbus.MinValue,
		sbus.MaxValue,
		2,
	))

	//ESC Value
	currentState := mixState.esc
	if opts.UseHPattern {
		for i := 10; i < 20; i++ {
			if inputs[i].Value > inputs[i].Min {
				if i == 19 {
					mixState.gear = -1 //reverse
				} else if i > 15 {
					mixState.gear = 0 //set neutral when unsupported gear pressed
					continue
				} else {
					mixState.gear = i - 9 //supported gears
				}
				break //only 1 gear can be active at a time, so stop when one found
			} else if i == 19 {
				mixState.gear = 0 //no gear button pressed, set to neutral
			}
		}

		if getInputChangeAmount(inputs[1]) > getInputChangeAmount(inputs[2]) { //throttle is pressed more than brake
			if mixState.gear == 0 { //Neutral so keep esc at center value
				frame.Ch[1] = uint16(sbus.MidValue)
			} else if mixState.gear == -1 { //Reverse
				switch currentState {
				case "forward": //Going to reverse from forward needs to turn on brakes first then reverse to get esc into reverse mode
					value := MapToRange( //Map throttle to bottom half of esc channel
						inputs[1].Value,
						inputs[1].Min,
						inputs[1].Max,
						sbus.MinValue,
						sbus.MidValue,
					)
					frame.Ch[1] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
					if frame.Ch[1] < uint16(sbus.MidValue) {
						mixState.esc = "brake"
					}
				case "brake":
					frame.Ch[1] = uint16(sbus.MidValue) //set mid to get the esc out of brake
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
					frame.Ch[1] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
				}

			} else if mixState.gear > 0 && mixState.gear <= 6 {
				value := int(float64(inputs[1].Value) / float64(6) * float64(mixState.gear)) //Scale throttle to gear
				if mixState.gear == 6 {
					value = inputs[1].Value //let top gear have full range without rounding issues
				}

				value = MapToRange(
					value,
					inputs[1].Min,
					inputs[1].Max,
					sbus.MidValue,
					sbus.MaxValue,
				)
				frame.Ch[1] = uint16(value)
				if frame.Ch[1] > uint16(sbus.MidValue) {
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
				frame.Ch[1] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
				if frame.Ch[1] < uint16(sbus.MidValue) {
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
				frame.Ch[1] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
				if frame.Ch[1] > uint16(sbus.MidValue-10) {                 //brakes not/or barely pushed
					frame.Ch[1] = uint16(sbus.MidValue + 10) //set enough forward keep esc out of reverse
					mixState.esc = "forward"
					slog.Debug("keeping brakes from going to reverse, by setting slightly forward")
				}

			case "reverse":
				frame.Ch[1] = uint16(sbus.MidValue + 10) //set enough forward to get the esc out of reverse
				mixState.esc = "forward"
				slog.Debug("getting esc out of reverse before pressing the brakes")
			}

		}
	} else { //map without using gear selections
		if getInputChangeAmount(inputs[1]) > getInputChangeAmount(inputs[2]) { //throttle is pressed more than brake
			frame.Ch[1] = uint16(MapToRangeWithDeadzoneLow(
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
			frame.Ch[1] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
		}
	}

	//Gyro Gain
	frame.Ch[2] = uint16(MapToRange(
		mixState.trims["gyro_gain"],
		-100,
		100,
		sbus.MinValue,
		sbus.MaxValue,
	))

	slog.Debug("mixed frame", "gear", mixState.gear, "esc_state", currentState, "steer", frame.Ch[0], "esc", frame.Ch[1])

	return frame, mixState
}
