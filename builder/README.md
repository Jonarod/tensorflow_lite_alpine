For some reasons, I found `bazel` created a tflite library with much inferior performances than `make` (probably because it does not compile some CPU-wise support libraries... I did not investigate much). The `Dockerfile.bazel` is left for reference only but **should not be used**.

The `Dockerfile` uses the Makefile to build TFlite for C instead. You should use it too if needed.