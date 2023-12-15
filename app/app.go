package app

import (
	"context"
	"fmt"

	"github.com/Speshl/pi_drift_wheel/config"
	"github.com/Speshl/pi_drift_wheel/controllers"
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
	//group, _ := errgroup.WithContext(ctx)

	controllerManager := controllers.NewControllerManager(a.cfg.ControllerManagerCfg)
	err := controllerManager.LoadControllers()
	if err != nil {
		return fmt.Errorf("failed loading controllers: %w", err)
	}

	// err := group.Wait()
	// if err != nil {
	// 	if errors.Is(err, context.Canceled) {
	// 		log.Println("context was cancelled")
	// 		return nil
	// 	} else {
	// 		return fmt.Errorf("server stopping due to error - %w", err)
	// 	}
	// }
	return nil
}
