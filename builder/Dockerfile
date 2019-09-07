FROM golang:alpine

ENV JAVA_HOME /usr/lib/jvm/default-jvm

RUN apk add --no-cache --virtual=.build-deps \
        bash \
        cmake \
        curl \
        freetype-dev \
        g++ \
        gcc \
        git \
        libjpeg-turbo-dev \
        libpng-dev \
        libstdc++ \
        linux-headers \
        make \
        musl-dev \
        openblas-dev \
        openjdk8 \
        patch \
        perl \
        python3 \
        python3-dev \
        py-numpy-dev \
        rsync \
        sed \
        swig \
        zip \
    && cd /tmp \
    && pip3 install --no-cache-dir wheel \
&& $(cd /usr/bin && ln -s python3 python)

# ===================================================
# Download Tensorflow
# ===================================================
ENV TENSORFLOW_VERSION 1.14.0

# WORKDIR /tmp
WORKDIR /go/src/github.com/tensorflow
RUN git clone https://github.com/tensorflow/tensorflow.git


# ===================================================
# Build Tensorflow Lite
# ===================================================

WORKDIR /go/src/github.com/tensorflow/tensorflow

COPY .Makefile ./tensorflow/lite/experimental/c
RUN ./tensorflow/lite/tools/make/download_dependencies.sh
RUN make -f ./tensorflow/lite/tools/make/Makefile
RUN make -f ./tensorflow/lite/experimental/c/Makefile
RUN cp ./tensorflow/lite/tools/make/gen/linux_x86_64/lib/libtensorflow-lite.a /usr/local/lib/
RUN cp ./tensorflow/lite/tools/make/gen/linux_x86_64/lib/libtensorflow-lite.a /home
RUN cp ./tensorflow/lite/experimental/c/libtensorflowlite_c.so /usr/local/lib/
RUN cp ./tensorflow/lite/experimental/c/libtensorflowlite_c.so /home


# FROM scratch
# COPY --from=0 /tmp/tensorflow/bazel-genfiles/tensorflow/lite/experimental/c/libtensorflowlite_c.so /tensorflow/libtensorflowlite_c.so
# RUN cp /go/src/github.com/tensorflow/tensorflow/bazel-genfiles/tensorflow/lite/experimental/c/libtensorflowlite_c.so /go/src/github.com/tensorflow/tensorflow/tensorflow/lite/experimental/c/libtensorflowlite_c.so
# RUN cp /go/src/github.com/tensorflow/tensorflow/bazel-genfiles/tensorflow/lite/experimental/c/libtensorflowlite_c.so /usr/local/lib/libtensorflowlite_c.so
# RUN cp /go/src/github.com/tensorflow/tensorflow/bazel-genfiles/tensorflow/lite/experimental/c/libtensorflowlite_c.so /home/libtensorflowlite_c.so
