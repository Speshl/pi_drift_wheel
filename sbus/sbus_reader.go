package sbus

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/albenik/go-serial/v2"

	"github.com/Speshl/pi_drift_wheel/config"
)

type SBusReader struct {
	lock  sync.RWMutex
	Path  string
	frame Frame
}

func NewSBusReader(cfg config.SBusConfig) *SBusReader {
	return &SBusReader{
		Path: cfg.SBusInPath,
	}
}

func (r *SBusReader) Start(ctx context.Context) error {
	port, err := serial.Open(r.Path,
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
				slog.Debug("appending")
				frame = append(frame, buff[i])
				if len(frame) >= frameLength {
					midFrame = false
					if buff[i] == endByte { //this is a complete frame
						frame, err := UnmarshalFrame([25]byte(frame))
						if err != nil {
							slog.Warn("frame should have parsed but failed", "error", err)
						}
						r.lock.Lock()
						r.frame = frame //set the latest frame
						r.lock.Unlock()
					} else {
						slog.Debug("found frame start but not frame end")
					}
				} else {
					slog.Debug("building frame", "length", len(frame))
				}
			} else if int(buff[i]) == int(startByte) { //Looking for the start of the next frame
				midFrame = true
				frame = append(frame[:0], buff[i])
				slog.Debug("found a match", "length", len(frame))
			} else {
				slog.Debug("outside frame, but didn't find start")
			}
		}
		slog.Debug("read", "num_read", n, "data", buff[:n])
	}
}

func (r *SBusReader) GetLatestFrame() Frame {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.frame
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
