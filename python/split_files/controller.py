from threading import Thread
from enum import Enum
import evdev
from evdev import InputDevice, ecodes, ff

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