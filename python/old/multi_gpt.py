import asyncio
from evdev import InputDevice, list_devices

async def read_events(device, collected_events):
    async for event in device.async_read_loop():
        # Process the event immediately
        collected_events[device].append(event)

async def main():
    devices = [InputDevice(fn) for fn in list_devices()]

    collected_events = {device.fd: [] for device in devices}

    for device in devices:
        print(device.path, device.name, device.phys)

    # Start reading events for each device asynchronously
    read_tasks = [asyncio.create_task(read_events(device, collected_events)) for device in devices]

    try:
        # Run the event reading tasks indefinitely
        while True:
            await asyncio.sleep(.01)  # Adjust the interval as needed
            # Access the collected events for each device and print
            for device, events in collected_events.items():
                if events:
                    print(f"Events for {device.name}: {events}")
                    collected_events[device] = []  # Clear processed events

    except asyncio.CancelledError:
        for task in read_tasks:
            task.cancel()

if __name__ == "__main__":
    asyncio.run(main())