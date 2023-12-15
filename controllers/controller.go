package controllers

import (
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/holoplot/go-evdev"
)

type Controller struct {
	device  *evdev.InputDevice
	name    string
	path    string
	buttons map[string]int
	lock    sync.RWMutex
}

func NewController(inputPath evdev.InputPath, device *evdev.InputDevice) *Controller {
	return &Controller{
		device:  device,
		name:    inputPath.Name,
		path:    inputPath.Path,
		buttons: make(map[string]int, 0),
	}
}

func (c *Controller) Sync() error {
	e, err := c.device.ReadOne()
	if err != nil {
		return fmt.Errorf("failed reading from device: %w", err)
	}

	//ts := log.Sprintf("Event: time %d.%06d", e.Time.Sec, e.Time.Usec)

	switch e.Type {
	case evdev.EV_SYN:
		switch e.Code {
		case evdev.SYN_MT_REPORT:
			slog.Info("event", "code", e.Code, "code_name", e.CodeName(), "value", e.Value)
		case evdev.SYN_DROPPED:
			slog.Info("event", "code", e.Code, "code_name", e.CodeName(), "value", e.Value)
		default:
			slog.Info("event", "code", e.Code, "code_name", e.CodeName(), "value", e.Value)
		}
	default:
		log.Println("unknown code: %d", e.Type)
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
