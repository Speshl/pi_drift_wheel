#https://github.com/Vinz1911/PyPad2/blob/main/src/pypad2/gamepad.py
import random
import time
import pigpio
import evdev
from evdev import InputDevice, ecodes, ff
from threading import Thread
from enum import Enum

class X:
    GAP=300
    WAVES=3

    def __init__(self, pi, gpio, channels=8, frame_ms=27):
      self.pi = pi
      self.gpio = gpio

      if frame_ms < 5:
         frame_ms = 5
         channels = 2
      elif frame_ms > 100:
         frame_ms = 100

      self.frame_ms = frame_ms

      self._frame_us = int(frame_ms * 1000)
      self._frame_secs = frame_ms / 1000.0

      if channels < 1:
         channels = 1
      elif channels > (frame_ms // 2):
         channels = int(frame_ms // 2)

      self.channels = channels

      self._widths = [1500] * channels # set each channel to middle pulse width

      self._wid = [None]*self.WAVES
      self._next_wid = 0

      pi.write(gpio, pigpio.LOW)

      self._update_time = time.time()
      self._thread = Thread(target=self._update, daemon=True)
      self._thread.start()

    def _update(self):
      while True:
        #print("updating ppm")
        wf =[]
        micros = 0
        index = 0
        for i in self._widths:
            #print(f"Channel {index} value {i}")
            wf.append(pigpio.pulse(1<<self.gpio, 0, self.GAP))
            wf.append(pigpio.pulse(0, 1<<self.gpio, i-self.GAP))
            micros += i
            index += 1
        # off for the remaining frame period
        wf.append(pigpio.pulse(1<<self.gpio, 0, self.GAP))
        micros += self.GAP
        wf.append(pigpio.pulse(0, 1<<self.gpio, self._frame_us-micros))

        self.pi.wave_add_generic(wf)
        wid = self.pi.wave_create()
        self.pi.wave_send_using_mode(wid, pigpio.WAVE_MODE_REPEAT_SYNC)
        self._wid[self._next_wid] = wid

        self._next_wid += 1
        if self._next_wid >= self.WAVES:
            self._next_wid = 0

        
        remaining = self._update_time + self._frame_secs - time.time()
        if remaining > 0:
            time.sleep(remaining)
        self._update_time = time.time()

        wid = self._wid[self._next_wid]
        if wid is not None:
            self.pi.wave_delete(wid)
            self._wid[self._next_wid] = None

    def update_channel(self, channel, width):
      self._widths[channel] = width
      #self._update()

    def update_channels(self, widths):
      self._widths[0:len(widths)] = widths[0:self.channels]
      #self._update()

    def cancel(self):
      self.pi.wave_tx_stop()
      for i in self._wid:
         if i is not None:
            self.pi.wave_delete(i)


class DIYHandBrakeKeymap(Enum):
    AXE_BRAKE = 0

class G27Keymap(Enum):
    BTN_TOP_LEFT_WHEEL = 292
    BTN_MID_LEFT_WHEEL = 708
    BTN_BOTTOM_LEFT_WHEEL = 709

    BTN_TOP_RIGHT_WHEEL = 292
    BTN_MID_RIGHT_WHEEL = 706
    BTN_BOTTOM_RIGHT_WHEEL = 707

    BTN_UPSHIFT = 293
    BTN_DOWNSHIFT = 292

    BTN_FIRST_GEAR = 300
    BTN_SECOND_GEAR = 301
    BTN_THIRD_GEAR = 302
    BTN_FOURTH_GEAR = 303
    BTN_FIFTH_GEAR = 704
    BTN_SIXTH_GEAR = 705
    BTN_REVERSE_GEAR = 710

    BTN_RED1 = 299
    BTN_RED2 = 296
    BTN_RED3 = 297
    BTN_RED4 = 298

    BTN_Y = 291
    BTN_B = 290
    BTN_A = 288
    BTN_X = 289

    HAT_UP_DOWN = 17
    HAT_LEFT_RIGHT = 16

    AXE_WHEEL = 0
    AXE_GAS = 2
    AXE_BRAKE = 5
    AXE_CLUTH = 3

class Gamepad:
    def __init__(self, path):
        self.__input: InputDevice = None
        self.__path: str = path
        self.__buttons: dict = {}
        self.__thread = Thread(target=self.__read_input, daemon=True)
        self.__open()

    def pressed(self) -> dict:
        return self.__buttons

    def name(self) -> str:
        if not self.__input: return str()
        return self.__input.name
    
    def __determineKeyMap(self) -> Enum:
        if self.__input.name == "G27 Racing Wheel":
            return G27Keymap
        elif self.__input.name == "Arduino LLC Arduino Micro":
            return DIYHandBrakeKeymap

    def __open(self):
        self.__input = InputDevice(self.__path)
        self.__keymap = self.__determineKeyMap()
        for entry in self.__keymap: self.__buttons[entry] = 999999
        self.__thread.start()

    def __read_input(self):
        input_events = self.__input.read_loop()
        for event in input_events:
            if event.type == ecodes.EV_KEY or event.type == ecodes.EV_ABS:
                try:
                    key = self.__keymap(event.code)
                    self.__buttons[key] = event.value
                except:
                    #print(f"got unused key {event.code} with value {event.value}")
                    continue
                    

class State:
    def __init__(self, channels):
        self.channels = channels

def mapGamepadToState(name, buttons) -> State:
    if name == "G27 Racing Wheel":
        return mapG27ToState(buttons)
    elif name == "Arduino LLC Arduino Micro":
        return mapDIYToState(buttons)

def mapG27ToState(buttons) -> State:
    steerValue = buttons[G27Keymap.AXE_WHEEL] # 0-16383
    if steerValue == 999999:
        steerValue = 8190

    gasValue = buttons[G27Keymap.AXE_GAS] # 0-255 inverted
    if gasValue == 999999:
        gasValue = 255
    gasValue = 255 - gasValue
    
    brakeValue = buttons[G27Keymap.AXE_BRAKE] # 0-255 inverted
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
    brakeValue = buttons[DIYHandBrakeKeymap.AXE_BRAKE] # -127 - 127
    if brakeValue == 999999:
        brakeValue = -127

    brakeValue = mapToRangeWithDeadzone(brakeValue, -127, 127, 1000, 1500, 5)
    flippedBrakeValue = 1500 - brakeValue + 1000 #flip range for bottom half

    return State([1500, flippedBrakeValue, 1500,1500,1500,1500,1500,1500])

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
    mappedValue = (maxReturn-minReturn)*(value-minInput)/(maxInput-minInput) + minReturn

    if midPoint + deadzone > mappedValue and mappedValue > midPoint:
        return midPoint
    elif midPoint - deadzone < mappedValue and mappedValue < midPoint:
        return midPoint
    elif mappedValue > maxReturn:
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

    ppm = X(pi, 8, frame_ms=20)

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
        gamepads.append(Gamepad(path))

    startTime = time.time()
    updates = 0
    while True:
        states = []
        for gamepad in gamepads:
            states.append(mapGamepadToState(gamepad.name(), gamepad.pressed()))

        if len(states) == 0:
            print("no gamepad states found")
            exit(0)

        finalState = mergeStates(states)
        
        finalValues = []
        for value in finalState.channels:
            finalValues.append(int(value))

        ppm.update_channels(finalValues)

        updates+=1
        difTime = time.time()-startTime
        print(f"Merged: {finalState.channels} Updates: {updates} Elapsed: {difTime}")
        time.sleep(0.016)


if __name__ == "__main__":
    main()