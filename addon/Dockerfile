ARG BUILD_FROM
FROM $BUILD_FROM as builder

RUN apk update && apk add --no-cache \
	g++ \
	gcc \
	git \
	libpcap-dev

WORKDIR /tmp
# Build this fork; pin PHEV2MQTT_REF to a tag or commit for reproducible builds
ARG PHEV2MQTT_REF=main
RUN git clone https://github.com/ntigch/phev2mqtt.git && \
    git -C phev2mqtt checkout ${PHEV2MQTT_REF}
COPY --from=golang:alpine /usr/local/go/ /usr/local/go/
RUN cd /tmp/phev2mqtt && \
    /usr/local/go/bin/go build

FROM $BUILD_FROM
RUN apk update && apk add --no-cache \
	libpcap-dev

COPY --from=builder /tmp/phev2mqtt/phev2mqtt /opt/phev2mqtt

COPY run.sh /opt
RUN chmod a+x /opt/run.sh

CMD [ "/opt/run.sh" ]
