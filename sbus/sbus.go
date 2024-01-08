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
)

const (
	RxTypeControl   = "control"
	RxTypeTelemetry = "telemetry"
)

var (
	ErrNoPath   = fmt.Errorf("no path provided for sbus")
	ErrNoAction = fmt.Errorf("no rx or tx action")
)

type SBusCfgOpts struct {
	Type string //TODO goenum
}

type SBus struct {
	path string

	read      bool
	receiving bool
	rxLock    sync.RWMutex
	rxFrame   Frame

	write        bool
	transmitting bool
	txLock       sync.RWMutex
	txFrame      Frame

	opts SBusCfgOpts
}

func NewSBus(path string, read bool, write bool, opts *SBusCfgOpts) (*SBus, error) {
	if path == "" {
		return nil, ErrNoPath
	}
	if !read && !write {
		return nil, ErrNoAction
	}

	if opts == nil {
		opts = &SBusCfgOpts{
			Type: RxTypeTelemetry,
		}
	}

	return &SBus{
		path:    path,
		read:    read,
		rxFrame: NewFrame(),
		write:   write,
		txFrame: NewFrame(),
		opts:    *opts,
	}, nil
}

func (s *SBus) Start(ctx context.Context) error {
	port, err := serial.Open(s.path,
		serial.WithBaudrate(100000),
		serial.WithDataBits(8),
		serial.WithParity(serial.EvenParity),
		serial.WithStopBits(serial.TwoStopBits),
		serial.WithReadTimeout(1000),
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

func (s *SBus) IsReceiving() bool {
	return s.receiving
}

func (s *SBus) startReader(ctx context.Context, port *serial.Port) error {
	if !s.read {
		return nil
	}
	s.receiving = true
	defer func() {
		s.receiving = false
	}()

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
							slog.Error("frame should have parsed but failed", "error", err)
						} else {
							s.rxLock.Lock()
							s.rxFrame = frame //set the latest frame
							s.rxLock.Unlock()
						}
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

func (s *SBus) IsTransmitting() bool {
	return s.transmitting
}

func (s *SBus) startWriter(ctx context.Context, port *serial.Port) error {
	if !s.write {
		return nil
	}
	s.transmitting = true
	defer func() {
		s.transmitting = false
	}()

	slog.Info("start writing to sbus", "path", s.Path)
	ticker := time.NewTicker(6 * time.Millisecond) //TODO sync with config
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
				slog.Warn("sbus write incorrect length")
			}

		}
	}
}

func (s *SBus) Path() string {
	return s.path
}

func (s *SBus) Type() string {
	return s.opts.Type
}

func (s *SBus) GetReadFrame() Frame {
	s.rxLock.RLock()
	defer s.rxLock.RUnlock()
	return s.rxFrame
}

func (s *SBus) SetWriteFrame(frame Frame) {
	s.txLock.Lock()
	defer s.txLock.Unlock()
	s.txFrame = frame
}
