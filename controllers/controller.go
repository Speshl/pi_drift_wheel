package controllers

import (
	"fmt"
	"log"

	"github.com/Speshl/pi_drift_wheel/channels"
	"github.com/holoplot/go-evdev"
)

type Mapping struct {
	CodeName string
	Channel  int
	Type     string //full, bottom, top
	Min      int
	Max      int
	Inverted bool
}

type Controller struct {
	device *evdev.InputDevice
	name   string
	path   string
	keyMap map[string]Mapping

	channels *channels.ChannelGroup
}

func NewController(inputPath evdev.InputPath, device *evdev.InputDevice, keyMap map[string]Mapping) *Controller {
	return &Controller{
		device:   device,
		keyMap:   keyMap,
		name:     inputPath.Name,
		path:     inputPath.Path,
		channels: channels.NewChannelGroup(),
	}
}

func (c *Controller) Sync() error {
	e, err := c.device.ReadOne()
	if err != nil {
		return fmt.Errorf("failed reading from device: %w", err)
	}

	//slog.Info("event", "type", e.Type, "code", e.Code, "code_name", e.CodeName(), "value", e.Value)
	mapping, ok := c.keyMap[e.CodeName()]
	if ok {
		updatedValue := int(e.Value)
		if mapping.Inverted {
			updatedValue = mapping.Max - updatedValue + mapping.Min
		}
		c.channels.SetChannel(mapping.Channel, updatedValue, mapping.Type, mapping.Min, mapping.Max)
	}

	return nil
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
