package controllers

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/Speshl/pi_drift_wheel/config"
	"github.com/holoplot/go-evdev"
)

const (
	MaxControllers = 8
)

type ControllerManager struct {
	Controllers []*Controller
}

type Controller struct {
	device  *evdev.InputDevice
	name    string
	path    string
	buttons map[string]int
}

func NewControllerManager(cfg config.ControllerManagerConfig) *ControllerManager {
	return &ControllerManager{
		Controllers: make([]*Controller, 0, MaxControllers),
	}
}

func NewController(inputPath evdev.InputPath, device *evdev.InputDevice) *Controller {
	return &Controller{
		device:  device,
		name:    inputPath.Name,
		path:    inputPath.Path,
		buttons: make(map[string]int, 0),
	}
}

func (c *ControllerManager) LoadControllers() error {
	inputPaths, err := evdev.ListDevicePaths()
	if err != nil {
		return fmt.Errorf("failed listing device paths: %w", err)
	}
	for _, inputPath := range inputPaths {
		if !c.isSupported(inputPath.Name) {
			log.Printf("unsupported: %s:\t%s\n", inputPath.Path, inputPath.Name)
			continue
		}

		device, err := evdev.Open(inputPath.Path)
		if err != nil {
			return fmt.Errorf("failed reading %s: %w\n", inputPath.Path, err)
		}

		inputId, err := device.InputID()
		if err != nil {
			return fmt.Errorf("failed getting input id: %w", err)
		}
		uniqueId, err := device.UniqueID()
		if err != nil {
			return fmt.Errorf("failed getting unique id: %w", err)
		}

		slog.Info("loaded device",
			"name", inputPath.Name,
			"path", inputPath.Path,
			"bus_type", inputId.BusType,
			"vendor", inputId.Vendor,
			"product", inputId.Product,
			"version", inputId.Version,
			"uniqueId", uniqueId,
		)

		c.showControllerCaps(device)

		c.Controllers = append(c.Controllers, NewController(inputPath, device))
	}
	return nil
}

func (c *ControllerManager) isSupported(name string) bool {
	switch name {
	case "G27 Racing Wheel", "Arduino LLC Arduino Micro":
		return true
	default:
		return false
	}
}

func (c *ControllerManager) showControllerCaps(device *evdev.InputDevice) {
	for _, t := range device.CapableTypes() {
		log.Printf("  Event type %d (%s)\n", t, evdev.TypeName(t))

		state, err := device.State(t)
		if err == nil {
			for code, value := range state {
				log.Printf("    Event code %d (%s) state %v\n", code, evdev.CodeName(t, code), value)
			}
		}

		if t != evdev.EV_ABS {
			continue
		}

		absInfos, err := device.AbsInfos()
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
