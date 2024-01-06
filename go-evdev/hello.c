#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <linux/input.h>

void Hello(){
    printf("Hello world\n");
}

int upload_effect(uintptr_t fd,  void *effect_data){
    struct ff_effect effect = {};
    //memmove(&effect, effect_data, sizeof(struct ff_effect));

    effect.type = FF_CONSTANT;
    effect.id = -1;           // Unique ID for the effect (use -1 for auto-assignment)
    effect.direction = 0;     // Direction of the effect (0 for omni-directional)
    effect.trigger.button = 0; // Button that triggers the effect (0 for no button)
    effect.trigger.interval = 0; // Interval between triggers (0 for continuous)
    effect.replay.length = 100;  // Duration of the effect in milliseconds
    effect.replay.delay = 0;     // Delay before replaying the effect (0 for no delay)

    // Parameters specific to the constant effect
    effect.u.constant.level = 0x8000; // Example: Constant force level (signed 16-bit)
    int error = ioctl(fd, EVIOCSFF, &effect);
    if(error != 0){
        return -1;
    }

    struct input_event play_event = {
        .type = EV_FF,
        .code = effect.id,
        .value = 1, // 1 for start playing, 0 for stop
    };

    error = write(fd, &play_event, sizeof(play_event));
    if (ret != 0) {
        return -1;
    }

    return effect.id;
}

/*
upload_effect(PyObject *self, PyObject *args)
{
    int fd, ret;
    PyObject* effect_data;
    ret = PyArg_ParseTuple(args, "iO", &fd, &effect_data);
    if (!ret) return NULL;

    void* data = PyBytes_AsString(effect_data);
    struct ff_effect effect = {};
    memmove(&effect, data, sizeof(struct ff_effect));

    // print_ff_effect(&effect);

    ret = ioctl(fd, EVIOCSFF, &effect);
    if (ret != 0) {
        PyErr_SetFromErrno(PyExc_IOError);
        return NULL;
    }

    return Py_BuildValue("i", effect.id);
}
*/