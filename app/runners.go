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
	"github.com/Speshl/pi_drift_wheel/controllers"
	"github.com/Speshl/pi_drift_wheel/controllers/models"
	"github.com/Speshl/pi_drift_wheel/crsf"
	sbus "github.com/Speshl/pi_drift_wheel/sbus"
	"golang.org/x/sync/errgroup"
)

func (a *App) startControllers(ctx context.Context, group *errgroup.Group, cancel context.CancelFunc) error {
	a.controllerManager = controllers.NewControllerManager(a.cfg.ControllerManagerCfg, models.ControllerOptions{UseHPattern: true})
	err := a.controllerManager.LoadControllers()
	if err != nil {
		return fmt.Errorf("failed loading controllers: %w", err)
	}
	group.Go(func() error {
		defer cancel()
		slog.Info("starting controller manager")
		defer slog.Info("stopping controller manager")
		return a.controllerManager.Start(ctx)
	})
	return nil
}

func (a *App) startSbus(ctx context.Context, group *errgroup.Group, cancel context.CancelFunc) error {
	a.sBusConns = make([]*sbus.SBus, 0, config.MaxSbus)
	for i := 0; i < config.MaxSbus; i++ {
		i := i
		sBus, err := sbus.NewSBus(
			a.cfg.SbusCfgs[i].SBusPath,
			a.cfg.SbusCfgs[i].SBusRx,
			a.cfg.SbusCfgs[i].SBusTx,
			&sbus.SBusCfgOpts{
				Type: sbus.RxTypeControl,
			},
		)
		if err != nil { //TODO: Remove when more channels supported
			if !errors.Is(err, sbus.ErrNoPath) {
				slog.Error("failed starting sbus conn", "index", i, "error", err)
			}
			continue
		}

		a.sBusConns = append(a.sBusConns, sBus)
		group.Go(func() error {
			defer cancel()
			err := ListPorts()
			if err != nil {
				return err
			}
			slog.Info("starting sbus", "index", i, "path", a.cfg.SbusCfgs[i].SBusPath)
			defer slog.Info("stopping sbus", "index", i, "path", a.cfg.SbusCfgs[i].SBusPath)
			return sBus.Start(ctx)
		})
	}
	return nil
}

func (a *App) startCRSF(ctx context.Context, group *errgroup.Group, cancel context.CancelFunc) {
	// dmesg | grep "tty"
	a.crsfConns = make([]*crsf.CRSF, 0, config.MaxCRSF)
	for i := 0; i < config.MaxCRSF; i++ {
		i := i
		crsf := crsf.NewCRSF(
			a.cfg.CRSFCfgs[i].CRSFPath,
			&crsf.CRSFOptions{
				BaudRate: config.CRSFBaudRate,
			},
		)

		a.crsfConns = append(a.crsfConns, crsf)
		group.Go(func() error {
			defer cancel()
			//TODO: List ports for crsf
			slog.Info("starting crsf", "index", i, "path", a.cfg.CRSFCfgs[i].CRSFPath)
			defer slog.Info("stopping crsf", "index", i, "path", a.cfg.CRSFCfgs[i].CRSFPath)
			return crsf.Start(ctx)
		})
	}
}

func (a *App) startKillListener(ctx context.Context, group *errgroup.Group, cancel context.CancelFunc) {
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
}
