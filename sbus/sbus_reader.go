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
	Path string
	Baud int
}

func NewSBusReader(cfg config.SBusConfig) *SBusReader {
	return &SBusReader{
		Path: cfg.SBusInPath,
		Baud: cfg.SBusInBaud,
	}
}

func (r *SBusReader) Start(ctx context.Context) error {
	port, err := r.open()
	if err != nil {
		return err
	}
	defer port.Close() //Can error

	//dataBuffer := make([]byte, 0, 64)
	readBuffer := make([]byte, 0, 64)
	slog.Info("start reading serial")
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		clear(readBuffer)
		numRead, err := port.Read(readBuffer)
		if err != nil {
			return fmt.Errorf("failed reading from serial - %w", err)
		}
		log.Printf("read %d bytes", numRead)
	}
}

func (r *SBusReader) open() (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: r.Baud,
		Parity:   serial.EvenParity,
		StopBits: serial.TwoStopBits,
		DataBits: 8,
	}
	port, err := serial.Open(r.Path, mode)
	if err != nil {
		return nil, fmt.Errorf("failed opening serial connection - %w", err)
	}
	slog.Info("serial connection opened:", "path", r.Path)
	return port, err
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
