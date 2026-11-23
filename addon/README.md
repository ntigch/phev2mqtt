# PHEV Home Assistant add-on

Runs the `phev2mqtt client mqtt` gateway as a local Home Assistant add-on. It builds
phev2mqtt from this fork (`ntigch/phev2mqtt`, `master` branch by default). The car
appears in Home Assistant via MQTT discovery.

## Before you start

 * The Home Assistant host must already be connected to the car's Wifi, and its MAC
   address registered with the car (see the main [README](../README.md)). Check that
   you can ping `192.168.8.46` from the host.
 * You need an MQTT broker, e.g. the Mosquitto broker add-on, with MQTT set up in Home
   Assistant.

## Install

Copy the files in this directory into a directory (e.g. `phev`) in your local add-ons:

| How you connect | Path |
|---|---|
| SSH into the main OS | `/usr/share/hassio/addons/local/phev` |
| Terminal add-on | `/addons/phev` (or the symlink `/root/addons/phev`) |
| `hassio_supervisor` docker container | `/data/addons/local/phev` |

Then refresh the add-on store, install **PHEV** from the local add-ons, and set the
options below before starting it. More about local add-ons:
https://developers.home-assistant.io/docs/add-ons/tutorial/

To build a fixed version instead of the latest `master`, set the `PHEV2MQTT_REF` build
arg in the `Dockerfile` to a tag or commit.

## Options

| Option | Description |
|---|---|
| `mqtt_server` | Broker as `host:port`, without `tcp://`. On a docker install with the Mosquitto add-on this is often `172.30.32.1:1883` (it can differ per install). |
| `mqtt_user` / `mqtt_password` | MQTT login |
| `debug` | If on, the container just starts and sleeps, so you can open a shell in it and run `/opt/phev2mqtt` by hand, e.g. `/opt/phev2mqtt client watch` |

## Notes

 * MQTT state is retained, so entities keep their last known values when Home
   Assistant restarts. They show as unavailable while the car isn't connected.
 * The image contains a static build without the pcap decoder, so `decode pcap` is
   not available inside the add-on.
 * On 32-bit Raspberry Pi, recent Alpine images can fail to build with old
   `libseccomp` versions (including the ones in the Raspbian repos). If so, install a
   newer version manually:

   ```bash
   wget http://ftp.us.debian.org/debian/pool/main/libs/libseccomp/libseccomp2_2.4.4-1~bpo10+1_armhf.deb
   dpkg -i libseccomp2_2.4.4-1~bpo10+1_armhf.deb
   ```

See [CHANGELOG.md](CHANGELOG.md) for changes between versions.
