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

type SBus struct {
	Path string

	read    bool
	inLock  sync.RWMutex
	inFrame Frame

	write    bool
	outLock  sync.RWMutex
	outFrame Frame
}

func NewSBus(cfg config.SBusConfig) *SBus {
	return &SBus{
		Path:     cfg.SBusPath,
		read:     cfg.SBusIn,
		inFrame:  NewFrame(),
		write:    cfg.SBusOut,
		outFrame: NewFrame(),
	}
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
						s.inLock.Lock()
						s.inFrame = frame //set the latest frame
						s.inLock.Unlock()
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

	ticker := time.NewTicker(6 * time.Millisecond)
	var writeBytes []byte
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.outLock.RLock()
			writeBytes = s.outFrame.Marshal()
			s.outLock.RUnlock()
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
	s.inLock.RLock()
	defer s.inLock.RUnlock()
	return s.inFrame
}

func (s *SBus) SetWriteFrame(frame Frame) {
	s.outLock.Lock()
	defer s.outLock.Unlock()
	s.outFrame = frame
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
