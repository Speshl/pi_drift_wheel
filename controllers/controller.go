package controllers

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/Speshl/pi_drift_wheel/sbus"
	"github.com/holoplot/go-evdev"
)

type Mixer func([]int, map[string]string, ControllerOptions) (sbus.Frame, map[string]string)

type Mapping struct {
	CodeName string
	Code     int
	Type     int
	Channel  int
	RawInput int
	MapType  string
	Min      int
	Max      int
	Inverted bool
}

type Controller struct {
	device    *evdev.InputDevice
	Name      string
	path      string
	keyMap    map[string]Mapping
	mixer     Mixer
	rawInputs []int
	mixState  map[string]string

	//options
	ControllerOptions
}

type ControllerOptions struct {
	useHPattern bool
}

func NewController(inputPath evdev.InputPath, device *evdev.InputDevice, keyMap map[string]Mapping, mixer Mixer, opts ControllerOptions) *Controller {
	rawInputs := make([]int, len(keyMap))
	for i := range rawInputs {
		rawInputs[i] = sbus.MidValue
	}

	return &Controller{
		device:            device,
		keyMap:            keyMap,
		mixer:             mixer,
		Name:              inputPath.Name,
		path:              inputPath.Path,
		rawInputs:         rawInputs,
		ControllerOptions: opts,
	}
}

func (c *Controller) Sync() error {
	e, err := c.device.ReadOne()
	if err != nil {
		return fmt.Errorf("failed reading from device: %w", err)
	}

	slog.Debug("event", "type", e.Type, "code", e.Code, "code_name", e.CodeName(), "value", e.Value)
	mapping, ok := c.keyMap[fmt.Sprintf("%d:%d", e.Type, e.Code)]
	if ok {
		updatedValue := int(e.Value)
		if mapping.Inverted {
			updatedValue = mapping.Max - updatedValue + mapping.Min
		}

		//update raw input
		c.rawInputs[mapping.RawInput] = MapToRange(
			updatedValue,
			mapping.Min,
			mapping.Max,
			sbus.MinValue,
			sbus.MaxValue,
		)
	}
	return nil
}

func (c *Controller) BuildFrame() sbus.Frame {
	frame, state := c.mixer(c.rawInputs, c.mixState, c.ControllerOptions)
	c.mixState = state
	return frame
}

func (c *Controller) ShowCaps() {
	for _, t := range c.device.CapableTypes() {
		log.Printf("  Event type %d (%s)\n", t, evdev.TypeName(t))

		state, err := c.device.State(t)
		if err == nil {
			for code, value := range state {
				log.Printf("    Event code %d (%s) state %v\n", code, evdev.CodeName(t, code), value)
			}
		}

		if t != evdev.EV_ABS {
			continue
		}

		absInfos, err := c.device.AbsInfos()
		if err != nil {
			continue
		}

		for code, absInfo := range absInfos {
			log.Printf("    Event code %d (%s)\n", code, evdev.CodeName(t, code))
			log.Printf("      Value: %d\n", absInfo.Value)
			log.Printf("      Min: %d\n", absInfo.Minimum)
			log.Printf("      Max: %d\n", absInfo.Maximum)

			if absInfo.Fuzz != 0 {
				log.Printf("      Fuzz: %d\n", absInfo.Fuzz)
			}
			if absInfo.Flat != 0 {
				log.Printf("      Flat: %d\n", absInfo.Flat)
			}
			if absInfo.Resolution != 0 {
				log.Printf("      Resolution: %d\n", absInfo.Resolution)
			}
		}
	}
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
	} else if minReturn+deadZone > value {
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
