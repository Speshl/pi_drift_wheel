package controllers

import (
	"log/slog"

	sbus "github.com/Speshl/pi_drift_wheel/sbus"
)

//Mapping - 0 steer, 1 esc, 2 gyro gain

func WheelMixer(inputs []Input, mixState MixState, opts ControllerOptions) (sbus.SBusFrame, MixState) {
	frame := sbus.NewSBusFrame()

	if mixState.IsEmpty() {
		mixState = NewMixState()
		mixState.Esc = "forward"
		mixState.Gear = 0
	}

	//Check for button state changes
	for i := range inputs {
		if inputs[i].Value == mixState.Buttons[inputs[i].Label] {
			continue
		}

		mixState.Buttons[inputs[i].Label] = inputs[i].Value

		if inputs[i].Value == inputs[i].Min && inputs[i].Rests == "low" ||
			inputs[i].Value == inputs[i].Max && inputs[i].Rests == "high" ||
			inputs[i].Value == 0 && inputs[i].Rests == "mid" {
			continue //input is at its resting value (not pressed)
		}

		switch inputs[i].Label {
		case "upshift":
			slog.Info("mixing upshift")
			if mixState.Gear >= -1 {
				mixState.Gear++
			}
		case "downshift":
			slog.Info("mixing downshift")
			if mixState.Gear <= 6 {
				mixState.Gear--
			}
		case "top_left":
			if mixState.Trims["gyro_gain"] > -100 {
				mixState.Trims["gyro_gain"]--
			}
		case "top_right":
			if mixState.Trims["gyro_gain"] < 100 {
				mixState.Trims["gyro_gain"]++
			}
		}
	}

	//Build frame values based on current state/buttons

	//Steer Value
	frame.Frame.Ch[0] = uint16(MapToRangeWithDeadzoneMid(
		inputs[0].Value,
		inputs[0].Min,
		inputs[0].Max,
		sbus.MinValue,
		sbus.MaxValue,
		2,
	))

	//ESC Value
	currentState := mixState.Esc
	if opts.UseHPattern {
		for i := 10; i < 20; i++ {
			if inputs[i].Value > inputs[i].Min {
				if i == 19 {
					mixState.Gear = -1 //reverse
				} else if i > 15 {
					mixState.Gear = 0 //set neutral when unsupported gear pressed
					continue
				} else {
					mixState.Gear = i - 9 //supported gears
				}
				break //only 1 gear can be active at a time, so stop when one found
			} else if i == 19 {
				mixState.Gear = 0 //no gear button pressed, set to neutral
			}
		}

		if getInputChangeAmount(inputs[1]) > getInputChangeAmount(inputs[2]) { //throttle is pressed more than brake
			if mixState.Gear == 0 { //Neutral so keep esc at center value
				frame.Frame.Ch[1] = uint16(sbus.MidValue)
			} else if mixState.Gear == -1 { //Reverse
				switch currentState {
				case "forward": //Going to reverse from forward needs to turn on brakes first then reverse to get esc into reverse mode
					// value := MapToRange( //Map throttle to bottom half of esc channel
					// 	inputs[1].Value,
					// 	inputs[1].Min,
					// 	inputs[1].Max,
					// 	sbus.MinValue,
					// 	sbus.MidValue,
					// )
					// frame.Frame.Ch[1] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
					// if frame.Frame.Ch[1] < uint16(sbus.MidValue) {
					// 	mixState.Esc = "brake"
					// }
					frame.Frame.Ch[1] = uint16(sbus.MinValue) //full brakes
					frame.Priority = true
					mixState.Esc = "brake"
					slog.Info("setting esc to full brakes before switching to reverse")

				case "brake":
					frame.Frame.Ch[1] = uint16(sbus.MidValue) //set mid to get the esc out of brake
					frame.Priority = true
					mixState.Esc = "reverse"
					slog.Info("setting esc center to get out of brake and prepare for reverse")
				case "reverse":
					value := MapToRange( //Map throttle to bottom half of esc channel
						inputs[1].Value,
						inputs[1].Min,
						inputs[1].Max,
						sbus.MinValue,
						sbus.MidValue,
					)
					frame.Frame.Ch[1] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
				}

			} else if mixState.Gear > 0 && mixState.Gear <= 6 {
				value := int(float64(inputs[1].Value) / float64(6) * float64(mixState.Gear)) //Scale throttle to gear
				if mixState.Gear == 6 {
					value = inputs[1].Value //let top gear have full range without rounding issues
				}

				value = MapToRange(
					value,
					inputs[1].Min,
					inputs[1].Max,
					sbus.MidValue,
					sbus.MaxValue,
				)
				frame.Frame.Ch[1] = uint16(value)
				if frame.Frame.Ch[1] > uint16(sbus.MidValue) {
					mixState.Esc = "forward"
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
				frame.Frame.Ch[1] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
				if frame.Frame.Ch[1] < uint16(sbus.MidValue) {
					mixState.Esc = "brake"
					frame.Priority = true
					slog.Info("braking from forward")
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
				frame.Frame.Ch[1] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
				if frame.Frame.Ch[1] > uint16(sbus.MidValue-40) {                 //brakes not/or barely pushed
					frame.Frame.Ch[1] = uint16(sbus.MidValue + 50) //set enough forward keep esc out of reverse
					frame.Priority = true
					mixState.Esc = "forward"
					slog.Info("keeping brakes from going to reverse, by setting slightly forward")
				}

			case "reverse":
				frame.Frame.Ch[1] = uint16(sbus.MaxValue) //set enough forward to get the esc out of reverse
				frame.Priority = true
				mixState.Esc = "forward"
				slog.Info("getting esc out of reverse before pressing the brakes")
			}

		}
	} else { //map without using gear selections
		if getInputChangeAmount(inputs[1]) > getInputChangeAmount(inputs[2]) { //throttle is pressed more than brake
			frame.Frame.Ch[1] = uint16(MapToRangeWithDeadzoneLow(
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
			frame.Frame.Ch[1] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
		}
	}

	//Gyro Gain
	frame.Frame.Ch[2] = uint16(MapToRange(
		mixState.Trims["gyro_gain"],
		-100,
		100,
		sbus.MinValue,
		sbus.MaxValue,
	))

	slog.Debug("mixed frame", "gear", mixState.Gear, "esc_state", currentState, "steer", frame.Frame.Ch[0], "esc", frame.Frame.Ch[1])

	return frame, mixState
}
