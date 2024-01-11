package controllers

import (
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/Speshl/pi_drift_wheel/go-evdev"
)

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

type Controller struct {
	device *evdev.InputDevice
	Name   string
	path   string
	keyMap map[string]Mapping

	deviceLock sync.RWMutex

	inputLock sync.RWMutex
	rawInputs []Input
}

func NewController(inputPath evdev.InputPath, device *evdev.InputDevice, keyMap map[string]Mapping) *Controller {

	rawInputs := make([]Input, 64)
	for i := range keyMap {
		rawInputs[keyMap[i].RawInput] = NewInput(keyMap[i])
	}

	return &Controller{
		device:    device,
		keyMap:    keyMap,
		Name:      inputPath.Name,
		path:      inputPath.Path,
		rawInputs: rawInputs,
	}
}

func NewInput(keyMap Mapping) Input {
	value := 0
	switch keyMap.Rests {
	case "high":
		value = keyMap.Max
	case "middle":
		value = (keyMap.Max + keyMap.Min) / 2
	case "low":
		fallthrough
	default:
		value = keyMap.Min
	}

	return Input{
		Value: value,
		Min:   keyMap.Min,
		Max:   keyMap.Max,
		Rests: keyMap.Rests,
	}
}

func (c *Controller) Sync() error {
	c.deviceLock.Lock()
	e, err := c.device.ReadOne()
	if err != nil {
		return fmt.Errorf("failed reading from device: %w", err)
	}
	c.deviceLock.Unlock()

	slog.Info("event", "type", e.Type, "code", e.Code, "code_name", e.CodeName(), "value", e.Value)
	mapping, ok := c.keyMap[fmt.Sprintf("%d:%d", e.Type, e.Code)]
	if ok {
		slog.Debug("mapped event", "label", mapping.Label, "type", e.Type, "code", e.Code, "code_name", e.CodeName(), "value", e.Value)
		updatedValue := int(e.Value)
		if mapping.Inverted {
			updatedValue = mapping.Max - updatedValue + mapping.Min
		}

		//update raw input
		c.inputLock.Lock()
		c.rawInputs[mapping.RawInput] = Input{
			Label: mapping.Label,
			Value: updatedValue,
			Min:   mapping.Min,
			Max:   mapping.Max,
			Rests: mapping.Rests,
		}
		c.inputLock.Unlock()
	}
	return nil
}

func (c *Controller) GetRawInputs() []Input {
	c.inputLock.RLock()
	defer c.inputLock.RUnlock()
	return c.rawInputs
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

func (c *Controller) SetForceFeedback(level int16) error {
	c.deviceLock.Lock()
	defer c.deviceLock.Unlock()
	err := c.device.UploadEffect(level)
	if err != nil {
		return err
	}

	//c.forceFeedback = level

	return nil
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
