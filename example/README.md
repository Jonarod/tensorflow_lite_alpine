In this example, we'll use the [go-tflite](https://github.com/mattn/go-tflite) package written in Go that consumes the `libtensorflowlite_c` C API.

We'll use one of the cool examples provided by @mattn in the `go-tflite` repo: [examples/label_image](https://github.com/mattn/go-tflite/tree/master/_example/label_image).

Go allows to build a standalone static executable extremely portable. We could leverage Go to minimize our TFlite project. Unfortunately, Tensorflow does not provide a way to compile a fully static `libtensorflowlite_c.a` yet ([follow this issue](https://github.com/bazelbuild/bazel/issues/1920)), so we need the dynamically-linked `libtensorflowlite_c.so` which `go-tflite` depends on.

In `go-tflite`'s README there is in fact a small example to help you build the library with a `make` file. However, **the build output will not work with the lighter docker image `alpine`**. In fact, chances are that your system will build the required lib using the C compiler `gcc` which is, by very far, the most common. Unfortunately, In order to be this lightweight, `alpine` uses `musl` as a lighter C compiler alternative, making `gcc` dependent libraries incompatible. And this is what this repository is all about: give an easy way to build `libtensorflowlite_c.so` compatible with `musl` in order to make your projects Alpine-ready (and tiny!).


First of all, it is assumed that all `go-tflite` dependencies have been installed using `go get`. On top of that, `go-tflite` needs the tensorflow repository to be located at `$GOPATH/src/github.com/tensorflow`. So let's do this quickly:
```
mkdir -p $GOPATH/src/github.com/tensorflow
cd $GOPATH/src/github.com/tensorflow
git clone https://github.com/tensorflow/tensorflow.git
```

Now the last step is to get `libtensorflowlite_c.so`.


# Using Docker

This is the easiest and quickest way to build our project into an alpine compatible executable.

```bash
docker run --rm -it \
   -v $GOPATH:/go jonarod/tensorflow_lite_alpine:builder \
   /bin/sh -c "cd /go/<PATH-TO-YOUR-PROJECT-FROM-YOUR-LOCAL-GOPATH>; \
   go build -ldflags '-w' -o classify ."
```

Ok let's decompose it line by line:
- `docker run --rm -it`: run the container
- `-v $GOPATH:/go`: this part is **IMPORTANT**. It mounts your local `$GOPATH` into the containers `$GOPATH`. It is important as it will load in the container your local UP TO DATE version of `tensorflow`, `go-tflite`, `resize`... as well as your project.
- `/bin/sh -c "cd /go/<PATH-TO-YOUR-PROJECT-FROM-YOUR-LOCAL-GOPATH>;`: just replace `<...>` with the actual path of your project AFTER your `$GOPATH`. For example, in our case, our project was in `$GOPATH/src/github.com/wemakeai/example-go-tflite-alpine`. We would then replace the path with: `cd /go/src/github.com/wemakeai/example-go-tflite-alpine;`.
- `go build -ldflags '-w' -o classify .`: builds our project using the linker flag `-w` (disable DWARF debugging table) and name the output executable as `classify` (change that like you want).

By executing this long command, the container will load our local files and build them as needed using its own copy of `libtensorflowlite_c`. Since we use a volume, the resulting output file should be located in your Go project's folder (`$GOPATH/src/<path-to-our-project>/<NAME-YOUR-OUTPUT>`) and it should be compatible with Alpine ONLY.

If you wish to use this output outside of Alpine, you may consider installing `musl` like described earlier, and use `CC=/usr/local/musl/bin/musl-gcc` before executing your program.



# Using non-Alpine environment (Ubuntu, Debian...)

Let's grab the correct library and copy it to both `$GOPATH/tensorflow/tensorflow/tensorflow/lite/experimental/c/` and `/usr/local/lib/`.

```bash
docker run -it --rm --name libtflite jonarod/tensorflow_lite_alpine
```
then, from another terminal:
```bash
docker cp libtflite:/home $GOPATH/tensorflow/tensorflow/tensorflow/lite/experimental/c/
docker cp libtflite:/home /usr/local/lib/
```

Now that we have the correct file, we would like to try our project, right ?!

Not so fast !

Remember that the library we just got is compiled using `musl`, so it will not be compatbile as-is. We need to install `musl` to compile the `main.go` into an Alpine-compatible static program. 

```bash
wget http://www.musl-libc.org/releases/musl-1.1.22.tar.gz
tar -xvf musl-1.1.22.tar.gz
cd musl-1.1.22
./configure --enable-wrapper=gcc
make
sudo make install
```

This will install `musl` into `/usr/local/musl` from which we can use the lightweight `gcc` alternative `musl-gcc` located at `/usr/local/musl/bin/musl-gcc`.

We can now compile our `main.go`:

```bash
CC=/usr/local/musl/bin/musl-gcc go build --ldflags '-w' -o classify .
```

This command first instructs Go to use `musl` as C Compiler, then use the linker flags `-w` which is a hacky way to reduce our package further by removing the DWARF symbol table (this will prevent debugging the app, but in a production-ready context we might not really need this). Finally, we use classic C compiler flag `-o` to give a name to our output file.

The output `classify` will then be a program dynamically-linked to `libtensorflowlite_c.so` so we need both files to distribute our program. Also, remember that `libtensorflowlite_c.so` depends on `libstdc++.so.6` (`apk add libsdtc++`) so we actually need those 3 files for our program to work as exepcted.


# Use our Alpine-ready build into Alpine

Finally !! We've built our project using `musl` and are now ready to ship it !!
Here is an example of a minimal Dockerfile:

```Dockerfile
FROM jonarod/tensorflow_lite_alpine
FROM alpine
COPY ./classify /home
COPY ./mobilenet_quant_v1_224.tflite /home
COPY ./peacock.png /home
COPY --from=0 /home /usr/local/lib
RUN apk add --no-cache libstdc++
ENTRYPOINT [ "/home/classify" ]
CMD []
```

After building this Dockerfile, running it should execute our `classify` program which reads our `peacock.png` image, processes it against the quantized `mobilenet_v1` model which returns the classification for this image. 

**The resulting project alltogether is 14MB !! including the image, model and labels list** :)
