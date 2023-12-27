package crsf

import (
	"context"
	"log"
	"log/slog"

	"github.com/albenik/go-serial/v2"
	"golang.org/x/sync/errgroup"
)

type CRSF struct {
	path string

	receiving bool
	//transmitting bool

	//dataLock sync.RWMutex
	data CRSFData
}

type CRSFData struct {
}

func NewCRSF(path string) *CRSF {
	return &CRSF{
		path: path,
		data: NewCRSFData(),
	}
}

func NewCRSFData() CRSFData {
	return CRSFData{}
}

func (c *CRSF) Start(ctx context.Context) error {
	port, err := serial.Open(c.path,
		serial.WithBaudrate(420000), //Looks like this can be multiple baudrates
		serial.WithDataBits(8),
		//serial.WithParity(serial.EvenParity),
		serial.WithStopBits(serial.OneStopBit),
		serial.WithReadTimeout(1000),
	)
	if err != nil {
		return err
	}

	sbusGroup, ctx := errgroup.WithContext(ctx)

	sbusGroup.Go(func() error {
		return c.startReader(ctx, port)
	})

	// sbusGroup.Go(func() error {
	// 	return c.startWriter(ctx, port)
	// })

	return sbusGroup.Wait()
}

func (c *CRSF) startReader(ctx context.Context, port *serial.Port) error {
	c.receiving = true
	defer func() {
		c.receiving = false
	}()

	slog.Info("start reading from crsf", "path", c.path)
	buff := make([]byte, 64)
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
			switch buff[i] {
			case 0xEE:
				slog.Info("msg to transmitter module")
			case 0xEA:
				slog.Info("msg to handset")
			case 0xC8:
				slog.Info("msg to flight controller")
			case 0xEC:
				slog.Info("msg to receiver")

			//Alt
			case 0x00:
				slog.Info("msg to broadcast")
			case 0x10:
				slog.Info("msg to usb")
			}
		}

		for i := range buff[:n] {
			switch buff[i] {
			case 0x02:
				slog.Info("gps")
			case 0x08:
				slog.Info("battery")
			case 0x14:
				slog.Info("link stats")
			case 0x10:
				slog.Info("opentx sync")
			case 0x3A:
				slog.Info("radio id")
			case 0x16:
				slog.Info("channels")
			case 0x1E:
				slog.Info("attitude")
			case 0x21:
				slog.Info("flight mode")
			case 0x28:
				slog.Info("device ping")
			case 0x29:
				slog.Info("device info")
			case 0x2B:
				slog.Info("parameter entry")
			case 0x2C:
				slog.Info("parameter read")
			case 0x2D:
				slog.Info("parameter write")
			case 0x32:
				slog.Info("command")

			//MSP
			case 0x7A:
				slog.Info("msp req")
			case 0x7B:
				slog.Info("msp resp")
			case 0x7C:
				slog.Info("msp write")
			}
		}
		//slog.Info("read", "num_read", n, "data", buff[:n])
	}
}
