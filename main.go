package main

import (
	"context"
	"log"

	"github.com/Speshl/pi_drift_wheel/app"
	"github.com/Speshl/pi_drift_wheel/config"
)

func main() {
	cfg := config.GetConfig()

	app := app.NewApp(cfg)

	err := app.Start(context.Background())
	if err != nil {
		log.Printf("client shutdown with error: %s", err.Error())
	} else {
		log.Println("client shutdown successfully")
	}
}
