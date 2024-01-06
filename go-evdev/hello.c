#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <linux/input.h>

void Hello(){
    printf("Hello world\n");
}

int upload_effect(uintptr_t fd,  void *effect_data){
    struct ff_effect effect = {};
    memmove(&effect, effect_data, sizeof(struct ff_effect));
    return ioctl(fd, EVIOCSFF, &effect);
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