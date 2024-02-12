#!/usr/bin/env python3
import serial
import time
import argparse
from enum import IntEnum

class PacketsTypes(IntEnum):
    GPS = 0x02
    VARIO = 0x07
    BATTERY_SENSOR = 0x08
    HEARTBEAT = 0x0B
    VIDEO_TRANSMITTER = 0x0F
    LINK_STATISTICS = 0x14
    RC_CHANNELS_PACKED = 0x16
    ATTITUDE = 0x1E
    FLIGHT_MODE = 0x21
    DEVICE_INFO = 0x29
    CONFIG_READ = 0x2C
    CONFIG_WRITE = 0x2D
    RADIO_ID = 0x3A

def crc8_dvb_s2(crc, a) -> int:
  crc = crc ^ a
  for ii in range(8):
    if crc & 0x80:
      crc = (crc << 1) ^ 0xD5
    else:
      crc = crc << 1
  return crc & 0xFF

def crc8_data(data) -> int:
    crc = 0
    for a in data:
        crc = crc8_dvb_s2(crc, a)
    return crc

def crsf_validate_frame(frame) -> bool:
    return crc8_data(frame[2:-1]) == frame[-1]

def handleCrsfPacket(ptype, data):
    if ptype == PacketsTypes.RADIO_ID and data[5] == 0x10:
        print(f"OTX sync")
        pass
    elif ptype == PacketsTypes.LINK_STATISTICS:
        rssi1 = int.from_bytes(data[3:4], byteorder='big', signed=True)
        rssi2 = int.from_bytes(data[4:5], byteorder='big', signed=True)
        lq = data[5]
        mode = data[8]
        print(f"RSSI={rssi1}/{rssi2}dBm LQ={lq:03} mode={mode}")
    elif ptype == PacketsTypes.ATTITUDE:
        pitch = data[3] << 8 | data[4]
        roll = data[5] << 8 | data[6]
        yaw = data[7] << 8 | data[8]
        print(f"Attitude: Pitch={pitch} Roll={roll} Yaw={yaw}")
    elif ptype == PacketsTypes.FLIGHT_MODE:
        packet = ''.join(map(chr, data[3:-2]))
        print(f"Flight Mode: {packet}")
    elif ptype == PacketsTypes.BATTERY_SENSOR:
        vbat = data[3] << 8 | data[4]
        curr = data[5] << 8 | data[6]
        pct = data[7]
        mah = data[8] << 16 | data[9] << 7 | data[10]
        print(f"Battery: {vbat/10.0}V {curr}A {pct}% {mah}mAh")
    elif ptype == PacketsTypes.DEVICE_INFO:
        packet = ' '.join(map(hex, data))
        print(f"Device Info: {packet}")
    elif data[2] == PacketsTypes.GPS:
        lat = int.from_bytes(data[3:7], byteorder='big', signed=True) / 1e7
        lon = int.from_bytes(data[7:11], byteorder='big', signed=True) / 1e7
        gspd = (data[11] << 8 | data[12]) / 36.0
        hdg =  (data[13] << 8 | data[14]) / 100.0
        alt = (data[15] << 8 | data[16]) - 1000
        sats = data[17]
        print(f"GPS: Pos={lat} {lon} GSpd={gspd:0.1f}m/s Hdg={hdg:0.1f} Alt={alt}m Sats={sats}")
    elif ptype == PacketsTypes.VARIO:
        vspd = int.from_bytes(data[3:5], byteorder='big', signed=True) / 10.0
        print(f"VSpd: {vspd:0.1f}m/s")
    elif ptype == PacketsTypes.RC_CHANNELS_PACKED:
        print(f"Channels: (data)")
        pass
    else:
        packet = ' '.join(map(hex, data))
        print(f"Unknown 0x{ptype:02x}{PacketsTypes.LINK_STATISTICS:02x}: {packet}")

parser = argparse.ArgumentParser()
parser.add_argument('-P', '--port', default='/dev/ttyAMA0', required=False)  #ttyACM0 for USB, ttyAMA0for pins
parser.add_argument('-b', '--baud', default=420000, required=False) #921600 for CRSF
args = parser.parse_args()

with serial.Serial(args.port, args.baud, timeout=2) as ser:
    input = bytearray()
    while True:
        if ser.in_waiting > 0:
            input.extend(ser.read(ser.in_waiting))
            if len(input) > 2:
                expected_len = input[1] + 2
                if expected_len > 64 or expected_len < 4:
                    input = []
                elif len(input) >= expected_len:
                    single = input[:expected_len] # copy out this whole packet
                    input = input[expected_len:] # and remove it from the buffer

                    if not crsf_validate_frame(single): # single[-1] != crc:
                        packet = ' '.join(map(hex, single))
                        print(f"crc error: {packet}")
                    else:
                        print(f"packet: {packet}")
                        handleCrsfPacket(single[2], single)
        else:
            time.sleep(0.020)