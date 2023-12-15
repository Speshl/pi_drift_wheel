package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/Speshl/pi_drift_wheel/config"
	"github.com/Speshl/pi_drift_wheel/controllers"
	"golang.org/x/sync/errgroup"
)

type App struct {
	cfg config.Config
}

func NewApp(cfg config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Start(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	controllerManager := controllers.NewControllerManager(a.cfg.ControllerManagerCfg)
	err := controllerManager.LoadControllers()
	if err != nil {
		return fmt.Errorf("failed loading controllers: %w", err)
	}

	group.Go(func() error {
		return controllerManager.Start(ctx)
	})

	group.Go(func() error {
		for {
			controller := controllerManager.Controllers[0]
			channelGroup := controller.GetChannelGroup()
			channelData := channelGroup.GetChannels()
			slog.Info("controller state",
				"name", controller.Name,
				"chan0", channelData[0],
				"chan1", channelData[1],
			)
			time.Sleep(1000)
		}
	})

	err = group.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("context was cancelled")
			return nil
		} else {
			return fmt.Errorf("server stopping due to error - %w", err)
		}
	}
	return nil
}
