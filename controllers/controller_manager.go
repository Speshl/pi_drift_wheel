package controllers

import (
	"fmt"
	"log"

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
			return fmt.Errorf("Cannot read %s: %w\n", inputPath.Path, err)
		}

		log.Printf("loaded: %s:\t%s\n", inputPath.Path, inputPath.Name)
		c.Controllers = append(c.Controllers, NewController(inputPath, device))
	}
	return nil
}

func (c *ControllerManager) isSupported(name string) bool {
	return false
}
