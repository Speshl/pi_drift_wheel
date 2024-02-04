package models

import "math"

func GetInputChangeAmount(input Input) int {
	inputChangeAmt := 0
	switch input.Rests {
	case "low":
		inputChangeAmt = input.Value - input.Min
	case "middle":
		midValue := (input.Min + input.Max) / 2
		inputChangeAmt = int(math.Abs(float64(input.Value - midValue)))
	case "high":
		inputChangeAmt = input.Max - input.Value
	}
	return inputChangeAmt
}

func MapToRangeWithDeadzoneMid(value, min, max, minReturn, maxReturn, deadzone int) int {
	midValue := (maxReturn + minReturn) / 2

	mappedValue := MapToRange(value, min, max, minReturn, maxReturn)
	if midValue+deadzone > mappedValue && mappedValue > midValue {
		return midValue
	} else if midValue-deadzone < mappedValue && mappedValue < midValue {
		return midValue
	} else {
		return mappedValue
	}
}

func MapToRangeWithDeadzoneLow(value, min, max, minReturn, maxReturn, deadZone int) int {
	mappedValue := MapToRange(value, min, max, minReturn, maxReturn)

	if mappedValue > maxReturn {
		return maxReturn
	} else if mappedValue < minReturn {
		return minReturn
	} else if minReturn+deadZone > mappedValue {
		return minReturn
	} else {
		return mappedValue
	}
}

func MapToRange(value, min, max, minReturn, maxReturn int) int {
	mappedValue := (maxReturn-minReturn)*(value-min)/(max-min) + minReturn

	if mappedValue > maxReturn {
		return maxReturn
	} else if mappedValue < minReturn {
		return minReturn
	} else {
		return mappedValue
	}
}
