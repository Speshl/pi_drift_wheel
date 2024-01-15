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
	DefaultMinPitch = 300 //300 //50
	DefaultMidPitch = 500
	DefaultMaxPitch = 750 //750 //950

	DefaultMinYaw = -180 //102 / 117
	DefaultMidYaw = 0
	DefaultMaxYaw = 180 //-124/109
)

type App struct {
	cfg config.Config

	setMinPitch int
	setMidPitch int
	setMaxPitch int

	feedback       int
	mappedFeedback int
	diffFeedback   float64
	feedbackLevel  float64
}

func NewApp(cfg config.Config) *App {
	return &App{
		cfg:         cfg,
		setMinPitch: DefaultMinPitch,
		setMidPitch: DefaultMidPitch,
		setMaxPitch: DefaultMaxPitch,
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
		slog.Info("starting controller manager")
		defer slog.Info("stopping controller manager")
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
			defer slog.Info("stopping sbus", "index", i, "path", a.cfg.SbusCfgs[i].SBusPath)
			return sBus.Start(ctx)
		})
	}

	// Start CRSF read/write
	// dmesg | grep "tty"
	crsf := crsf.NewCRSF("/dev/ttyACM0", &crsf.CRSFOptions{ //controller = /dev/ttyACM0 //module = /dev/ttyUSB0
		BaudRate: 921600,
	})
	group.Go(func() error {
		slog.Info("starting crsf", "path", "/dev/ttyACM0")
		defer slog.Info("stopping crsf", "path", "/dev/ttyACM0")
		return crsf.Start(ctx)
	})

	//Process data
	group.Go(func() error {
		slog.Info("start processing")
		defer slog.Info("stopping processing")

		time.Sleep(500 * time.Millisecond) //give some time for signals to warm up

		mergeTicker := time.NewTicker(7 * time.Millisecond)
		//mergeTicker := time.NewTicker(1 * time.Second) //Slow ticker
		logTicker := time.NewTicker(100 * time.Millisecond) //fast logger
		//logTicker := time.NewTicker(1 * time.Second) //slow logger
		ffTicker := time.NewTicker(60 * time.Millisecond)

		framesToMerge := make([]sbus.SBusFrame, 0, len(controllerManager.Controllers)+len(sBusConns))
		mergedFrame := sbus.NewSBusFrame()
		disableFF := false

		lastWriteTime := time.Now()
		lastFFLevel := int16(0)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-logTicker.C:
				slog.Info("details",
					"steer", mergedFrame.Frame.Ch[0],
					// "esc", mergedFrame.Ch[1],
					// "gyro_gain", mergedFrame.Ch[2],
					// "tilt", mergedFrame.Ch[3],
					// "roll", mergedFrame.Ch[4],
					// "pan", mergedFrame.Ch[5],
					"mappedFeedback", a.mappedFeedback,
					"diffFeedback", a.diffFeedback,
					// "feedback", a.feedback,
					"feedbackLevel", a.feedbackLevel,
					// "minPitch", a.setMinPitch,
					// "maxPitch", a.setMaxPitch,
				)
			case <-ffTicker.C:
				if !disableFF {
					ffLevel := int16(a.feedbackLevel * (65535 / 2))
					if ffLevel != int16(lastFFLevel) {
						controllerManager.SetForceFeedback(ffLevel)
					}
					lastFFLevel = ffLevel
				}

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
						newFrame := sbus.NewSBusFrame()
						for _, j := range a.cfg.SbusCfgs[i].SBusChannels { //Only pull over values we care about
							newFrame.Frame.Ch[j] = readFrame.Ch[j]
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
				a.feedback = int(attitude.Pitch)
				if a.feedback >= a.setMidPitch {
					a.mappedFeedback = controllers.MapToRange(a.feedback, a.setMidPitch, a.setMaxPitch, sbus.MidValue, sbus.MaxValue)
				} else {
					a.mappedFeedback = controllers.MapToRange(a.feedback, a.setMinPitch, a.setMidPitch, sbus.MinValue, sbus.MidValue)
				}

				diffPitch := int(mergedFrame.Frame.Ch[0]) - a.mappedFeedback
				a.diffFeedback = float64(diffPitch) / float64(sbus.MaxValue-sbus.MinValue)

				a.feedbackLevel = 0.0
				if a.diffFeedback > 0.01 || a.diffFeedback < -0.01 { //deadzone
					a.feedbackLevel = a.diffFeedback * 2
				}

				if a.feedbackLevel > 1.0 { //limit
					a.feedbackLevel = 1.0
				} else if a.feedbackLevel < -1.0 {
					a.feedbackLevel = -1.0
				}
				//end

				red1 := controlState.Buttons["red1"]
				lrValue := controlState.Buttons["left/right"]
				udValue := controlState.Buttons["up/down"]

				if red1 == 1 {
					if lrValue > 0 { //Set right end point
						disableFF = true
						a.setMaxPitch = a.feedback
						slog.Info("setting max (right) ff endpoint", "max_mapped_pitch", a.mappedFeedback, "max_pitch", a.feedback)
					} else if lrValue < 0 { //Set left end point
						disableFF = true
						a.setMinPitch = a.feedback
						slog.Info("setting min (left) ff endpoint", "min_mapped_pitch", a.mappedFeedback, "min_pitch", a.feedback)
					} else if udValue > 0 { //Set left end point
						disableFF = true
						a.setMidPitch = a.feedback
						slog.Info("setting mid (center) ff endpoint", "mid_mapped_pitch", a.mappedFeedback, "mid_pitch", a.feedback)
					} else {
						disableFF = false
					}
				} else {
					disableFF = false
				}

				//Output
				mergedFrame = InvertChannels(mergedFrame, a.cfg.AppCfg.InvertOutputs)

				if time.Since(lastWriteTime) > (11 * time.Millisecond) {
					slog.Warn("slow processing", "duration", time.Since(lastWriteTime))
				}
				lastWriteTime = time.Now()

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
func MergeFrames(frames []sbus.SBusFrame) sbus.SBusFrame {
	if len(frames) == 0 {
		return sbus.NewSBusFrame()
	}

	mergedFrame := frames[0]
	for i := range frames {
		for j := range frames[i].Frame.Ch {
			mergedDistFromMid := math.Abs(float64(mergedFrame.Frame.Ch[j]) - float64(sbus.MidValue))
			frameDisFromMid := math.Abs(float64(frames[i].Frame.Ch[j]) - float64(sbus.MidValue))

			if frameDisFromMid > mergedDistFromMid {
				mergedFrame.Frame.Ch[j] = frames[i].Frame.Ch[j]
			}
		}
	}
	return mergedFrame
}

func InvertChannels(inputFrame sbus.SBusFrame, invertChannels []bool) sbus.SBusFrame {
	returnFrame := inputFrame
	for i := range invertChannels {
		if invertChannels[i] {
			midOffset := returnFrame.Frame.Ch[i] - uint16(sbus.MidValue)
			inverted := uint16(sbus.MidValue) - midOffset
			returnFrame.Frame.Ch[i] = inverted
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
