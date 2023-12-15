import asyncio, evdev
import time

wheel = None
handBrake = None


devices = evdev.list_devices()
for device in devices:
    tmpDevice = evdev.InputDevice(device)
    print(tmpDevice)
    if tmpDevice.name == "G27 Racing Wheel":
        wheel = tmpDevice
        print("found wheel")
    elif tmpDevice.name == "Arduino LLC Arduino Micro":
        handBrake = tmpDevice
        print("found handbrake")

async def print_events(device):
    async for event in device.async_read_loop():
        print(device.path, evdev.categorize(event), sep=': ')

for device in wheel, handBrake:
    asyncio.ensure_future(print_events(device))

loop = asyncio.get_event_loop()
loop.run_forever()

while True:
    print("running")
    time.sleep(1)