import pygame
import time

pygame.init()




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


def main():
    
    joysticks = {}
    while handleEvents(joysticks):
        for joystick in joysticks.values():
            name = joystick.get_name()
            if name == "LLC Arduino Micro":
                value = joystick.get_axis(0)
                print(f"Value {value}")
        time.sleep(1)

if __name__ == "__main__":
    main()
    # If you forget this line, the program will 'hang'
    # on exit if running from IDLE.
    pygame.quit()