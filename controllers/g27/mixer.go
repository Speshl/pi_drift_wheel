package g27

import (
	"log/slog"

	"github.com/Speshl/pi_drift_wheel/controllers/models"
	sbus "github.com/Speshl/pi_drift_wheel/sbus"
)

//Mapping - 0 steer, 1 esc, 2 gyro gain

func Mixer(inputs []models.Input, mixState models.MixState, opts models.ControllerOptions) (sbus.SBusFrame, models.MixState) {
	frame := sbus.NewSBusFrame()

	if mixState.IsEmpty() {
		mixState = models.NewMixState()
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
	frame.Frame.Ch[0] = uint16(models.MapToRangeWithDeadzoneMid(
		inputs[0].Value,
		inputs[0].Min,
		inputs[0].Max,
		sbus.MinValue,
		sbus.MaxValue,
		2,
	))

	frame.Frame.Ch[1], frame.Priority, mixState = getEscValue(inputs, mixState, opts)

	//Gyro Gain
	frame.Frame.Ch[2] = uint16(models.MapToRange(
		mixState.Trims["gyro_gain"],
		-100,
		100,
		sbus.MinValue,
		sbus.MaxValue,
	))

	slog.Debug("mixed frame", "gear", mixState.Gear, "esc_state", mixState.Esc, "steer", frame.Frame.Ch[0], "esc", frame.Frame.Ch[1])

	return frame, mixState
}

func getEscValue(inputs []models.Input, mixState models.MixState, opts models.ControllerOptions) (uint16, int, models.MixState) {
	if opts.UseHPattern {
		return getEscValueWithHPattern(inputs, mixState)
	}
	return getEscValueWithoutGears(inputs, mixState)
}

func getEscValueWithHPattern(inputs []models.Input, mixState models.MixState) (uint16, int, models.MixState) {
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

	if mixState.Gear == -1 { //reverse
		return getEscValueReverse(inputs, mixState)
	} else if mixState.Gear > 0 && mixState.Gear < 7 { //forward gear
		return getEscValueForward(inputs, mixState)
	}
	return uint16(sbus.MidValue), 0, mixState //neutral

}

func getEscValueReverse(inputs []models.Input, mixState models.MixState) (uint16, int, models.MixState) {
	returnValue := uint16(sbus.MinValue)
	returnPriority := 0
	gasChange := models.GetInputChangeAmount(inputs[1])
	brakeChange := models.GetInputChangeAmount(inputs[2])

	if gasChange >= brakeChange && gasChange > 5 { //Throttle pressed
		switch mixState.Esc {
		case "forward": //go to brakes
			returnValue = uint16(sbus.MidValue) - 50
			returnPriority = 3
			mixState.Esc = "brake"
			slog.Info("to brake from forward in reverse", "esc", returnValue, "state", mixState.Esc)
		case "brake": //go to reverse
			returnValue = uint16(sbus.MidValue)
			returnPriority = 3
			mixState.Esc = "reverse"
			slog.Info("to reverse from brake in reverse", "esc", returnValue, "state", mixState.Esc)
		case "reverse": //set reverse value
			value := inputs[1].Value
			value = models.MapToRange(
				value,
				inputs[1].Min,
				inputs[1].Max,
				sbus.MinValue,
				sbus.MidValue,
			)
			returnValue = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
			slog.Info("reverseing in reverse in reverse", "esc", returnValue, "state", mixState.Esc)
		}
	} else if brakeChange >= gasChange && brakeChange > 5 { //brake pressed
		switch mixState.Esc {
		case "forward": //set value
			returnValue = uint16(sbus.MidValue) - 50
			returnPriority = 5
			mixState.Esc = "brake"
			slog.Info("to brake from forward in reverse", "esc", returnValue, "state", mixState.Esc)
		case "brake": //set value
			value := models.MapToRangeWithDeadzoneLow(
				inputs[2].Value,
				inputs[2].Min,
				inputs[2].Max,
				sbus.MinValue,
				sbus.MidValue,
				2,
			)
			returnValue = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
			slog.Info("braking in brake in reverse", "esc", returnValue)
		case "reverse": //go to forward
			returnValue = uint16(sbus.MidValue) + 50
			returnPriority = 5
			mixState.Esc = "forward"
			slog.Info("to forward from reverse braking in reverse", "esc", returnValue, "state", mixState.Esc)
		}
	} else {
		//put back in reverse esc state since no input to prepare for next input
		switch mixState.Esc {
		case "forward":
			returnValue = uint16(sbus.MidValue)
			//returnPriority = 3
			//mixState.Esc = "brake"
			//slog.Info("no input in forward, switch to brakes", "esc", returnValue, "state", mixState.Esc)
		case "brake":
			returnValue = uint16(sbus.MidValue) + 50
			returnPriority = 5
			mixState.Esc = "forward"
			slog.Info("no input in brake in reverse", "esc", returnValue, "state", mixState.Esc)
		case "reverse":
			returnValue = uint16(sbus.MidValue) + 50
			returnPriority = 5
			mixState.Esc = "forward"
			slog.Info("no input in reverse", "esc", returnValue, "state", mixState.Esc)
		}
	}
	return returnValue, returnPriority, mixState
}

func getEscValueForward(inputs []models.Input, mixState models.MixState) (uint16, int, models.MixState) {
	returnValue := uint16(sbus.MinValue)
	returnPriority := 0
	gasChange := models.GetInputChangeAmount(inputs[1])
	brakeChange := models.GetInputChangeAmount(inputs[2])

	if gasChange >= brakeChange && gasChange > 5 { //Throttle pressed
		value := int(float64(inputs[1].Value) / float64(6) * float64(mixState.Gear)) //Scale throttle to gear
		if mixState.Gear == 6 {
			value = inputs[1].Value //let top gear have full range without rounding issues
		}

		value = models.MapToRange(
			value,
			inputs[1].Min,
			inputs[1].Max,
			sbus.MidValue,
			sbus.MaxValue,
		)
		returnValue = uint16(value)
		if returnValue > uint16(sbus.MidValue) && mixState.Esc != "forward" {
			mixState.Esc = "forward"
			returnPriority = 3
		}
		slog.Info("setting forward in forward", "esc", returnValue, "next_state", mixState.Esc)
	} else if brakeChange >= gasChange && brakeChange > 5 { //brake pressed
		switch mixState.Esc {
		case "forward": //brake in a forward gear when esc is in forward state

			value := models.MapToRangeWithDeadzoneLow(
				inputs[2].Value,
				inputs[2].Min,
				inputs[2].Max,
				sbus.MinValue,
				sbus.MidValue,
				2,
			)
			returnValue = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
			if returnValue < uint16(sbus.MidValue) {
				mixState.Esc = "brake"
				returnPriority = 3
				slog.Info("to brake from forward", "esc", returnValue)
			} else {
				slog.Info("braking from forward", "esc", returnValue)
			}

		case "reverse": //put esc back in forward state before hitting the brakes
			returnValue = uint16(sbus.MidValue) + 50
			returnPriority = 3
			mixState.Esc = "forward"
		case "brake":

			value := models.MapToRangeWithDeadzoneLow(
				inputs[2].Value,
				inputs[2].Min,
				inputs[2].Max,
				sbus.MinValue,
				sbus.MidValue,
				2,
			)
			returnValue = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half

		case "default":
			slog.Info("invalid esc state")
		}

	} else {
		//put back in forward esc state since no input to prepare for next input
		if mixState.Esc != "forward" {
			returnValue = uint16(sbus.MidValue) + 50
			returnPriority = 3
			mixState.Esc = "forward"
			slog.Info("no input and not in forward, putting back to forward")
		} else {
			returnValue = uint16(sbus.MidValue)
			slog.Info("no input and in forward")
		}
	}

	return returnValue, returnPriority, mixState
}

func getEscValueWithoutGears(inputs []models.Input, mixState models.MixState) (uint16, int, models.MixState) {
	returnValue := uint16(sbus.MinValue)
	returnPriority := 0

	gasChange := models.GetInputChangeAmount(inputs[1])
	brakeChange := models.GetInputChangeAmount(inputs[2])
	if gasChange > brakeChange && gasChange > 10 { //throttle is pressed more than brake
		returnValue = uint16(models.MapToRangeWithDeadzoneLow(
			inputs[1].Value,
			inputs[1].Min,
			inputs[1].Max,
			sbus.MidValue,
			sbus.MaxValue,
			2,
		))
		slog.Debug("gas without gears")
	} else if brakeChange > gasChange && brakeChange > 10 { //brake pressed more or equal to throttle
		value := models.MapToRangeWithDeadzoneLow(
			inputs[2].Value,
			inputs[2].Min,
			inputs[2].Max,
			sbus.MinValue,
			sbus.MidValue,
			2,
		)
		returnValue = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
		slog.Debug("brake without gears")
	} else {
		returnValue = uint16(sbus.MidValue)
		slog.Debug("no pedall without gears", "throttle", gasChange, "brake", brakeChange)
	}
	return returnValue, returnPriority, mixState
}

func getEscValueOld(inputs []models.Input, mixState models.MixState, opts models.ControllerOptions) (uint16, int, models.MixState) {
	returnValue := uint16(sbus.MinValue)
	returnPriority := 0

	gasChange := models.GetInputChangeAmount(inputs[1])
	brakeChange := models.GetInputChangeAmount(inputs[2])
	//ESC Value
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
				returnValue = uint16(sbus.MidValue)
				slog.Info("neutral")
			} else if mixState.Gear == -1 { //Reverse
				switch mixState.Esc {
				case "forward": //Going to reverse from forward needs to turn on brakes first then reverse to get esc into reverse mode
					returnValue = uint16(sbus.MinValue) //full brakes
					returnPriority = 6
					mixState.Esc = "brake"
					slog.Info("setting esc to full brakes before switching to reverse")

				case "brake":
					returnValue = uint16(sbus.MidValue) //set mid to get the esc out of brake
					returnPriority = 6
					mixState.Esc = "reverse"
					slog.Info("setting esc center to get out of brake and prepare for reverse")
				case "reverse":
					value := models.MapToRange( //Map throttle to bottom half of esc channel
						inputs[1].Value,
						inputs[1].Min,
						inputs[1].Max,
						sbus.MinValue,
						sbus.MidValue,
					)
					returnValue = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
					slog.Info("setting reverse in reverse", "value", returnValue)
				}

			} else if mixState.Gear > 0 && mixState.Gear <= 6 {
				value := int(float64(inputs[1].Value) / float64(6) * float64(mixState.Gear)) //Scale throttle to gear
				if mixState.Gear == 6 {
					value = inputs[1].Value //let top gear have full range without rounding issues
				}

				value = models.MapToRange(
					value,
					inputs[1].Min,
					inputs[1].Max,
					sbus.MidValue,
					sbus.MaxValue,
				)
				returnValue = uint16(value)
				if returnValue > uint16(sbus.MidValue) {
					mixState.Esc = "forward"
				}
				slog.Info("setting forward in forward", "value", returnValue)
			} else {
				slog.Warn("gear out of bounds")
			}
		} else if brakeChange > gasChange && brakeChange > 10 { //Brake is pressed more than throttle
			switch mixState.Esc {
			case "forward":
				value := models.MapToRangeWithDeadzoneLow(
					inputs[2].Value,
					inputs[2].Min,
					inputs[2].Max,
					sbus.MinValue,
					sbus.MidValue,
					2,
				)
				returnValue = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
				if returnValue < uint16(sbus.MidValue) {

					mixState.Esc = "brake"
					returnPriority = 6
					slog.Info("to brake from forward", "esc", returnValue, "base", inputs[2].Value)
				} else {
					slog.Info("braking from forward", "esc", returnValue, "base", inputs[2].Value)
				}
			case "brake":
				value := models.MapToRangeWithDeadzoneLow(
					inputs[2].Value,
					inputs[2].Min,
					inputs[2].Max,
					sbus.MinValue,
					sbus.MidValue,
					2,
				)
				returnValue = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
				if returnValue < uint16(sbus.MidValue) {                    //brakes not/or barely pushed
					returnValue = uint16(sbus.MidValue + 100) //set enough forward keep esc out of reverse
					returnPriority = 6
					mixState.Esc = "forward"
					slog.Info("keeping brakes from going to reverse, by setting slightly forward", "esc", returnValue)
				} else {
					slog.Info("brakes in brake state", "esc", returnValue)
				}
			case "reverse":
				returnValue = uint16(sbus.MidValue + 50) //set enough forward to get the esc out of reverse
				returnPriority = 6
				mixState.Esc = "forward"
				slog.Info("getting esc out of reverse before pressing the brakes", "esc", returnValue)
			}
		} else {
			returnValue = uint16(sbus.MidValue)
			//slog.Info("no peddals", "throttle", gasChange, "brake", brakeChange)
		}
	} else { //map without using gear selections
		if gasChange > brakeChange && gasChange > 10 { //throttle is pressed more than brake
			returnValue = uint16(models.MapToRangeWithDeadzoneLow(
				inputs[1].Value,
				inputs[1].Min,
				inputs[1].Max,
				sbus.MidValue,
				sbus.MaxValue,
				2,
			))
			slog.Info("no gear gas")
		} else if brakeChange > gasChange && brakeChange > 10 { //brake pressed more or equal to throttle
			value := models.MapToRangeWithDeadzoneLow(
				inputs[2].Value,
				inputs[2].Min,
				inputs[2].Max,
				sbus.MinValue,
				sbus.MidValue,
				2,
			)
			returnValue = uint16(sbus.MidValue - value + sbus.MinValue) //invert since on bottom half
			slog.Info("no gear brake")
		} else {
			returnValue = uint16(sbus.MidValue)
			//slog.Info("no gears no peddals", "throttle", gasChange, "brake", brakeChange)
		}
	}

	return returnValue, returnPriority, mixState
}
