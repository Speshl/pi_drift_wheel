package controllers

import (
	"fmt"
	"math"

	"github.com/Speshl/pi_drift_wheel/sbus"
)

type Mixer func([]Input, MixState, ControllerOptions) (sbus.Frame, MixState)

type MixState struct {
	Buttons map[string]int
	Esc     string
	Gear    int
	Trims   map[string]int
	//Aux     map[string]int
}

func NewMixState() MixState {
	return MixState{
		Buttons: make(map[string]int, 32),
		Esc:     "",
		Trims:   make(map[string]int, 10),
		//Aux:     make(map[string]string, 1),
	}
}

func (m *MixState) IsEmpty() bool {
	if m.Esc == "" && len(m.Buttons) == 0 /*&& len(m.Aux) == 0*/ {
		return true
	}
	return false
}

func (c *ControllerManager) GetMixedFrame() (sbus.Frame, error) {
	if len(c.Controllers) == 0 {
		return sbus.Frame{}, fmt.Errorf("no controllers loaded")
	}

	mixedInputs := c.Controllers[0].GetRawInputs()

	for i := 1; i < len(c.Controllers); i++ {
		inputs := c.Controllers[i].GetRawInputs()

		for j := range inputs {
			currInputChange := getInputChangeAmount(mixedInputs[j])
			newInputChange := getInputChangeAmount(inputs[j])
			if newInputChange > currInputChange {
				mixedInputs[j] = inputs[j]
			}
		}
	}

	frame, state := c.mixer(mixedInputs, c.mixState, c.ControllerOptions)
	c.mixState = state
	return frame, nil

}

func getInputChangeAmount(input Input) int {
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
