package crsf

//Followed specification as defined on the wiki here: https://github.com/crsf-wg/crsf/wiki
import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"sync"

	"github.com/albenik/go-serial/v2"
	"golang.org/x/sync/errgroup"
)

type CRSF struct {
	path      string
	opts      CRSFOptions
	receiving bool
	//transmitting bool

	dataLock sync.RWMutex
	data     CRSFData
}

type CRSFOptions struct {
	BaudRate int
}

func NewCRSF(path string, opts *CRSFOptions) *CRSF {
	if opts == nil {
		opts = &CRSFOptions{
			BaudRate: 400000,
		}
	}

	return &CRSF{
		path: path,
		opts: *opts,
		data: NewCRSFData(),
	}
}

func (c *CRSF) String() string {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data.String()
}

func (c *CRSF) Start(ctx context.Context) error {
	port, err := serial.Open(c.path,
		serial.WithBaudrate(c.opts.BaudRate), //Looks like this can be multiple baudrates
		serial.WithDataBits(8),
		serial.WithParity(serial.NoParity),
		serial.WithStopBits(serial.OneStopBit),
		serial.WithReadTimeout(1000),
	)
	if err != nil {
		return err
	}

	crsfGroup, ctx := errgroup.WithContext(ctx)

	readChan := make(chan byte, 4096)

	crsfGroup.Go(func() error {
		defer close(readChan)
		return c.startReader(ctx, port, readChan)
	})

	crsfGroup.Go(func() error {
		return c.startReadParser(ctx, readChan)
	})

	// sbusGroup.Go(func() error {
	// 	return c.startWriter(ctx, port)
	// })

	return crsfGroup.Wait()
}

func (c *CRSF) startReader(ctx context.Context, port *serial.Port, readChan chan byte) error {
	c.receiving = true
	defer func() {
		c.receiving = false
	}()

	slog.Info("start reading from crsf", "path", c.path)
	buff := make([]byte, 128)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		n, err := port.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		slog.Info("read bytes", "num", n, "bytes", PrintBytes(buff[:n]))
		for i := range buff[:n] {
			readChan <- buff[i]
		}
	}
}

func (c *CRSF) startReadParser(ctx context.Context, readChan chan byte) error {
	for {
		addressByte, err := c.getByte(ctx, readChan)
		if err != nil {
			return err
		}
		//[sync] [len] [type] [payload] [crc8]
		if AddressType(addressByte).IsValid() {
			//slog.Info("found address type", "type", AddressType(addressByte).String())
			//next byte should be the length of the payload
			lengthByte, err := c.getByte(ctx, readChan)
			if err != nil {
				return err
			}

			if lengthByte == 0 {
				slog.Warn("payload has no length")
				continue
			}

			if lengthByte > 62 {
				slog.Warn("payload length to high")
				continue
			}

			//length should be the type + payload + CRC
			fullPayload, err := c.getBytes(ctx, readChan, int(lengthByte))
			if err != nil {
				return err
			}

			//first byte of the full payload should be the frame type
			//slog.Info("update looking for frame", "length", int(lengthByte), "frame", fullPayload[0], "type", FrameType(fullPayload[0]))
			switch FrameType(fullPayload[0]) {
			case FrameTypeChannels:
				err = c.updateChannels(fullPayload)
			//telemetry
			case FrameTypeGPS:
				err = c.updateGps(fullPayload)
			case FrameTypeVario:
				err = c.updateVario(fullPayload)
			case FrameTypeBatterySensor:
				err = c.updateBatterySensor(fullPayload)
			case FrameTypeBarometer:
				err = c.updateBarometer(fullPayload)
			case FrameTypeLinkStats:
				err = c.updateLinkStats(fullPayload)
			case FrameTypeLinkRx:
				err = c.updateLinkRx(fullPayload)
			case FrameTypeLinkTx:
				err = c.updateLinkTx(fullPayload)
			case FrameTypeAttitude:
				err = c.updateAttitude(fullPayload)
			case FrameTypeFlightMode:
				err = c.updateFlightMode(fullPayload)
			//unused
			// case FrameTypeOpenTxSync:
			// 	err = c.updateOpenTxSync(fullPayload)
			// case FrameTypeDevicePing:
			// 	err = c.updateDevicePing(fullPayload)
			// case FrameTypeDeviceInfo:
			// 	err = c.updateDeviceInfo(fullPayload)
			// case FrameTypeRequestSettings:
			// 	err = c.updateRequestSettings(fullPayload)
			// case FrameTypeParameterEntry:
			// 	err = c.updateParameterEntry(fullPayload)
			// case FrameTypeParameterRead:
			// 	err = c.updateParameterRead(fullPayload)
			// case FrameTypeParameterWrite:
			// 	err = c.updateParameterWrite(fullPayload)
			// case FrameTypeCommand:
			// 	err = c.updateCommand(fullPayload)
			// case FrameTypeRadioId:
			// 	err = c.updateRadioId(fullPayload)
			// case FrameTypeMspRequest:
			// 	err = c.updateMspRequest(fullPayload)
			// case FrameTypeMspResponse:
			// 	err = c.updateMspResponse(fullPayload)
			// case FrameTypeMspWrite:
			// 	err = c.updateMspWrite(fullPayload)
			// case FrameTypeDisplayCommand:
			// 	err = c.updateDisplayCommand(fullPayload)
			default:
				//slog.Warn("unsupported frame type", "type", fullPayload[0], "length", len(fullPayload))
			}
			if err != nil {
				return fmt.Errorf("failed parsing frame: %w", err)
			}
		} else {
			//slog.Warn("unsupported address", "byte", addressByte)
		}
	}
}

func (c *CRSF) getBytes(ctx context.Context, readChan chan byte, n int) ([]byte, error) {
	returnBytes := make([]byte, 0, n)

	for len(returnBytes) < n {
		select {
		case <-ctx.Done():
			return returnBytes, ctx.Err()
		case readByte, ok := <-readChan:
			if !ok {
				return returnBytes, fmt.Errorf("reader stopped")
			}
			returnBytes = append(returnBytes, readByte)
		}
	}
	return returnBytes, nil
}

func (c *CRSF) getByte(ctx context.Context, readChan chan byte) (byte, error) {
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case readByte, ok := <-readChan:
			if !ok {
				return 0, fmt.Errorf("reader stopped")
			}
			return readByte, nil
		}
	}
}

func PrintBytes(data []byte) string {
	returnString := ""
	for i := range data {
		returnString = fmt.Sprintf("%s,0x%s", returnString, strconv.FormatInt(int64(data[i]), 16))
	}
	return returnString
}
