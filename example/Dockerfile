FROM jonarod/tensorflow_lite_alpine

FROM alpine

# Our program compiled with Golang
# From a gcc environment:
# CC=/usr/local/musl/bin/musl-gcc go build --ldflags '-w' -o classify .
# or, from jonarod/tensorflow_lite_alpine:
# docker run --rm -it -v $GOPATH:/go jonarod/tensorflow_lite_alpine:builder /bin/sh -c "cd /go/src/github.com/Jonarod/tf_classifier; go build -ldflags '-w' -o classify ."
COPY ./classify-musl-t4 /home

# Our model: mobilenet v1
COPY ./mobilenet_quant_v1_224.tflite /home
# Model's labels
COPY ./labels.txt /home
# Our test image
COPY ./peacock.png /home

# Copy the libtensorflowlite_c.so shared library
COPY --from=0 /home /usr/local/lib
# Let linker to be aware of that new library
RUN ldconfig /usr/local/lib
# `ldd -ltensorflowlite_c` indicates we still requires libstdc++, so we install it
RUN apk add --no-cache libstdc++

WORKDIR /home

ENTRYPOINT [ "/home/classify" ]
CMD []