package models

import "github.com/Speshl/pi_drift_wheel/sbus"

type Mixer func([]Input, MixState, ControllerOptions) (sbus.SBusFrame, MixState)

type Mapping struct {
	Label    string
	CodeName string
	Code     int
	Type     int
	Channel  int
	RawInput int
	MapType  string
	Min      int
	Max      int
	Rests    string
	Inverted bool
}

type Input struct {
	Value int
	Min   int
	Max   int
	Rests string
	Label string
}

type ControllerOptions struct {
	UseHPattern bool
}

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
