#include <stdio.h>
#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include <unistd.h>  // Include for the write() function
#include <linux/input.h>

int upload_effect(uintptr_t fd,  int16_t level, bool effectExists){
    struct ff_effect effect = {};

    effect.type = FF_CONSTANT;
    if(!effectExists){
        effect.id = -1;           // Unique ID for the effect (use -1 for auto-assignment)
    }else{
        // ioctl(fd, EVIOCRMFF, 0);
        // effect.id = -1;
        effect.id = 0;
    }
   
    effect.direction = 20000;     // Direction of the effect (0 for omni-directional)
    effect.trigger.button = 0; // Button that triggers the effect (0 for no button)
    effect.trigger.interval = 0; // Interval between triggers (0 for continuous)
    effect.replay.length = 0;  // Duration of the effect in milliseconds
    effect.replay.delay = 0;     // Delay before replaying the effect (0 for no delay)

    // Parameters specific to the constant effect
    effect.u.constant.level = level; // Example: Constant force level (signed 16-bit)
    int error = ioctl(fd, EVIOCSFF, &effect);
    if(error != 0){
        return -2;
    }

    struct input_event event;
    struct timeval tval;
    gettimeofday(&tval, 0);
    event.input_event_usec = tval.tv_usec;
    event.input_event_sec = tval.tv_sec;
    event.type = EV_FF;
    event.code = effect.id;
    event.value = 1;

    if (write(fd, &event, sizeof(event)) != sizeof(event)) {
        return -3;
    }
    return effect.id;
}