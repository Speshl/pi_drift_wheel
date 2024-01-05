package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Speshl/pi_drift_wheel/config"
	"github.com/Speshl/pi_drift_wheel/go-evdev"

	//"github.com/holoplot/go-evdev"
	"golang.org/x/sync/errgroup"
)

const (
	MaxControllers = 128
)

type Input struct {
	Value int
	Min   int
	Max   int
	Rests string
	Label string
}

type ControllerManager struct {
	Controllers []*Controller
	mixer       Mixer
	mixState    MixState

	ControllerOptions
}

type ControllerOptions struct {
	UseHPattern bool
}

func NewControllerManager(cfg config.ControllerManagerConfig, mixer Mixer, opts ControllerOptions) *ControllerManager {
	return &ControllerManager{
		Controllers:       make([]*Controller, 0, MaxControllers),
		mixer:             mixer,
		ControllerOptions: opts,
	}
}

func (c *ControllerManager) Start(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	doneChan := make(chan error)

	for i := range c.Controllers {
		group.Go(func() error {
			for {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				err := c.Controllers[i].Sync()
				if err != nil {
					return fmt.Errorf("failed syncing event for %s: %w", c.Controllers[i].Name, err)
				}
			}
		})
	}

	go func() {
		doneChan <- group.Wait() //block in goroutine so we can still listen for ctx close
	}()

	for {
		select {
		case <-ctx.Done():
			slog.Info("controller manager context was cancelled")
			return ctx.Err()
		case groupErr := <-doneChan:
			return fmt.Errorf("controller manager stopping due to error - %w", groupErr)
		}
	}
}

func (c *ControllerManager) LoadControllers() error {
	inputPaths, err := evdev.ListDevicePaths()
	if err != nil {
		return fmt.Errorf("failed listing device paths: %w", err)
	}
	for _, inputPath := range inputPaths {
		if !c.isSupported(inputPath.Name) {
			slog.Info("unsupported device", "path", inputPath.Path, "name", inputPath.Name)
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

		keyMap, err := c.GetKeyMap(inputPath.Name)
		if err != nil {
			return fmt.Errorf("failed getting keymap for %s: %w", inputPath.Name, err)
		}

		controller := NewController(inputPath, device, keyMap)
		controller.ShowCaps()
		c.Controllers = append(c.Controllers, controller)
	}
	return nil
}

func (c *ControllerManager) isSupported(name string) bool {
	_, err := c.GetKeyMap(name)
	if err != nil {
		return false
	}
	return true
}

func (c *ControllerManager) GetKeyMap(name string) (map[string]Mapping, error) {
	switch name {
	case "G27 Racing Wheel":
		return GetG27KeyMap(), nil
	// case "Arduino LLC Arduino Micro":
	// 	return GetDIYHandBrakeKeyMap(), nil
	default:
		return nil, fmt.Errorf("no keymap found")
	}
}

func (c *ControllerManager) SetForceFeedback() error {
	return c.Controllers[0].SetForceFeedback()
}
