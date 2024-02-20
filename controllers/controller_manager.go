package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Speshl/pi_drift_wheel/config"
	"github.com/Speshl/pi_drift_wheel/controllers/g27"
	"github.com/Speshl/pi_drift_wheel/controllers/handbrake_diy"
	"github.com/Speshl/pi_drift_wheel/controllers/models"
	"github.com/Speshl/pi_drift_wheel/go-evdev"
	"github.com/Speshl/pi_drift_wheel/sbus"

	//"github.com/holoplot/go-evdev"
	"golang.org/x/sync/errgroup"
)

const (
	MaxControllers = 128
)

type ControllerManager struct {
	Controllers []*Controller
	mixer       models.Mixer
	mixState    models.MixState

	models.ControllerOptions
}

func NewControllerManager(cfg config.ControllerManagerConfig, opts models.ControllerOptions) *ControllerManager {
	return &ControllerManager{
		Controllers:       make([]*Controller, 0, MaxControllers),
		ControllerOptions: opts,
	}
}

func (c *ControllerManager) Start(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	doneChan := make(chan error)

	for i := range c.Controllers {
		i := i //it got me
		group.Go(func() error {
			for {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				slog.Debug("syncing event for", "controller", c.Controllers[i].Name)
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
			//return fmt.Errorf("failed reading %s: %w\n", inputPath.Path, err)
			slog.Warn("failed opening device", "path", inputPath.Path, "name", inputPath.Name, "err", err)
			continue
		}

		inputId, err := device.InputID()
		if err != nil {
			//return fmt.Errorf("failed getting input id: %w", err)
			slog.Warn("failed getting input id", "path", inputPath.Path, "name", inputPath.Name, "err", err)
			continue
		}

		uniqueId, err := device.UniqueID()
		if err != nil {
			//return fmt.Errorf("failed getting unique id: %w", err)
			slog.Warn("failed getting unique id", "path", inputPath.Path, "name", inputPath.Name, "err", err)
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

func (c *ControllerManager) GetMixState() models.MixState {
	return c.mixState
}

func (c *ControllerManager) GetKeyMap(name string) (map[string]models.Mapping, error) {
	switch name {
	case "G27 Racing Wheel":
		c.mixer = g27.Mixer
		return g27.GetKeyMap(), nil
	case "Arduino LLC Arduino Micro":
		return handbrake_diy.GetKeyMap(), nil
	default:
		return nil, fmt.Errorf("no keymap found")
	}
}

func (c *ControllerManager) SetForceFeedback(level int16) error {
	err := c.Controllers[0].SetForceFeedback(level)
	if err != nil {
		return err
	}
	return err
}

func (c *ControllerManager) GetMixedFrame() (sbus.SBusFrame, error) {
	if len(c.Controllers) == 0 {
		return sbus.NewSBusFrame(), fmt.Errorf("no controllers loaded")
	}

	mixedInputs := c.Controllers[0].GetRawInputs()
	slog.Info("initial inputs device", "name", c.Controllers[0], "brake", mixedInputs[2])

	for i := 1; i < len(c.Controllers); i++ {
		i := i
		inputs := c.Controllers[i].GetRawInputs()
		for j := range inputs {
			currInputChange := models.GetScaledInputChange(mixedInputs[j])
			newInputChange := models.GetScaledInputChange(inputs[j])

			if newInputChange > currInputChange {
				if j == 2 {
					slog.Info("brake updated", "currInputChange", currInputChange, "newInputChange", newInputChange, "mixedInputs", mixedInputs[j], "inputs", inputs[j])
				}
				mixedInputs[j] = inputs[j]
			} else {
				if j == 2 {
					slog.Info("brake not updated", "currInputChange", currInputChange, "newInputChange", newInputChange, "mixedInputs", mixedInputs[j], "inputs", inputs[j])
				}
			}
		}
	}

	frame, state := c.mixer(mixedInputs, c.mixState, c.ControllerOptions)
	c.mixState = state
	return frame, nil
}
