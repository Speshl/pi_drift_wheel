package sbus

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/albenik/go-serial/v2"
	//"go.bug.st/serial"

	"github.com/Speshl/pi_drift_wheel/config"
)

type SBusReader struct {
	Path string
	Baud int
}

func NewSBusReader(cfg config.SBusConfig) *SBusReader {
	return &SBusReader{
		Path: cfg.SBusInPath,
		Baud: cfg.SBusInBaud,
	}
}

func (r *SBusReader) Start2(ctx context.Context) error {
	port, err := serial.Open("/dev/ttyAMA0",
		serial.WithBaudrate(100000),
		serial.WithDataBits(8),
		serial.WithParity(serial.EvenParity),
		serial.WithStopBits(serial.TwoStopBits),
		serial.WithReadTimeout(1000),
		// serial.WithWriteTimeout(1000),
		// serial.WithHUPCL(false),
	)
	if err != nil {
		return err
	}

	buff := make([]byte, 25)
	frame := make([]byte, 0, 25)
	midFrame := false
	for {
		clear(buff)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		n, err := port.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		for i := range buff[:n] {
			if midFrame { //already found start byte so looking for end byte
				slog.Info("appending")
				frame = append(frame, buff[i])
				if len(frame) >= framelength {
					midFrame = false
					if buff[i] == endbyte { //this is a complete frame
						frame, err := UnmarshalFrame([25]byte(frame))
						if err != nil {
							slog.Warn("frame should have parsed but failed", "error", err)
						}
						slog.Info("found frame", "frame", frame)
						//do something with the read frame
					} else {
						slog.Warn("found frame start but not frame end")
					}
				} else {
					slog.Info("building frame", "length", len(frame))
				}
			} else if int(buff[i]) == int(startbyte) { //Looking for the start of the next frame
				midFrame = true
				frame = append(frame[:0], buff[i])
				slog.Info("found a match", "length", len(frame))
			} else {
				slog.Info("outside frame, but didn't find start")
			}
		}
		slog.Info("read", "num_read", n, "data", buff[:n])
	}
}

func (r *SBusReader) ListPorts() error {
	ports, err := serial.GetPortsList()
	if err != nil {
		return err
	}
	if len(ports) == 0 {
		return fmt.Errorf("no serial ports found")
	}
	for _, port := range ports {
		log.Printf("found port: %v\n", port)
	}
	return nil
}
