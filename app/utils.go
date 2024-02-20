package app

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/Speshl/pi_drift_wheel/controllers/models"
	sbus "github.com/Speshl/pi_drift_wheel/sbus"
	"github.com/albenik/go-serial/v2"
)

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

func calculateFFLevel(min, mid, max, feedback, newSteer int) float64 {
	//Get a ff level from the servo feedback
	var mappedFeedback int
	if feedback >= mid {
		mappedFeedback = models.MapToRange(feedback, mid, max, sbus.MidValue, sbus.MaxValue)
	} else {
		mappedFeedback = models.MapToRange(feedback, min, mid, sbus.MinValue, sbus.MidValue)
	}

	diffPitch := newSteer - mappedFeedback
	diffFeedback := float64(diffPitch) / float64(sbus.MaxValue-sbus.MinValue)

	feedbackLevel := 0.0
	if diffFeedback > 0.01 || diffFeedback < -0.01 { //deadzone
		feedbackLevel = diffFeedback * 2
	}

	if feedbackLevel > 1.0 { //limit
		feedbackLevel = 1.0
	} else if feedbackLevel < -1.0 {
		feedbackLevel = -1.0
	}
	return feedbackLevel
}
