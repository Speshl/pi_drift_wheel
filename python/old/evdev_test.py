from evdev import InputDevice
from select import select

gamepad = InputDevice('/dev/input/event2')
print(gamepad)
while True:
    r,w,x = select([gamepad], [], [])
    for event in gamepad.read():
        if event.type == 1: #button
            if event.code == 292:
                print("top left wheel")
            elif event.code == 708:
                print("mid left wheel")
            elif event.code == 709:
                print("bottom left wheel")
            elif event.code == 294:
                print("top right wheel")
            elif event.code == 706:
                print("mid left wheel")
            elif event.code == 707:
                print("bottom left wheel")

            elif event.code == 293:
                print("down shift")
            elif event.code == 292:
                print("up shift")

            elif event.code == 300:
                print("first gear")
            elif event.code == 301:
                print("second gear")
            elif event.code == 302:
                print("third gear")
            elif event.code == 303:
                print("fourth gear")
            elif event.code == 704:
                print("fifth gear")
            elif event.code == 705:
                print("sixth gear")
            elif event.code == 710:
                print("reverse gear")

            elif event.code == 299:
                print("red 1")
            elif event.code == 296:
                print("red 2")
            elif event.code == 297:
                print("red 3")
            elif event.code == 298:
                print("red 4")

            elif event.code == 291:
                print("Y")
            elif event.code == 290:
                print("B")
            elif event.code == 288:
                print("A")
            elif event.code == 289:
                print("X")
            
            elif event.code == 17:
                if event.val == -1:
                    print("UP")
                elif event.val == 1:
                    print("DOWN")
            elif event.code == 16:
                if event.val == -1:
                    print("LEFT")
                elif event.val == 1:
                    print("RIGHT")

        elif event.type == 3: #axis
            if event.code == 0:
                print(f"steer {event.value}")
            elif event.code == 2:
                print(f"gas {event.value}")
            elif event.code == 5:
                print(f"brake {event.value}")
            elif event.code == 3:
                print(f"clutch {event.value}")
            #else:
                #print(event)
        #else:
            #print(event)
