package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Speshl/pi_drift_wheel/config"
	"github.com/Speshl/pi_drift_wheel/controllers"
	sbus "github.com/Speshl/pi_drift_wheel/sbus"
	"github.com/albenik/go-serial/v2"
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
	controllerManager := controllers.NewControllerManager(a.cfg.ControllerManagerCfg, controllers.WheelMixer, controllers.ControllerOptions{UseHPattern: true})
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

		sBusConns = append(sBusConns, sBus)
		group.Go(func() error {
			defer cancel()
			err := ListPorts()
			if err != nil {
				return err
			}
			slog.Info("starting sbus", "index", i, "path", a.cfg.SbusCfgs[i].SBusPath)
			return sBus.Start(ctx)
		})
	}

	//Start CRSF read/write
	//dmesg | grep "tty"
	// crsf := crsf.NewCRSF("/dev/ttyACM0", &crsf.CRSFOptions{ //controller = /dev/ttyACM0 //module = /dev/ttyUSB0
	// 	BaudRate: 921600,
	// })
	// group.Go(func() error {
	// 	return crsf.Start(ctx)
	// })

	//Process data
	group.Go(func() error {
		time.Sleep(500 * time.Millisecond) //give some time for signals to warm up
		framesToMerge := make([]sbus.Frame, 0, len(controllerManager.Controllers)+len(sBusConns))
		mergeTicker := time.NewTicker(6 * time.Millisecond)
		//mergeTicker := time.NewTicker(1 * time.Second) //Slow ticker
		//logTicker := time.NewTicker(5 * time.Second)
		mergedFrame := sbus.NewFrame()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			// case <-logTicker.C:
			// 	slog.Info("frame details",
			// 		"steer", mergedFrame.Ch[0],
			// 		"esc", mergedFrame.Ch[1],
			// 		"gyro_gain", mergedFrame.Ch[2],
			// 		"head_tilt", mergedFrame.Ch[3],
			// 		"head_roll", mergedFrame.Ch[4],
			// 		"head_pan", mergedFrame.Ch[5],
			// 	)
			case <-mergeTicker.C:
				framesToMerge = framesToMerge[:0] //clear out frames before next merge

				controllerFrame, err := controllerManager.GetMixedFrame()
				if err != nil {
					return err
				}
				framesToMerge = append(framesToMerge, controllerFrame)

				for i := range sBusConns {
					if sBusConns[i].IsReceiving() && sBusConns[i].Type() == sbus.RxTypeControl {

						readFrame := sBusConns[i].GetReadFrame()
						newFrame := sbus.NewFrame()
						for _, j := range a.cfg.SbusCfgs[i].SBusChannels { //Only pull over values we care about
							newFrame.Ch[j] = readFrame.Ch[j]
						}
						slog.Debug("sbus frame", "port", i, "channels", a.cfg.SbusCfgs[i].SBusChannels, "read", readFrame, "newFrame", newFrame)
						framesToMerge = append(framesToMerge, newFrame)
					} else if sBusConns[i].IsReceiving() && sBusConns[i].Type() == sbus.RxTypeTelemetry {
						slog.Info("sbus telemetry", "frame", sBusConns[i].GetReadFrame())
					}
				}
				mergedFrame = MergeFrames(framesToMerge)
				for i := range sBusConns {
					if sBusConns[i].IsTransmitting() {
						sBusConns[i].SetWriteFrame(mergedFrame)
					}
				}

				// attitude := crsf.GetAttitude()
				// yaw := int(attitude.YawDegree()) //expect value between -90 and 90
				// mappedYaw := controllers.MapToRange(yaw, -90, 90, sbus.MinValue, sbus.MaxValue)
				// diff := int(mergedFrame.Ch[1]) - mappedYaw
				// diffPercent := float64(diff) / float64(sbus.MaxValue-sbus.MinValue)

				// level := 0.0
				// if diffPercent > 0.03 || diffPercent < -0.03 {
				// 	level = diffPercent * 0.75
				// }

				// if level > 1.0 {
				// 	level = 1.0
				// } else if level < -1.0 {
				// 	level = -1.0
				// }

				// controllerManager.SetForceFeedback(int16(level * (65535 / 2)))

				// if level < 0 {
				// 	slog.Info("ff info", "direction", "left", "level", level, "mappedYaw", mappedYaw, "steer", mergedFrame.Ch[1])
				// } else {
				// 	slog.Info("ff info", "direction", "right", "level", level, "mappedYaw", mappedYaw, "steer", mergedFrame.Ch[1])
				// }

				//slog.Debug("FF Extra Info", "yaw", yaw, "steer", mergedFrame.Ch[1], "mappedYaw", mappedYaw, "diff", diff, "percent", diffPercent, "level", level)
				slog.Debug("frame sent", "frame", mergedFrame)
				slog.Info("details",
					"steer", mergedFrame.Ch[0],
					"esc", mergedFrame.Ch[1],
					"gyro_gain", mergedFrame.Ch[2],
					"tilt", mergedFrame.Ch[3],
					"roll", mergedFrame.Ch[4],
					"pan", mergedFrame.Ch[5],
				)

			}
		}
	})

	// Test Force feedback
	// group.Go(func() error {
	// 	time.Sleep(500 * time.Millisecond) //give some time for signals to warm up
	// 	logTicker := time.NewTicker(250 * time.Millisecond)
	// 	dir := 1
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return ctx.Err()
	// 		case <-logTicker.C:
	// 			slog.Info("sending FF")
	// 			dir = dir * -1
	// 			err := controllerManager.SetForceFeedback(int16(dir * (65535 / 2)))
	// 			if err != nil {
	// 				slog.Error("ff error", "error", err)
	// 				return err
	// 			}

	// 		}
	// 	}
	// })

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
			slog.Info("app context was cancelled", "error", err)
			return nil
		} else {
			return fmt.Errorf("app stopping due to error - %w", err)
		}
	}
	return nil
}

// Merge all provided frames into 1 frame. Use the furthest channel value from the midpoint of each frame
func MergeFrames(frames []sbus.Frame) sbus.Frame {
	if len(frames) == 0 {
		return sbus.NewFrame()
	}

	mergedFrame := frames[0]
	for i := range frames {
		for j := range frames[i].Ch {
			mergedDistFromMid := math.Abs(float64(mergedFrame.Ch[j]) - float64(sbus.MidValue))
			frameDisFromMid := math.Abs(float64(frames[i].Ch[j]) - float64(sbus.MidValue))

			if frameDisFromMid > mergedDistFromMid {
				mergedFrame.Ch[j] = frames[i].Ch[j]
			}
		}
	}
	return mergedFrame
}

func ListPorts() error {
	ports, err := serial.GetPortsList()
	if err != nil {
		return err
	}
	if len(ports) == 0 {
		return fmt.Errorf("no serial ports found")
	}
	for _, port := range ports {
		slog.Info("found port", "port", port)
	}
	return nil
}
