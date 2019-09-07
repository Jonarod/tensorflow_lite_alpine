This Docker image brings Tensorflow Lite into a tiny 1.4MB C library ready for Alpine.
It is useful if you want to keep your Docker images small and as portable as possible.


# Quick start

#### Get `libtensorflowlite_c.so` precompiled for Alpine
```bash
docker run -it --rm --name libtflite jonarod/tensorflow_alpine_builder
```
then, from another terminal:

```bash
docker cp libtflite:/home $GOPATH/tensorflow/tensorflow/tensorflow/lite/experimental/c/
docker cp libtflite:/home /usr/local/lib/
```

Don't forget to:
- install `musl` first (**NOT REQUIRED ON ALPINE LINUX** as it comes preinstalled)
- add libstdc++ `apk add libstdc++`

#### Build project for Alpine

Example with Golang:

```bash
docker run --rm -it \
   -v $GOPATH:/go jonarod/tensorflow_alpine_builder:builder \
   /bin/sh -c "cd /go/<PATH-TO-YOUR-PROJECT-FROM-YOUR-LOCAL-GOPATH>; \
   go build -ldflags '-w' -o classify ."
```


> If you want to use TFLite from a classic `gcc` based environment like Ubuntu/Debian etc... This project is not for you !! In fact, building `libtensorflowlite_c.so` with `gcc` is way easier and [well documented](https://stackoverflow.com/questions/55107823/how-to-build-tensorflow-lite-with-c-api-support). I believe that when image size is a matter, then Alpine is a good choice. If size is not a concern for your project, I highly recommand you use classic TensorFlow core (not Lite) using the `ubuntu` image instead. ;)


**The good:**
- Creates the official `libtensorflowlite_c.so` file
- **Alpine ready:** The dynamically-linked dependencies are based on `musl`, the natural C compiler on Alpine Linux
- **Small:** `libtensorflowlite_c.so` alone is 1.4MB
- **Compact Docker images:**: the example project in this repo generates a final image of **10MB** including Tensorflow Lite, our code logic, Mobilenet, labels, and a sample image. In contrast, the official `tensorflow/tensorflow` image is **1.17GB** (with much much more functionnalities and models of course).


**The bad:**
- **!! For inference only !!** Training is out of the scope of this project.
- **Limited:** To make things really tiny, we use Tensorflow **LITE** and, as its name suggests, it is a lighter (yet very capable) version of Tensorflow normally used for embeded devices like Mobile phones and IoT. It has officially limited set of operations. Look at the official docs [here](https://www.tensorflow.org/lite).
- **Performance:** Tensorflow Lite is optimized for Mobile and IoT architectures like ARM. It is not officially created for x86_64 architectures (like servers and laptops). For instance, GPU is not available. Also, CPU usage might not be optimized under specific conditions. This means, you may find better performances using the core (yet bulkier) Tensorflow.

**The cool part :)** is that, even though this repo is specifically created for TFLite, you may apply its techniques to reduce the fully-fledged core TensorFlow by uncommenting some parts of the Dockerfile.


# So what ?! Should I learn C to use this ?

**No !!**

If you have C skills, you can in fact use the library as-is right away, the API is here: [c_api.h](https://github.com/tensorflow/tensorflow/blob/master/tensorflow/c/c_api.h).

Otherwise, cool and skilled people already mapped the C API into higher level languages like [Go](https://golang.org/) and [Rust](https://www.rust-lang.org/).

- TFLite for Go: [go-tflite](https://github.com/mattn/go-tflite)
- TFlite for Rust: [tflite crate](https://crates.io/crates/tflite)

You can then use the `libtensorflowlite_c.so` library with these projects. Have a look at the `/example` folder exposing usage with `go-tflite`.


# Usage

This Docker image generates a `libtensorflowlite_c.so` built with `make` in Alpine Linux. This makes Tensorflow Lite's C API available in Alpine Linux using `musl-gcc`.

The resulting file is 1.4MB, depends on `libstdc++.so.6` and is compatible with `musl` linker and environment.


### Get the generated library

There are three ways to get the `libtensorflowlite_c.so` library.

##### Using multi-step Dockerfile

Directly use the docker image as a base image and import the file located at `/home`. Docker allowing multistep builds, you can just:


```Dockerfile
# Import the image and name this first step as "builder" (for example)
FROM jonarod/tensorflow_alpine_builder as libtflite

# You want to start your image with Alpine (this is the point...)
FROM alpine

# Do other things to construct your image...
# ...
# Do other things to construct your image...

# ================================
# HERE IS THE INTERESTING PART
# ================================
# We extract the `libtensorflowlite_c.so` file from the first step, 
# and put it into the `/usr/local/lib` of this very image
COPY --from=libtflite /home /usr/local/lib

# IMPORTANT: let the linker system aware of the new library we put into /usr/local/lib
RUN ldconfig /usr/local/lib

# `libtensorflowlite_c.so` depends on libstdc++, so we install it
RUN apk add --no-cache libstdc++
# ================================

# Do other things to construct your image...
# ...
# Do other things to construct your image...

```

##### Copying from running docker

One could also pull this image, run it, and then use `docker cp` to copy the `libtensorflowlite_c.so` from the running docker to the host.

This will pull the image if not already, then run it interactively and name the container for easier copy later.

```bash
docker run -it --rm --name libtflite jonarod/tensorflow_alpine_builder
```
then, from another terminal, we can: 

```bash
docker cp libtflite:/home ./lib
```

This will eventually copy all the files contained in `/home` (only the `libtensorflowlite_c.so` is there...) right into the current host's `./lib` directory.

Then you'll have the `musl` compatible `libtensorflowlite_c.so` in your system to do whatever you need.

To use the library with `go-tflite` specifically, don't forget that, at build time (`go build ...`), Go looks for `libtensorflowlite_c.so` using linker. So you need to place the shared library according to linker's lookup directories. Generally, it is under `/usr/lib` or `/usr/local/lib`. If, for some reasons, you get an error like `/bin/ld -ltensorflowlite_c cannot be found`, this is because linker can't find the `libtensorflowlite_c.so`. To sort this out, run `ld -ltensorflowlite_c --verbose` and see where linker is actually looking for the file, then place it accordingly. Finally, executing `sudo ldconfig /path/to/the/libary` might help linker to get aware of your lib.

At runtime (`go run ...`), `go-tflite` will look for the library in `$GOPATH/tensorflow/tensorflow/tensorflow/lite/experimental/c`, so here again make sure you have this in place.


##### Build it yourself

Well, you can also grab the original `Dockerfile` in the `/builder` folder, and build a new version of the `libtensorflowlite_c.so`. The advantage is that you get more flexibility and can control some things easily. For instance, you can generate the library with a newer `bazel` and/or newer `TFLite` version. As a bonus, I left some comments to generate the core `Tensorflow` using the same workflow if you need it.

For that, just copy the content of the `Dockerfile`, tweak it to your needs, then just:

```bash
docker build --rm -t <YOUR-NAME>/<YOUR-FINAL-IMAGE-NAME> .
```

It took around 30mn to build the TFLite lib on a 8GB RAM, 4CPU @2GHz machine.
Be careful if you decide to build the core Tensorflow as it is way heavier and might take some HOURS to compile...


# Credits

I could make this happen thanks to these interesting reads:

[Official TFLite C docs](https://www.tensorflow.org/lite)

[go-tflite](https://github.com/mattn/go-tflite)

[bazel-alpine-package](https://github.com/davido/bazel-alpine-package)

[Dominik Honnef's blog](https://dominik.honnef.co/posts/2015/06/go-musl/)

[This StackOverflow answer](https://stackoverflow.com/questions/55125977/how-to-build-tensorflow-lite-as-a-static-library-and-link-to-it-from-a-separate/55144057#55144057)

[This StackOverflow answer](https://stackoverflow.com/questions/55107823/how-to-build-tensorflow-lite-with-c-api-support)

[Tensorflow-bin](https://github.com/PINTO0309/Tensorflow-bin)

[blog by Tomas Sedovic](https://aimlesslygoingforward.com/blog/2014/01/19/bundling-shared-libraries-on-linux/)