import pigpio
import evdev
from evdev import InputDevice, ecodes, ff
import time

import controller
import ppm
import read_sbus

ppm_gpio_pin = 8
sbus_gpio_pin = 4


class State:
    def __init__(self, channels):
        self.channels = channels

def mapGamepadToState(name, buttons) -> State:
    if name == "G27 Racing Wheel":
        return mapG27ToState(buttons)
    elif name == "Arduino LLC Arduino Micro":
        return mapDIYToState(buttons)

def mapG27ToState(buttons) -> State:
    steerValue = buttons[controller.G27Keymap.AXE_WHEEL] # 0-16383
    if steerValue == 999999:
        steerValue = 8190

    gasValue = buttons[controller.G27Keymap.AXE_GAS] # 0-255 inverted
    if gasValue == 999999:
        gasValue = 255
    gasValue = 255 - gasValue
    
    brakeValue = buttons[controller.G27Keymap.AXE_BRAKE] # 0-255 inverted
    if brakeValue == 999999:
        brakeValue = 255
    brakeValue = 255 - brakeValue

    steerValue = mapToRangeWithDeadzone(steerValue, 0, 16383, 1000, 2000, 5)
    gasValue = mapToRangeWithDeadzone(gasValue, 0, 255, 1500, 2000, 5)
    brakeValue = mapToRangeWithDeadzone(brakeValue, 0, 255, 1000, 1500, 5)
    flippedBrakeValue = 1500 - brakeValue + 1000 #flip range for bottom half

    highValueDiff = gasValue - 1500
    lowValueDiff = brakeValue - 1000

    escValue = flippedBrakeValue
    if highValueDiff > lowValueDiff:
        escValue = gasValue
    
    return State([steerValue, escValue, 1500,1500,1500,1500,1500,1500])

def mapDIYToState(buttons) -> State:
    brakeValue = buttons[controller.DIYHandBrakeKeymap.AXE_BRAKE] # -127 - 127
    if brakeValue == 999999:
        brakeValue = -127

    brakeValue = mapToRangeWithDeadzone(brakeValue, -127, 127, 1000, 1500, 5)
    flippedBrakeValue = 1500 - brakeValue + 1000 #flip range for bottom half

    return State([1500, flippedBrakeValue, 1500,1500,1500,1500,1500,1500])

def mapSbusToState(data) -> State:
    stateValues = []
    for entry in data:
        stateValues.append(mapToRange(entry, 0, 1000, 1000, 2000))
    return State(stateValues)

def mergeStates(states) -> State:
    index = 0
    finalState = states[0]
    for state in states:
        index = 0
        for channel in state.channels:
            diff = abs(channel - 1500)
            compDiff = abs(finalState.channels[index] - 1500)
            if diff > compDiff:
                finalState.channels[index] = channel
            index += 1
    return finalState


def isSupportedGamepad(name) -> bool:
    if name == "G27 Racing Wheel":
        return True
    elif name == "Arduino LLC Arduino Micro":
        return True
    else:
        return False
    
def mapToRangeWithDeadzone(value,minInput,maxInput,minReturn,maxReturn,deadzone):
    midPoint = (maxReturn + minReturn) / 2
    mappedValue = mapToRange(value,minInput,maxInput,minReturn,maxReturn)

    if midPoint + deadzone > mappedValue and mappedValue > midPoint:
        return midPoint
    elif midPoint - deadzone < mappedValue and mappedValue < midPoint:
        return midPoint
    else:
        return mappedValue
    
def mapToRange(value,minInput,maxInput,minReturn,maxReturn):
    mappedValue = (maxReturn-minReturn)*(value-minInput)/(maxInput-minInput) + minReturn

    if mappedValue > maxReturn:
        return maxReturn
    elif mappedValue < minReturn:
        return minReturn
    else:
        return mappedValue

def main():
    pi = pigpio.pi()
    if not pi.connected:
      exit(0)

    pi.wave_tx_stop() # Start with a clean slate.

    ppmSender = ppm.X(pi, ppm_gpio_pin, frame_ms=20)

    sbusReader = read_sbus.read_sbus_from_GPIO.SbusReader(sbus_gpio_pin)
    sbusReader.begin_listen()

    devices = evdev.list_devices()
    paths = []
    for device in devices:
        gamepad = InputDevice(device)
        print(f"found {gamepad.name}")
        if isSupportedGamepad(gamepad.name):
            print(f"using {gamepad.name}")
            evtdev = InputDevice(device)
            val = 65535 # val \in [0,65535]
            evtdev.write(ecodes.EV_FF, ecodes.FF_AUTOCENTER, val) #Enable centering spring
            print("auto centering enabled")
            paths.append(device)

    gamepads = []
    for path in paths:
        gamepads.append(controller.Gamepad(path))

    startTime = time.time()
    updates = 0
    while True:
        states = []
        for gamepad in gamepads:
            states.append(mapGamepadToState(gamepad.name(), gamepad.pressed()))

        if sbusReader.is_connected():
            sbusData = sbusReader.translate_latest_packet()
            states.append(mapSbusToState(sbusData))

        if len(states) == 0:
            print("no gamepad states found")
            exit(0)

        finalState = mergeStates(states)
        
        finalValues = []
        for value in finalState.channels:
            finalValues.append(int(value))

        ppmSender.update_channels(finalValues)

        updates+=1
        difTime = time.time()-startTime
        print(f"Merged: {finalState.channels} Updates: {updates} Elapsed: {difTime}")
        time.sleep(0.016)


if __name__ == "__main__":
    main()