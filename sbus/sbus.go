package sbus

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/albenik/go-serial/v2"
	"golang.org/x/sync/errgroup"

	"github.com/Speshl/pi_drift_wheel/config"
)

var (
	ErrNoPath   = fmt.Errorf("no path provided for sbus")
	ErrNoAction = fmt.Errorf("no rx or tx action")

	RxTypeControl   = "control"
	RxTypeTelemetry = "telemetry"
)

type SBus struct {
	Path string

	Type string

	read      bool
	Recieving bool
	rxLock    sync.RWMutex
	rxFrame   Frame

	write        bool
	Transmitting bool
	txLock       sync.RWMutex
	txFrame      Frame
}

func NewSBus(cfg config.SBusConfig) (*SBus, error) {
	if cfg.SBusPath == "" {
		return nil, ErrNoPath
	}
	if !cfg.SBusRx && !cfg.SBusTx {
		return nil, ErrNoAction
	}
	return &SBus{
		Path:    cfg.SBusPath,
		Type:    cfg.SBusType,
		read:    cfg.SBusRx,
		rxFrame: NewFrame(),
		write:   cfg.SBusTx,
		txFrame: NewFrame(),
	}, nil
}

func (s *SBus) Start(ctx context.Context) error {
	port, err := serial.Open(s.Path,
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

	sbusGroup, ctx := errgroup.WithContext(ctx)

	sbusGroup.Go(func() error {
		return s.startReader(ctx, port)
	})

	sbusGroup.Go(func() error {
		return s.startWriter(ctx, port)
	})

	return sbusGroup.Wait()
}

func (s *SBus) startReader(ctx context.Context, port *serial.Port) error {
	if !s.read {
		return nil
	}
	s.Recieving = true

	slog.Info("start reading from sbus", "path", s.Path)
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
						frame, err := UnmarshalFrame(frame)
						if err != nil {
							slog.Warn("frame should have parsed but failed", "error", err)
						}
						s.rxLock.Lock()
						s.rxFrame = frame //set the latest frame
						s.rxLock.Unlock()
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

func (s *SBus) startWriter(ctx context.Context, port *serial.Port) error {
	if !s.write {
		return nil
	}
	s.Transmitting = true

	slog.Info("start writing to sbus", "path", s.Path)
	ticker := time.NewTicker(6 * time.Millisecond)
	var writeBytes []byte
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.txLock.RLock()
			writeBytes = s.txFrame.Marshal()
			s.txLock.RUnlock()
			n, err := port.Write(writeBytes)
			if err != nil {
				return err
			}
			if n != len(writeBytes) {
				slog.Info("sbus write incorrect length")
			}

		}
	}
}

func (s *SBus) GetReadFrame() Frame {
	s.rxLock.RLock()
	defer s.rxLock.RUnlock()
	slog.Info("read sbus frame", "frame", s.rxFrame)
	return s.rxFrame
}

func (s *SBus) SetWriteFrame(frame Frame) {
	s.txLock.Lock()
	defer s.txLock.Unlock()
	s.txFrame = frame
}

func (s *SBus) ListPorts() error {
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
