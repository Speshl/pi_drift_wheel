package crsf

//Run below command to generate model_enum file
//go-enum -f ./crsf/model.go

/*
ENUM(
	FlightController = 0xC8, //Most should have this address
	RadioTransmitter = 0xEA,
	Receiver = 0xEC,
	Transmitter = 0xEE //channels should have this address
)
*/
type AddressType byte

/*
ENUM(
GPS = 0x02
Vario = 0x07
BatterySensor = 0x08
Barometer = 0x09
LinkStats = 0x14
Channels = 0x16
ChannelSubSet = 0x17
LinkRx = 0x1C
LinkTx = 0x1D
Attitude = 0x1E
FlightMode = 0x21
)
*/
type FrameType byte
