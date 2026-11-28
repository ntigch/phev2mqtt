# Version 1.10
* Fixed climate state showing *unknown* when climate is off on MY18/MY20 cars (they report off as 1, not 0)

# Version 1.9
* Fixed Home Assistant "Invalid option ... 'unknown'" warnings: the climate state select now shows *off* when the car reports *terminated* or *unknown*
* New read-only "climate status" sensor with the full climate state, including *terminated* and *unknown*

# Version 1.8
* New Home Assistant select "climate duration" (10, 20 or 30 minutes): the Heat, Cool and Windscreen switches and the climate state select now run for the chosen time instead of always 30 minutes. The choice is kept across restarts.

# Version 1.7
* Security update: built with Go 1.27.4; fixed GO-2026-6278 (gorilla/websocket); updated dependencies (golang.org/x/net, x/sys, x/text, x/sync, paho.mqtt, fsnotify, go-toml, mapstructure, pflag, yaml)
* Building from source now requires Go 1.27.4+ (the Debian/Raspberry Pi OS `golang-go` package is too old)

# Version 1.6
* Removed the vehicle emulator (`phev2mqtt emulator`)
* Fixed Home Assistant discovery: the climate select and charge timer switch now read the topics phev2mqtt actually publishes
* State topics are now retained, so Home Assistant keeps the last known values after a restart
* Car commands from MQTT no longer block other MQTT messages while waiting for the car
* Reconnect attempts to the car back off (up to 10s) while it is unreachable
* Smaller image: static build without libpcap or a C toolchain

# Version 1.5
* Add-on now builds from the ntigch/phev2mqtt fork instead of buxtronix/phev2mqtt (optionally pin with the PHEV2MQTT_REF build arg)
* Updated all Go dependencies (paho.mqtt, logrus, cobra, viper, x/*); requires Go 1.24+ to build
* Added CI with a dependency/vulnerability check (govulncheck) and Dependabot

# Version 1.4
* Changed "startup": "before" to "startup": "application" as "before" no longer supported and it should really start after Home Assistant

# Version 1.3
* Removed host networking as it wasn't required

# Version 1.2
* Removed fatal error if Outlander is not available

# Version 1.1
* Set inital values of mqtt_user and mqtt_password to null instead of dummy values to ensure that user enters them
* Added debug variable to allow start of container and then manual connection to it to run commands
* Added netcat test to check that Outlander is available
* Moved phev2mqtt binary from /tmp/phev2mqtt to /opt

# Version 1
Initial build
