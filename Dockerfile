FROM jonarod/tensorflow_lite_alpine:builder

# =======================================================
# =======================================================
FROM alpine

COPY --from=0 /home /home/
RUN apk add --no-cache libstdc++

WORKDIR /home