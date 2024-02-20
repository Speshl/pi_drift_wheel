package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Speshl/pi_drift_wheel/config"
	"github.com/Speshl/pi_drift_wheel/controllers"
	"github.com/Speshl/pi_drift_wheel/controllers/models"
	"github.com/Speshl/pi_drift_wheel/crsf"
	sbus "github.com/Speshl/pi_drift_wheel/sbus"
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

	controllerManager *controllers.ControllerManager
	sBusConns         []*sbus.SBus
	crsfConns         []*crsf.CRSF

	setMinPitch int
	setMidPitch int
	setMaxPitch int

	// feedback       int
	// mappedFeedback int
	// diffFeedback   float64
	ffLevel   float64
	ffEnabled bool
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

	a.startControllers(ctx, group, cancel)

	a.startSbus(ctx, group, cancel)

	a.startCRSF(ctx, group, cancel)

	a.startKillListener(ctx, group, cancel)

	group.Go(func() error {
		return a.processData(ctx)
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

/*
Reads Sbus RX and Controller Inputs to merge into a single sbus frame to be sent to all Sbus Tx.  Also reads CRSF telemetry to get feedback for force feedback.
*/
func (a *App) processData(ctx context.Context) error {
	slog.Info("start processing")
	defer slog.Info("stopping processing")

	time.Sleep(500 * time.Millisecond) //give some time for signals to warm up

	mergeTime := 7 * time.Millisecond
	mergeTicker := time.NewTicker(mergeTime)
	//mergeTicker := time.NewTicker(1 * time.Second) //Slow ticker

	// logTicker := time.NewTicker(100 * time.Millisecond) //fast logger
	//logTicker := time.NewTicker(1 * time.Second) //slow logger

	// ffTicker := time.NewTicker(60 * time.Millisecond)

	lastWriteTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		// case <-ffTicker.C: //TODO: might not need own ticket, needs testing with another wheel
		// 	if a.ffEnabled {
		// 		ffLevel := int16(a.ffLevel * (65535 / 2))
		// 		if ffLevel != int16(lastFFLevel) {
		// 			//controllerManager.SetForceFeedback(ffLevel)
		// 		}
		// 		lastFFLevel = ffLevel
		// 	}

		case <-mergeTicker.C:
			//gather inputs and combine all into a single frame
			mixedFrame, mixedController, err := a.gatherInputs()
			if err != nil {
				slog.Error("error gathering inputs", "error", err)
				continue //Might need return here
			}

			//do anything we need to at this point with the comibined input frame
			a.utilizeInputs(mixedFrame, mixedController)

			//finally send the combined frame to all sbus tx
			a.sendOutputs(mixedFrame)

			if time.Since(lastWriteTime) > (5*time.Millisecond)+mergeTime {
				slog.Warn("slow processing", "duration", time.Since(lastWriteTime))
			}
			lastWriteTime = time.Now()

			slog.Debug("details",
				"steer", mixedFrame.Frame.Ch[0],
				"esc", mixedFrame.Frame.Ch[1],
				"gyro_gain", mixedFrame.Frame.Ch[2],
				"tilt", mixedFrame.Frame.Ch[3],
				"roll", mixedFrame.Frame.Ch[4],
				"pan", mixedFrame.Frame.Ch[5],
				"levelFromFeedback", a.ffLevel,
			)
		}
	}
}

func (a *App) gatherInputs() (sbus.SBusFrame, models.MixState, error) {
	framesToMerge := make([]sbus.SBusFrame, 0, len(a.controllerManager.Controllers)+len(a.sBusConns))

	controllerFrame, err := a.controllerManager.GetMixedFrame() //Get one frame that has been pre-mixed from all connected controllers
	if err != nil {
		return sbus.NewSBusFrame(), a.controllerManager.GetMixState(), fmt.Errorf("error getting mixed frame - %w", err)
	}
	framesToMerge = append(framesToMerge, controllerFrame)

	for i := range a.sBusConns { //Get a frame from each sbus connection that is labeled as a control device
		if a.sBusConns[i].IsReceiving() && a.sBusConns[i].Type() == sbus.RxTypeControl {

			readFrame := a.sBusConns[i].GetReadFrame()
			newFrame := sbus.NewSBusFrame()
			for _, j := range a.cfg.SbusCfgs[i].SBusChannels { //Only pull over values we care about
				newFrame.Frame.Ch[j] = readFrame.Ch[j]
			}
			slog.Debug("sbus frame", "port", i, "channels", a.cfg.SbusCfgs[i].SBusChannels, "read", readFrame, "newFrame", newFrame)
			framesToMerge = append(framesToMerge, newFrame)
		} else if a.sBusConns[i].IsReceiving() && a.sBusConns[i].Type() == sbus.RxTypeTelemetry {
			slog.Info("sbus telemetry", "frame", a.sBusConns[i].GetReadFrame())
		}
	}

	return MergeFrames(framesToMerge), a.controllerManager.GetMixState(), nil
}

func (a *App) utilizeInputs(inputFrame sbus.SBusFrame, controlState models.MixState) {
	//Do anything we need to do with the input frame here
	//crsf device 0 attitude telemetry is used for feedback
	attitude := a.crsfConns[0].GetAttitude() //Todo: always using first crsf
	a.ffLevel = calculateFFLevel(a.setMinPitch, a.setMidPitch, a.setMaxPitch, int(attitude.Pitch), int(inputFrame.Frame.Ch[0]))

	red1 := controlState.Buttons["red1"]
	lrValue := controlState.Buttons["left/right"]
	udValue := controlState.Buttons["up/down"]

	if red1 == 1 {
		if lrValue > 0 { //Set right end point
			a.ffEnabled = true
			a.setMaxPitch = int(attitude.Pitch)
			slog.Info("setting max (right) ff endpoint", "max_pitch", a.setMaxPitch)
		} else if lrValue < 0 { //Set left end point
			a.ffEnabled = true
			a.setMinPitch = int(attitude.Pitch)
			slog.Info("setting min (left) ff endpoint", "min_pitch", a.setMinPitch)
		} else if udValue > 0 { //Set left end point
			a.ffEnabled = true
			a.setMidPitch = int(attitude.Pitch)
			slog.Info("setting mid (center) ff endpoint", "mid_pitch", a.setMidPitch)
		} else {
			a.ffEnabled = false
		}
	} else {
		a.ffEnabled = false
	}

}

func (a *App) sendOutputs(mixedFrame sbus.SBusFrame) {
	mixedFrame = InvertChannels(mixedFrame, a.cfg.AppCfg.InvertOutputs)
	for i := range a.sBusConns {
		if a.sBusConns[i].IsTransmitting() {
			a.sBusConns[i].SetWriteFrame(mixedFrame)
		}
	}
}
