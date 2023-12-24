package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Speshl/pi_drift_wheel/config"
	"github.com/Speshl/pi_drift_wheel/controllers"
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

	//Start Getting controller inputs
	controllerManager := controllers.NewControllerManager(a.cfg.ControllerManagerCfg)
	err = controllerManager.LoadControllers()
	if err != nil {
		return fmt.Errorf("failed loading controllers: %w", err)
	}
	group.Go(func() error {
		defer cancel()
		return controllerManager.Start(ctx)
	})

	//Start Sbus read/write
	sBusConns := make([]*sbus.SBus, 0, config.MaxSbus)
	for i := 0; i < config.MaxSbus; i++ {
		sBus, err := sbus.NewSBus(a.cfg.SbusCfgs[i])
		if err != nil {
			if i != 0 {
				continue
			}
		}

		sBusConns = append(sBusConns, sBus)
		group.Go(func() error {
			defer cancel()
			err := sBus.ListPorts()
			if err != nil {
				return err
			}
			return sBus.Start(ctx)
		})
	}

	// Process data
	group.Go(func() error {
		time.Sleep(500 * time.Millisecond) //give some time for signals to warm up

		framesToMerge := make([]sbus.Frame, 0, len(controllerManager.Controllers)+len(sBusConns))
		//ticker := time.NewTicker(6 * time.Millisecond) //fast ticker
		ticker := time.NewTicker(1000 * time.Millisecond) //slow ticker
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				startTime := time.Now()
				framesToMerge = framesToMerge[:0] //clear out frames before next merge

				for i := range controllerManager.Controllers {
					frame := controllerManager.Controllers[i].BuildFrame()
					slog.Info("controller frame", "frame", frame, "name", controllerManager.Controllers[i].Name)
					framesToMerge = append(framesToMerge, frame)
				}

				for i := range sBusConns {
					if sBusConns[i].Recieving && sBusConns[i].Type == sbus.RxTypeControl {
						framesToMerge = append(framesToMerge, sBusConns[i].GetReadFrame())
					} else if sBusConns[i].Recieving && sBusConns[i].Type == sbus.RxTypeTelemetry {
						slog.Debug("sbus telemetry", "frame", sBusConns[i].GetReadFrame())
					}
				}

				mergedFrame := sbus.MergeFrames(framesToMerge)

				for i := range sBusConns {
					if sBusConns[i].Transmitting {
						sBusConns[i].SetWriteFrame(mergedFrame)
					}
				}
				slog.Debug("frame sent", "frame", mergedFrame, "time_to_update", time.Since(startTime))
			}
		}
	})

	//kill listener
	group.Go(func() error {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		for {
			select {
			case sig := <-signalChannel:
				slog.Info("received signal", "value", sig)
				cancel()
				return fmt.Errorf("received signal: %s", sig)
			case <-ctx.Done():
				slog.Info("closing signal goroutine", "error", ctx.Err().Error())
				return ctx.Err()
			}
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
