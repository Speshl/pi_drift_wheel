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
	"github.com/Speshl/pi_drift_wheel/crsf"
	sbus "github.com/Speshl/pi_drift_wheel/sbus"
	"github.com/albenik/go-serial/v2"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultMinPitch = 0
	DefaultMaxPitch = 0

	DefaultMinYaw = 0
	DefaultMaxYaw = 0
)

type App struct {
	cfg config.Config

	setMinPitch int
	setMaxPitch int

	setMinYaw int
	setMaxYaw int

	gyro       int
	mappedGyro int
	diffGyro   float64
	gyroLevel  float64

	feedback       int
	mappedFeedback int
	diffFeedback   float64
	feedbackLevel  float64
}

func NewApp(cfg config.Config) *App {
	return &App{
		cfg:         cfg,
		setMinPitch: DefaultMinPitch,
		setMaxPitch: DefaultMaxPitch,
		setMinYaw:   DefaultMinYaw,
		setMaxYaw:   DefaultMaxYaw,
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

	// Start CRSF read/write
	// dmesg | grep "tty"
	crsf := crsf.NewCRSF("/dev/ttyACM0", &crsf.CRSFOptions{ //controller = /dev/ttyACM0 //module = /dev/ttyUSB0
		BaudRate: 921600,
	})
	group.Go(func() error {
		return crsf.Start(ctx)
	})

	//Process data
	group.Go(func() error {
		time.Sleep(500 * time.Millisecond) //give some time for signals to warm up
		framesToMerge := make([]sbus.Frame, 0, len(controllerManager.Controllers)+len(sBusConns))
		mergeTicker := time.NewTicker(6 * time.Millisecond)
		//mergeTicker := time.NewTicker(1 * time.Second) //Slow ticker
		logTicker := time.NewTicker(1 * time.Second)
		mergedFrame := sbus.NewFrame()

		disableFF := false

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-logTicker.C:
				slog.Info("details",
					"steer", mergedFrame.Ch[0],
					"esc", mergedFrame.Ch[1],
					"gyro_gain", mergedFrame.Ch[2],
					"tilt", mergedFrame.Ch[3],
					"roll", mergedFrame.Ch[4],
					"pan", mergedFrame.Ch[5],
					"gyroLevel", a.gyroLevel,
					"mappedGyro", a.mappedGyro,
					"diffGyro", a.diffGyro,
					"feedbackLevel", a.feedbackLevel,
					"mappedFeedback", a.mappedFeedback,
					"diffFeedback", a.diffFeedback,
					// "mappedPitch", mappedPitch,
					"minPitch", a.setMinPitch,
					"maxPitch", a.setMaxPitch,
					"yaw", a.gyro,
					// "mappedYaw", mappedYaw,
					"minYaw", a.setMinYaw,
					"maxYaw", a.setMaxYaw,
				)
			case <-mergeTicker.C:
				framesToMerge = framesToMerge[:0] //clear out frames before next merge
				//Input
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
				controlState := controllerManager.GetMixState()

				//Process
				attitude := crsf.GetAttitude()

				//Get a ff level from the servo feedback
				a.feedback = int(attitude.PitchDegree()) //expect value between -180 and 180
				a.mappedFeedback = controllers.MapToRange(a.feedback, -180, 180, sbus.MinValue, sbus.MaxValue)
				diffPitch := int(mergedFrame.Ch[1]) - a.mappedFeedback
				a.diffFeedback = float64(diffPitch) / float64(sbus.MaxValue-sbus.MinValue)

				a.feedbackLevel = 0.0
				if a.diffFeedback > 0.03 || a.diffFeedback < -0.03 {
					a.feedbackLevel = a.diffFeedback
				}

				if a.feedbackLevel > 1.0 {
					a.feedbackLevel = 1.0
				} else if a.feedbackLevel < -1.0 {
					a.feedbackLevel = -1.0
				}
				//end

				//Get a ff level from the gyro
				a.gyro = int(attitude.YawDegree()) //expect value between -180 and 180
				a.mappedGyro = controllers.MapToRange(a.gyro, -180, 180, sbus.MinValue, sbus.MaxValue)
				diffYaw := int(mergedFrame.Ch[1]) - a.mappedGyro
				a.diffGyro = float64(diffYaw) / float64(sbus.MaxValue-sbus.MinValue)

				a.gyroLevel = 0.0
				if a.diffGyro > 0.03 || a.diffGyro < -0.03 {
					a.gyroLevel = a.diffGyro
				}

				if a.gyroLevel > 1.0 {
					a.gyroLevel = 1.0
				} else if a.gyroLevel < -1.0 {
					a.gyroLevel = -1.0
				}
				//end

				red1 := controlState.Buttons["red1"]
				dpadValue := controlState.Buttons["left/right"]

				if red1 == 1 {
					if dpadValue > 0 { //Set right end point
						disableFF = true
						a.setMaxPitch = a.mappedFeedback
						a.setMaxYaw = a.mappedGyro
						slog.Info("setting max (right) ff endpoint")
					} else if dpadValue < 0 { //Set left end point
						disableFF = true
						a.setMinPitch = a.mappedFeedback
						a.setMinYaw = a.mappedGyro
						slog.Info("setting min (left) ff endpoint")
					} else {
						disableFF = false
					}
				} else {
					disableFF = false
				}

				if !disableFF {
					//controllerManager.SetForceFeedback(int16(levelFromGyro * (65535 / 2)))
				}

				//Output
				mergedFrame = InvertChannels(mergedFrame, a.cfg.AppCfg.InvertOutputs)

				for i := range sBusConns {
					if sBusConns[i].IsTransmitting() {
						sBusConns[i].SetWriteFrame(mergedFrame)
					}
				}

				// slog.Info("details",
				// 	"steer", mergedFrame.Ch[0],
				// 	"esc", mergedFrame.Ch[1],
				// 	"gyro_gain", mergedFrame.Ch[2],
				// 	"tilt", mergedFrame.Ch[3],
				// 	"roll", mergedFrame.Ch[4],
				// 	"pan", mergedFrame.Ch[5],
				// 	"levelFromGyro", a.gyroLevel,
				// 	"levelFromFeedback", a.feedbackLevel,
				// 	"pitch", pitch,
				// 	"mappedPitch", mappedPitch,
				// 	"minPitch", a.setMinPitch,
				// 	"maxPitch", a.setMaxPitch,
				// 	"yaw", yaw,
				// 	"mappedYaw", mappedYaw,
				// 	"minYaw", a.setMinYaw,
				// 	"maxYaw", a.setMaxYaw,
				// )
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

func InvertChannels(inputFrame sbus.Frame, invertChannels []bool) sbus.Frame {
	returnFrame := inputFrame
	for i := range invertChannels {
		if invertChannels[i] {
			midOffset := returnFrame.Ch[i] - uint16(sbus.MidValue)
			inverted := uint16(sbus.MidValue) - midOffset
			returnFrame.Ch[i] = inverted
		}
	}
	return returnFrame
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
