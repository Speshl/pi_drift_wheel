package crsf

//Run below command to regenerate model_enum file
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

//unused addresses
/*
Broadcast = 0x00
Usb = 0x10,
Bluetooth = 0x12
TbsCore = 0x80,
Reserved1 = 0x8A,
CurrentSensor = 0xC0,
Gps = 0xC2,
BlackBox = 0xC4,

Reserved2 = 0xCA,
Racetag = 0xCC,

Receiver = 0xEC,

ElrsLua = 0xEF
*/

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

//unsed frametypes
/*
OpenTxSync = 0x10
DevicePing = 0x28
DeviceInfo = 0x29
RequestSettings = 0x2A
ParameterEntry = 0x2B
ParameterRead = 0x2C
ParameterWrite = 0x2D
Command = 0x32
RadioId = 0x3A
MspRequest = 0x7A
MspResponse = 0x7B
MspWrite = 0x7C
DisplayCommand = 0x7D
*/
//Unused FrameTypes
