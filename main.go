package main

import (
	"context"
	"log/slog"

	"github.com/Speshl/pi_drift_wheel/app"
	"github.com/Speshl/pi_drift_wheel/config"
)

// go build --ldflags '-extldflags "-Wl,--allow-multiple-definition"'
func main() {
	cfg := config.GetConfig()

	app := app.NewApp(cfg)

	err := app.Start(context.Background())
	if err != nil {
		slog.Error("client shutdown with error", "error", err.Error())
	} else {
		slog.Info("client shutdown successfully")
	}
}

//https://www.kernel.org/doc/html/v4.12/input/ff.html
