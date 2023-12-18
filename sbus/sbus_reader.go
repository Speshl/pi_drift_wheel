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

// func (r *SBusReader) Start(ctx context.Context) error {
// 	port, err := r.open()
// 	if err != nil {
// 		return err
// 	}
// 	defer func() {
// 		if port != nil {
// 			slog.Info("closing serial port", "path", r.Path)
// 			port.Close()
// 		}
// 	}()

// 	readBuffer := make([]byte, 0, 64)
// 	dataChannel := make(chan []byte, 100)

// 	slog.Info("start reading serial")

// 	//port.SetReadTimeout(5 * time.Second)
// 	go func() {
// 		defer close(dataChannel)
// 		for {
// 			if ctx.Err() != nil {
// 				slog.Info("serial channel reader context cancelled")
// 				return //ctx.Err()
// 			}
// 			if port == nil {
// 				slog.Info("serial  port is nil")
// 				return
// 			}

// 			clear(readBuffer)
// 			numRead, err := port.Read(readBuffer)
// 			if err != nil {
// 				slog.Error("failed reading from serial", "error", err.Error())
// 				return //fmt.Errorf("failed reading from serial - %w", err)
// 			}
// 			slog.Info("read bytes", "numRead", numRead, "bytes", readBuffer[0:numRead])
// 			dataChannel <- readBuffer
// 		}
// 	}()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		case data, ok := <-dataChannel:
// 			if !ok {
// 				slog.Info("sbus reader channel closed")
// 				return nil
// 			}
// 			slog.Info("got data", "length", len(data), "data", data)
// 		}
// 	}

// }

func (r *SBusReader) Start2(ctx context.Context) error {
	port, err := serial.Open("/dev/ttyAMA0",
		serial.WithBaudrate(115200),
		serial.WithDataBits(8),
		// serial.WithParity(serial.NoParity),
		// serial.WithStopBits(serial.OneStopBit),
		serial.WithReadTimeout(1000),
		// serial.WithWriteTimeout(1000),
		// serial.WithHUPCL(false),
	)
	if err != nil {
		return err
	}

	buff := make([]byte, 100)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// Reads up to 100 bytes
		n, err := port.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		if n == 0 {
			slog.Info("serial eof")
			//break
		}
		slog.Info("read", "data", buff[:n])
	}
}

// func (r *SBusReader) open() (serial.Port, error) {
// 	// mode := &serial.Mode{
// 	// 	BaudRate: r.Baud,
// 	// 	// Parity:   serial.EvenParity,
// 	// 	// StopBits: serial.TwoStopBits,
// 	// 	DataBits: 8,
// 	// }
// 	slog.Info("opening serial connection", "path", r.Path)
// 	port, err := serial.Open(r.Path, &serial.Mode{
// 		BaudRate: 115200,
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed opening serial connection - %w", err)
// 	}
// 	slog.Info("serial connection opened:", "path", r.Path)
// 	return port, nil
// }

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
