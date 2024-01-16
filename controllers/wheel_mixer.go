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

	gasChange := getInputChangeAmount(inputs[1])
	brakeChange := getInputChangeAmount(inputs[2])

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

		if gasChange > brakeChange && gasChange > 10 { //throttle is pressed more than brake
			if mixState.Gear == 0 { //Neutral so keep esc at center value
				frame.Frame.Ch[1] = uint16(sbus.MidValue)
				slog.Info("neutral")
			} else if mixState.Gear == -1 { //Reverse
				switch currentState {
				case "forward": //Going to reverse from forward needs to turn on brakes first then reverse to get esc into reverse mode
					frame.Frame.Ch[1] = uint16(sbus.MinValue) //full brakes
					frame.Priority = 6
					mixState.Esc = "brake"
					slog.Info("setting esc to full brakes before switching to reverse")

				case "brake":
					frame.Frame.Ch[1] = uint16(sbus.MidValue) //set mid to get the esc out of brake
					frame.Priority = 6
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
					slog.Info("setting reverse in reverse", "value", frame.Frame.Ch[1])
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
				slog.Info("setting forward in forward", "value", frame.Frame.Ch[1])
			} else {
				slog.Warn("gear out of bounds")
			}
		} else if brakeChange > gasChange && brakeChange > 10 { //Brake is pressed more than throttle
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
				if frame.Frame.Ch[1] < uint16(sbus.MidValue-70) {

					mixState.Esc = "brake"
					frame.Priority = 6
					slog.Info("to brake from forward", "esc", frame.Frame.Ch[1], "base", inputs[2].Value)
				} else {
					slog.Info("braking from forward", "esc", frame.Frame.Ch[1], "base", inputs[2].Value)
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
				if frame.Frame.Ch[1] > uint16(sbus.MidValue-70) {                 //brakes not/or barely pushed
					frame.Frame.Ch[1] = uint16(sbus.MidValue + 50) //set enough forward keep esc out of reverse
					frame.Priority = 6
					mixState.Esc = "forward"
					slog.Info("keeping brakes from going to reverse, by setting slightly forward")
				} else {
					slog.Info("brakes in brake state", "esc", frame.Frame.Ch[1])
				}
			case "reverse":
				frame.Frame.Ch[1] = uint16(sbus.MaxValue) //set enough forward to get the esc out of reverse
				frame.Priority = 6
				mixState.Esc = "forward"
				slog.Info("getting esc out of reverse before pressing the brakes", "esc", frame.Frame.Ch[1])
			}
		} else {
			frame.Frame.Ch[1] = uint16(sbus.MidValue)
			//slog.Info("no peddals", "throttle", gasChange, "brake", brakeChange)
		}
	} else { //map without using gear selections
		if gasChange > brakeChange && gasChange > 10 { //throttle is pressed more than brake
			frame.Frame.Ch[1] = uint16(MapToRangeWithDeadzoneLow(
				inputs[1].Value,
				inputs[1].Min,
				inputs[1].Max,
				sbus.MidValue,
				sbus.MaxValue,
				2,
			))
			slog.Info("no gear gas")
		} else if brakeChange > gasChange && brakeChange > 10 { //brake pressed more or equal to throttle
			value := MapToRangeWithDeadzoneLow(
				inputs[2].Value,
				inputs[2].Min,
				inputs[2].Max,
				sbus.MinValue,
				sbus.MidValue,
				2,
			)
			frame.Frame.Ch[1] = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
			slog.Info("no gear brake")
		} else {
			frame.Frame.Ch[1] = uint16(sbus.MidValue)
			//slog.Info("no gears no peddals", "throttle", gasChange, "brake", brakeChange)
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
