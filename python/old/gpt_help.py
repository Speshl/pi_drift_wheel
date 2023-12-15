import asyncio
from evdev import InputDevice, list_devices

async def read_events(device, event_queue):
    async for event in device.async_read_loop():
        await event_queue.put(event)

async def main():
    devices = [InputDevice(fn) for fn in list_devices()]

    for device in devices:
        print(device.path, device.name, device.phys)

    # Replace '/dev/input/eventX' with the path of your input device
    device_path = '/dev/input/event2'  # Change this to your device's path
    device = InputDevice(device_path)

    event_queue = asyncio.Queue()

    while True:
        read_task = asyncio.ensure_future(read_events(device, event_queue))
        await asyncio.sleep(.01)
        read_task.cancel()

        # Accessing events collected in the queue after the async loop
        events = []
        while not event_queue.empty():
            event = await event_queue.get()
            events.append(event)

        # Process events or do something with the collected data
        for event in events:
            print("got event")
            #print(event)

if __name__ == "__main__":
    asyncio.run(main())