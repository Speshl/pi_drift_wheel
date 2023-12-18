package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Speshl/pi_drift_wheel/config"
	"github.com/Speshl/pi_drift_wheel/sbus"
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

func (a *App) Start(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	group, ctx := errgroup.WithContext(ctx)

	// controllerManager := controllers.NewControllerManager(a.cfg.ControllerManagerCfg)
	// err := controllerManager.LoadControllers()
	// if err != nil {
	// 	return fmt.Errorf("failed loading controllers: %w", err)
	// }

	// //Start data input processes
	// group.Go(func() error {
	// 	return controllerManager.Start(ctx)
	// })

	sbusReader := sbus.NewSBusReader(a.cfg.SbusCfg)

	group.Go(func() error {
		err := sbusReader.ListPorts()
		if err != nil {
			return err
		}
		defer sbusReader.Cleanup()
		return sbusReader.Start(ctx)
	})

	//Start data output processes
	// group.Go(func() error {
	// 	for {
	// 		controller := controllerManager.Controllers[0]
	// 		channelGroup := controller.GetChannelGroup()
	// 		channelData := channelGroup.GetChannels()
	// 		slog.Info("controller state",
	// 			"name", controller.Name,
	// 			"chan0", channelData[0],
	// 			"chan1", channelData[1],
	// 		)
	// 		time.Sleep(1000)
	// 	}
	// })

	//kill listener
	group.Go(func() error {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-signalChannel:
			slog.Info("received signal", "value", sig)
			cancel()
			return fmt.Errorf("received signal: %s", sig)
		case <-ctx.Done():
			slog.Info("closing signal goroutine", "error", ctx.Err().Error())
			return ctx.Err()
		}
	})

	err = group.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			slog.Info("app context was cancelled")
			return nil
		} else {
			return fmt.Errorf("app stopping due to error - %w", err)
		}
	}
	return nil
}
