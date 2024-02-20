package controllers

import (
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/Speshl/pi_drift_wheel/controllers/models"
	"github.com/Speshl/pi_drift_wheel/go-evdev"
)

type Controller struct {
	device *evdev.InputDevice
	Name   string
	path   string
	keyMap map[string]models.Mapping

	ffLock      sync.RWMutex
	ffLevel     int16
	lastFFLevel int16

	inputLock sync.RWMutex
	rawInputs []models.Input
}

func NewController(inputPath evdev.InputPath, device *evdev.InputDevice, keyMap map[string]models.Mapping) *Controller {

	rawInputs := make([]models.Input, 64)
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

func NewInput(keyMap models.Mapping) models.Input {
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

	return models.Input{
		Value: value,
		Min:   keyMap.Min,
		Max:   keyMap.Max,
		Rests: keyMap.Rests,
	}
}

func (c *Controller) Sync() error {

	e, err := c.device.ReadOne()
	if err != nil {
		return fmt.Errorf("failed reading from device: %w", err)
	}
	// c.ffLock.Lock()
	// ffLevel := c.ffLevel
	// c.ffLock.Unlock()

	// if ffLevel != c.lastFFLevel {
	// 	err = c.device.UploadEffect(ffLevel)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	c.lastFFLevel = ffLevel
	// }

	slog.Debug("event", "type", e.Type, "code", e.Code, "code_name", e.CodeName(), "value", e.Value)
	mapping, ok := c.keyMap[fmt.Sprintf("%d:%d", e.Type, e.Code)]
	if ok {
		slog.Info("mapped event", "label", mapping.Label, "type", e.Type, "code", e.Code, "code_name", e.CodeName(), "value", e.Value)
		updatedValue := int(e.Value)
		if mapping.Inverted {
			updatedValue = mapping.Max - updatedValue + mapping.Min
		}

		//update raw input
		c.inputLock.Lock()
		c.rawInputs[mapping.RawInput] = models.Input{
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

func (c *Controller) GetRawInputs() []models.Input {
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
	// c.ffLock.Lock()
	// defer c.ffLock.Unlock()
	err := c.device.UploadEffect(level)
	if err != nil {
		return err
	}

	// c.ffLevel = level

	return nil
}
