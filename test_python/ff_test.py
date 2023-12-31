from evdev import ecodes, InputDevice, ff, list_devices
import time
import sys

# Find first EV_FF capable event device (that we have permissions to use).
dev = None
for name in list_devices():
    dev = InputDevice(name)
    print(f"Name: {dev.name} Device {dev.path}") #Capabilities: {dev.capabilities(True,True)}
    if ecodes.EV_FF in dev.capabilities():
        break

force = -1
envelope = ff.Envelope(0, 0, 0, 0)  # Attack time, Attack level, Fade time, Fade level
constant = ff.Constant(int(force * (65535 / 2)), envelope)

effect = ff.Effect(
  ecodes.FF_CONSTANT, -1, 20000, #16384
  ff.Trigger(0, 0),
  ff.Replay(0, 0),
  ff.EffectType(ff_constant_effect=constant)
)

effect_id = dev.upload_effect(effect)

print(f"effect id {effect_id}")

repeat_count = 1
dev.write(ecodes.EV_FF, effect_id, repeat_count)

time.sleep(2)

force = 1
envelope = ff.Envelope(0, 0, 0, 0)  # Attack time, Attack level, Fade time, Fade level
constant = ff.Constant(int(force * (65535 / 2)), envelope)

effect = ff.Effect(
  ecodes.FF_CONSTANT, 0, 20000, #16384
  ff.Trigger(0, 0),
  ff.Replay(0, 0),
  ff.EffectType(ff_constant_effect=constant)
)

effect_id = dev.upload_effect(effect)

print(f"effect id {effect_id} with effect size {sys.getsizeof(effect)}")

repeat_count = 1
#dev.write(ecodes.EV_FF, effect_id, repeat_count)
time.sleep(2)