package sbus

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"go.bug.st/serial"

	"github.com/Speshl/pi_drift_wheel/config"
)

type SBusReader struct {
	Port serial.Port
	Path string
	Baud int
}

func NewSBusReader(cfg config.SBusConfig) *SBusReader {
	return &SBusReader{
		Path: cfg.SBusInPath,
		Baud: cfg.SBusInBaud,
	}
}

func (r *SBusReader) Cleanup() error {
	if r.Port != nil {
		slog.Info("closing serial port", "path", r.Path)
		return r.Port.Close()
	}
	return nil
}

func (r *SBusReader) Start(ctx context.Context) error {
	err := r.Open()
	if err != nil {
		return err
	}
	defer r.Cleanup()

	//dataBuffer := make([]byte, 0, 64)
	readBuffer := make([]byte, 0, 64)
	dataChannel := make(chan []byte, 100)

	slog.Info("start reading serial")

	go func() {
		defer close(dataChannel)
		for {
			if ctx.Err() != nil {
				return //ctx.Err()
			}
			if r.Port == nil {
				return //fmt.Errorf("port is nil")
			}

			clear(readBuffer)
			numRead, err := r.Port.Read(readBuffer)
			if err != nil {
				slog.Error("failed reading from serial", "error", err)
				return //fmt.Errorf("failed reading from serial - %w", err)
			}
			slog.Info("read bytes", "numRead", numRead, "bytes", readBuffer[0:numRead])
			dataChannel <- readBuffer
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data, ok := <-dataChannel:
			if !ok {
				slog.Info("sbus reader channel closed")
				return nil
			}
			slog.Info("got data", "length", len(data), "data", data)
		}
	}

}

func (r *SBusReader) Open() (err error) {
	mode := &serial.Mode{
		BaudRate: r.Baud,
		// Parity:   serial.EvenParity,
		// StopBits: serial.TwoStopBits,
		// DataBits: 8,
	}
	slog.Info("opening serial connection", "path", r.Path)
	r.Port, err = serial.Open(r.Path, mode)
	if err != nil {
		return fmt.Errorf("failed opening serial connection - %w", err)
	}
	slog.Info("serial connection opened:", "path", r.Path)
	return nil
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
