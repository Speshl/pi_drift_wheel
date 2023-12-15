import evdev
from evdev import ecodes, InputDevice

for device in evdev.list_devices():
    evtdev = InputDevice(device)
    val = 65535 # val \in [0,65535]
    evtdev.write(ecodes.EV_FF, ecodes.FF_AUTOCENTER, val)
    print("auto centering enabled")