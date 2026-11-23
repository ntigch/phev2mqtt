ARG BUILD_FROM
FROM $BUILD_FROM as builder

RUN apk add --no-cache git

WORKDIR /tmp
# Build this fork; pin PHEV2MQTT_REF to a tag or commit for reproducible builds
ARG PHEV2MQTT_REF=master
RUN git clone https://github.com/ntigch/phev2mqtt.git && \
    git -C phev2mqtt checkout ${PHEV2MQTT_REF}
COPY --from=golang:alpine /usr/local/go/ /usr/local/go/
# Static build without the pcap decoder, so no C toolchain or libpcap needed
RUN cd /tmp/phev2mqtt && \
    CGO_ENABLED=0 /usr/local/go/bin/go build

FROM $BUILD_FROM

COPY --from=builder /tmp/phev2mqtt/phev2mqtt /opt/phev2mqtt

COPY run.sh /opt
RUN chmod a+x /opt/run.sh

CMD [ "/opt/run.sh" ]
