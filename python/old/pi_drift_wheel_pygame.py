import pygame
import time
import pigpio
import evdev
from evdev import ecodes, InputDevice


try:
    import pigpio
except ImportError as e:
    pigpio = None

pygame.init()

# Raspberry Pi GPIO pin where to output the PPM signal.
# Pin map: http://wiki.mchobby.be/images/3/31/RASP-PIZERO-Correspondance-GPIO.jpg
# (Connect this pin to the RC transmitter trainer port.)
PPM_OUTPUT_PIN = 18

class inputGroup:
    def __init__(self, joystickId, buttonId, minValue, maxValue, defaultValue, invert):
        self.joystickId = joystickId
        self.buttonId = buttonId
        self.minValue = minValue
        self.maxValue = maxValue
        self.defaultValue = defaultValue
        self.invert = invert

class controls:
    def __init__(self,channelId, type, fullRange, highRange, lowRange):
        self.channelId = channelId
        self.type = type
        self.fullRange = fullRange
        self.highRange = highRange
        self.lowRange = lowRange

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

      self._widths = [1000] * channels # set each channel to minimum pulse width

      self._wid = [None]*self.WAVES
      self._next_wid = 0

      pi.write(gpio, pigpio.LOW)

      self._update_time = time.time()

   def _update(self):
      wf =[]
      micros = 0
      for i in self._widths:
         wf.append(pigpio.pulse(1<<self.gpio, 0, self.GAP))
         wf.append(pigpio.pulse(0, 1<<self.gpio, i-self.GAP))
         micros += i
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
      self._update()

   def update_channels(self, widths):
      self._widths[0:len(widths)] = widths[0:self.channels]
      self._update()

   def cancel(self):
      self.pi.wave_tx_stop()
      for i in self._wid:
         if i is not None:
            self.pi.wave_delete(i)

def getG27ControlMap(joystickId, secondaryId):
    controlsList = []

    wheelInput = []
    wheelInput.append(inputGroup(joystickId, 0, -1, 1, 0, False))
    controlsList.append(controls(0, "axis_wheel", wheelInput, None, None))

    escHighInput = []
    escHighInput.append(inputGroup(joystickId, 2, 0, 1, 0, True))

    escLowInput = []
    escLowInput.append(inputGroup(joystickId, 3, 0, 1, 0, True))
    
    if secondaryId != None: #Secondary joystick as a handbrake
         escLowInput.append(inputGroup(secondaryId, 0, -1, 1, -1, False))
   

    controlsList.append(controls(1, "axis_split_pedal", None, escHighInput, escLowInput))
    return controlsList

def handleEvents(joysticks):
    # Event processing step.
    # Possible joystick events: JOYAXISMOTION, JOYBALLMOTION, JOYBUTTONDOWN,
    # JOYBUTTONUP, JOYHATMOTION, JOYDEVICEADDED, JOYDEVICEREMOVED
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            return False #Stop

        if event.type == pygame.JOYBUTTONDOWN:
            print("Joystick button pressed.")

        if event.type == pygame.JOYBUTTONUP:
            print("Joystick button released.")

        # Handle hotplugging
        if event.type == pygame.JOYDEVICEADDED:
            # This event will be generated when the program starts for every
            # joystick, filling up the list without needing to create them manually.
            joy = pygame.joystick.Joystick(event.device_index)
            joysticks[joy.get_instance_id()] = joy
            print(f"Joystick {joy.get_instance_id()} connencted")

        if event.type == pygame.JOYDEVICEREMOVED:
            del joysticks[event.instance_id]
            print(f"Joystick {event.instance_id} disconnected")
    return True

def mapToRange(value,minInput,maxInput,minReturn,maxReturn):
    mappedValue = (maxReturn-minReturn)*(value-minInput)/(maxInput-minInput) + minReturn

    if mappedValue > maxReturn:
        return maxReturn
    elif mappedValue < minReturn:
        return minReturn
    else:
        return mappedValue
    
def getPriorityInput(joysticks, inputs):
    greatestMovement = 0
    priorityInput = inputs[0]
    for inpt in inputs:
        joystick = joysticks[inpt.joystickId]
        value = joystick.get_axis(inpt.buttonId)
        movement = getDistanceFromDefault(value, inpt)
        if movement > greatestMovement:
            greatestMovement = movement
            priorityInput = inpt
            #print(f"found new priority {joystick.get_name()} with {movement} movement")

    # priorityJoystick = joysticks[priorityInput.joystickId]
    # finalName = priorityJoystick.get_name()
    # finalValue = priorityJoystick.get_axis(inpt.buttonId)
    # print(f"Priority {finalName} value {round(finalValue,2)}")       
    return priorityInput

def getDistanceFromDefault(value, inpt):
    if inpt.minValue == -1 and inpt.maxValue == 1 and inpt.defaultValue == -1: #Shift a full axis that starts at the bottom to top half
        value = mapToRange(value, -1, 1, 0, 1)
        if inpt.invert:
            value = invertInputValue(value,inpt)
    elif inpt.minValue == 0 and inpt.maxValue == 1 and inpt.defaultValue == 0:
        if inpt.invert:
            value = invertInputValue(value,inpt)
    elif inpt.minValue == -1 and inpt.maxValue == 1 and inpt.defaultValue == 0:
        value = abs(value)
    return value

def invertInputValue(value, inpt):
    if inpt.minValue == inpt.defaultValue:
            value = inpt.maxValue-value
    elif (inpt.minValue - inpt.maxValue) == inpt.defaultValue:
            value = value * -1
    return value

def mapInputToRange(joysticks, inpt, minReturn, maxReturn):
    minInput = inpt.minValue
    maxInput = inpt.maxValue
    joystick = joysticks[inpt.joystickId]
    value = joystick.get_axis(inpt.buttonId)
    if inpt.invert:
        value = invertInputValue(value,inpt)

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

    ppm = X(pi, 4, frame_ms=20)

    for device in evdev.list_devices():
        evtdev = InputDevice(device)
        val = 65535 # val \in [0,65535]
        evtdev.write(ecodes.EV_FF, ecodes.FF_AUTOCENTER, val)
        print("auto centering enabled")

    joysticks = {}
    controlsList = []
    handleEvents(joysticks)

    for joystick in joysticks.values():
        name = joystick.get_name()
        print(f"Joystick: {name}")
        if name == "G27 Racing Wheel": #change to guid?

            #check for handbrake
            secondaryId = None
            for joystick2 in joysticks.values():
                name2 = joystick2.get_name()
                if name2 == "LLC Arduino Micro": #change to guid?
                    secondaryId = joystick2.get_instance_id()
            controlsList = getG27ControlMap(joystick.get_instance_id(), secondaryId)

    print("Listening for events")
    while handleEvents(joysticks):
        for control in controlsList:
            final = 1500
            if control.type == "axis_wheel":
                inpt = getPriorityInput(joysticks, control.fullRange)
                final = mapInputToRange(joysticks, inpt, 1000,2000)
                print(f"Channel {control.channelId} Type {control.type} Final {final}")

            if control.type == "axis_split_pedal":
                highInput = getPriorityInput(joysticks,control.highRange)
                lowInput = getPriorityInput(joysticks, control.lowRange)

                adjHighValue = mapInputToRange(joysticks, highInput, 1500, 2000)
                adjLowValue = mapInputToRange(joysticks, lowInput , 1000, 1500)
                flippedLowValue = 1500 - adjLowValue + 1000 #flip range for bottom half

                deadzone = 5
                if 1500+deadzone > adjHighValue:
                    adjHighValue = 1500
                if 1500-deadzone < flippedLowValue:
                    flippedLowValue = 1500
                
                highValueDiff = adjHighValue - 1500
                lowValueDiff = adjLowValue - 1000

                finalEscValue = flippedLowValue
                if highValueDiff > lowValueDiff:
                    finalEscValue = adjHighValue

                print(f"Channel {control.channelId} Type {control.type} adjLow {flippedLowValue} adjHigh {adjHighValue} final {finalEscValue}")

            ppm.update_channel(control.channelId, int(final))
        print("sleeping")
        time.sleep(2)


if __name__ == "__main__":
    main()
    # If you forget this line, the program will 'hang'
    # on exit if running from IDLE.
    pygame.quit()