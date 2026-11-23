# phev2mqtt - Mitsubishi Outlander PHEV to MQTT gateway

Utility to interact with a Mitsubishi Outlander PHEV via the Wifi remote
control protocol.
Forked from https://github.com/Will-Honor/phev2mqtt who got it from https://github.com/buxtronix/phev2mqtt

This fork adds:

 * NetworkManager scripts and docs for reliable car Wifi reconnects (see `scripts/`,
   tested on Debian trixie)
 * Fixes for car and MQTT reconnects, plus Home Assistant discovery fixes (the climate
   select and charge timer switch now show their state)
 * Retained MQTT state, so Home Assistant keeps the last known values after a restart
 * Car commands from MQTT are queued, so a slow car doesn't block other MQTT messages
 * Updated dependencies with CI vulnerability checks
 * A Home Assistant add-on that builds from this fork (see [addon/README.md](addon/README.md))

The vehicle emulator from the original project has been removed.

Inspired by https://github.com/phev-remote/ but written entirely in Go.

For further hacking, read the [protocol documentation](protocol/README.md).

Tested against a MY18 vehicle and my own MY20.

![Home Assistant Screenshot](phev-ha.png)

## Supported functionality

 * MQTT proxy to Phev
 * Home Assistant discovery
 * Register client to car
 * Fetch battery, charge, door, light status
 * Set lights and charge enable
 * Near-instant response to commands
 * *tested on MY18 & MY20*

Also includes some debugging utilities (see [Debugging utilities](#debugging-utilities)).

## Requirements

 * Go 1.24 or newer
 * A Wifi interface that can connect to the car

## Licence, etc

Licenced under the GPLv3 (or later) by the original developer:

Copyright 2021 Ben Buxton <bbuxton@gmail.com>

Modifications in this fork:

Copyright 2026 ntigch

Contributions and PRs are welcome.

## Getting started.

### Compiling

#### Install Go

 * Download and install the latest [Go compiler](https://golang.org/dl/) (Go 1.24 or newer is required)
   * Your distro packager may have a version thats too old
   * For raspbian choose the ARMv6 release

#### Install PCAP dev libraries

 * Optionally, you may want to have libpcap-dev package installed (if building with `-tags pcap`.)

#### Download, extract, and compile phev2mqtt

 * Download and extract the phev2mqtt archive, or `git clone https://github.com/ntigch/phev2mqtt.git`
 * Go into its top level directory and run `go build`
 * Verify it runs with `./phev2mqtt -h`

### Connecting to the vehicle.

#### Configure Wifi client on system running mqtt2phev

On your computer running the phev2mqtt tools, configure a new Wifi connection to the
car's SSID (it looks like `REMOTE45xxxx`; the password is on the car's Wifi
information screen). The Wifi MAC address of your computer must be the one registered
with the car (see below).

On a Raspberry Pi, follow [Raspberry Pi setup with auto-start](#raspberry-pi-setup-with-auto-start).
To see which network stack your system uses, run `systemctl is-active NetworkManager`
(`active` means NetworkManager; use the main instructions below, otherwise see the
legacy section).

#### Register the client to the car

Follow the [Mitsubishi instructions](https://www.mitsubishi-motors.com/en/products/outlander_phev/app/remote/)
to find the Wifi credentials provided with the car.

Verify that your Wifi connection to the car is established - your local IP address
should be 192.168.8.47.

Follow the [Mitsubishi instructions](https://www.mitsubishi-motors.com/en/products/outlander_phev/app/remote/)
and put the car into registration mode ("Setup Your Vehicle"). You may need to
re-establish the Wifi connection.

Register by running `phev2mqtt client register` and you should shortly see a message
indicating successful registration.

#### Testing the tool

Once connected to the car, you can sniff for messages by running *phev2mqtt client watch*.
The phone client needs to be disconnected for this to work.
You'll see a bunch of data go by - some of those will be decoded into readable
messages such as charge and AC status.

### MQTT Gateway

The primary feature of this code is to run as a proxy between the car and
MQTT. Registers with car status are sent to MQTT, both as raw register
values and decoded functional values. Commands sent on MQTT topics can
be used to control certain aspects of the vehicle.

Start the MQTT gateway with:

`./phev2mqtt client mqtt --mqtt_server tcp://<your_mqtt_address>:1883/ [--mqtt_username <mqtt_username>] [--mqtt_password <mqtt_password>]`

The gateway keeps retrying the car connection. While the car can't be reached it
waits a little longer between attempts (up to 10 seconds), and goes back to fast
retries once it has been connected.

#### Options

| Flag | Default | Description |
|---|---|---|
| `--address` | `192.168.8.46:8080` | Address of the car |
| `--mqtt_server` | `tcp://127.0.0.1:1883` | MQTT broker address |
| `--mqtt_username` / `--mqtt_password` | | MQTT login |
| `--mqtt_topic_prefix` | `phev` | Prefix for all topics below |
| `--mqtt_disable_register_set_command` | `false` | Don't subscribe to `phev/set/#`, so the car can't be controlled via MQTT |
| `--ha_discovery` | `true` | Publish Home Assistant discovery configs |
| `--ha_discovery_prefix` | `homeassistant` | Home Assistant discovery prefix |
| `--update_interval` | `5m` | How often to ask the car for a full register update |
| `--wifi_restart_time` | `0` (off) | Run `--wifi_restart_command` if the car hasn't been connected for this long |
| `--wifi_restart_retry_time` | `2m` | Minimum time between Wifi restart attempts |
| `--wifi_restart_command` | `sudo ip link set wlan0 down && sleep 3 && sudo ip link set wlan0 up` | Command run to restart Wifi; empty disables it |

Options can also be set in a config file (`--config`, default `$HOME/.phev2mqtt.yaml`)
or as upper-case environment variables (e.g. `MQTT_SERVER`).

#### Topics

The following topics are published. They are retained, so Home Assistant shows the
last known values after it restarts; `phev/available` shows whether they are current.

| Topic/prefix | Description |
|---|---|
| phev/register/[register] | Raw values of each register, as hex strings |
| phev/available | Wifi connection status to car. *online* or *offline* |
| phev/battery/level | Current drive battery level as a percent |
| phev/climate/state | Climate state: *off*, *cool*, *heat*, *windscreen*, *terminated* or *unknown* |
| phev/climate/[mode] | Per-mode state for *cool*, *heat*, *windscreen*: *on* or *off* |
| phev/charge/charging | Whether the battery is charging. *on* or *off* |
| phev/charge/plug | If the charging plug is *unplugged* or *connected*. |
| phev/charge/remaining | Minutes left, if charging. |
| phev/door/locked | Lock state: *closed* when locked, *open* when unlocked |
| ~~phev/door/front_left~~ | Deprecated alias of front_passenger |
| ~~phev/door/front_right~~ | Deprecated alias of driver |
| phev/door/front_passenger | State of doors. *closed* or *open* |
| phev/door/driver | State of doors. *closed* or *open* |
| phev/door/rear_left | State of doors. *closed* or *open* |
| phev/door/rear_right | State of doors. *closed* or *open* |
| phev/door/bonnet | State of doors. *closed* or *open* |
| phev/door/boot | State of doors. *closed* or *open* |
| phev/lights/parking | Parking lights. *on* or *off* |
| phev/lights/head | Head lights. *on* or *off* |
| phev/lights/hazard | Hazard lights. *on* or *off* |
| phev/lights/interior | Interior lights. *on* or *off* |
| phev/vin | Discovered VIN of the car |
| phev/registrations | Number of wifi clients registered to the car |
| phev/ecuversion | ECU software version |

The following topics are subscribed to and can be used to change state on the car.
Commands are queued and sent to the car one at a time; they are dropped if the car
isn't connected.

| Topic/prefix | Description |
|---|---|
| phev/set/register/[register] | Set register 0x[register] to value 0x[payload] |
| phev/set/parkinglights | Set parking lights *on* or *off* |
| phev/set/headlights | Set head lights *on* or *off* |
| phev/set/cancelchargetimer | Cancel charge timer (any payload) |
| phev/set/climate/[mode] | Set climate mode (cool/heat/windscreen/off). Payload is the duration in minutes: *10*, *20*, *30* (*on* = 30), or *off* |
| phev/set/climate/mode | Set climate mode from the payload (*cool*, *heat*, *windscreen* or *off*), for 30 minutes |
| phev/set/climate/state | `[payload]=reset` clears "terminated" state |
| phev/connection | Change car connection state to (on/off/restart) |
| phev/settings/dump | Log the car settings received so far, then clear them |

#### Home Assistant discovery

The client supports [Home Assistant MQTT Discovery](https://www.home-assistant.io/docs/mqtt/discovery/) by default.

The entities appear once phev2mqtt has connected to the car and read its VIN. They
are published again when Home Assistant or the MQTT broker restarts. You can
search for "phev" in your entity list. Your car should also appear as a device
in the Devices tab.

You can disable this with `--ha_discovery=false` or change the discovery prefix, the default is `--ha_discovery_prefix=homeassistant`.

#### Raspberry Pi setup with auto-start

It's useful to have the tool auto-start when running on e.g a Raspberry Pi. The following
describes how to set this up.

##### 1. Car Wifi (NetworkManager: Raspberry Pi OS Bookworm and newer, Debian 13 trixie)

Current Raspberry Pi OS releases manage Wifi with NetworkManager, not `wpa_supplicant.conf`.
Run the helper script with the car's SSID, password and the MAC address the car was
registered with:
```
sudo scripts/setup-car-wifi-nm.sh REMOTE45bhds 'blahblahbla12314' ee:4d:ec:de:7a:91
```

The last argument is the MAC address the car was registered with. The script
configures plain WPA2-PSK, a hidden SSID on 2.4 GHz, Wifi power saving off, a fixed
MAC, and unlimited autoconnect retries. (Tested on Debian 13 "trixie".) Without these NetworkManager can get stuck: the car's old access point
may reject NetworkManager's default security offer (`ASSOC-REJECT status_code=16`),
and after a few failures NetworkManager stops auto-connecting for good (so the Pi
doesn't reconnect when the car returns).

As a safety net, also make `phev2mqtt` re-activate the connection if the car
hasn't been seen for a while. Use this `ExecStart` in the service below (the
service user needs passwordless `sudo` for `nmcli`):

```
ExecStart=/usr/local/bin/phev2mqtt --config=/dev/null client mqtt --mqtt_server tcp://192.168.0.88:1883 --wifi_restart_time=2m --wifi_restart_retry_time=1m --wifi_restart_command="sudo nmcli device disconnect wlan0; sleep 2; sudo nmcli connection up REMOTE45bhds" -v=debug
```

Activating the connection again also clears NetworkManager's autoconnect block.

##### 2. Service

- Add the following to `/etc/systemd/system/phev2mqtt.service`, updating the MQTT address to
suit your setup:

```
[Unit]
Description=phev2mqtt service script
StartLimitIntervalSec=5
After=syslog.target network.target

[Service]
Type=exec
ExecStart=/usr/local/bin/phev2mqtt --config=/dev/null client mqtt --mqtt_server tcp://192.168.0.88:1883 -v=debug

# Restart script if stopped
Restart=always
# Wait 30s before restart
RestartSec=30s

# Tag things in the log
# View with: sudo journalctl -f -u phev2mqtt -o cat
SyslogIdentifier=phev2mqtt

StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

- Copy the `phev2mqtt` binary to /usr/local/bin and make sure it's executable.

- Start the service with `sudo systemctl start phev2mqtt.service`

- Enable the service to run at boot, with `sudo systemctl enable phev2mqtt.service`.

- Restart the Pi and verify that it can connect to the car. Also run `ifconfig` and check
that the `wlan0` interface has the correct mac address. You should also see this interface
have the IP address `192.168.8.47`.

- Verify that the phev2mqtt service is communicating with the car, by checking
the logs: `sudo journalctl -f -u phev2mqtt -o cat`

##### Legacy: Raspberry Pi OS Bullseye and older (dhcpcd + wpa_supplicant)

Only use this on older systems that do not run NetworkManager; do not combine it with
the NetworkManager setup above.

- Edit or add to `/etc/systemd/network/00-default.link` with the following:

```
[Match]
# This should be the 'real' (default) mac address of the Pi's wireless interface.
MACAddress=b8:27:eb:50:c0:52

[Link]
# This should be the MAC address to use to connect to the car, per above.
MACAddress=ee:4d:ec:de:7a:91
NamePolicy=kernel database onboard slot path

```

- Add the car's Wifi info to `/etc/wpa_supplicant/wpa_supplicant.conf`:

```
ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1
country=AU

network={
	ssid="REMOTE45bhds"
	scan_ssid=1
	psk="blahblahbla12314"
}

```

### Debugging utilities

All `client` commands connect to the car (use `--address` if it isn't at `192.168.8.46:8080`).
The official app must be disconnected.

| Command | Description |
|---|---|
| `phev2mqtt client watch` | Show register updates from the car as they arrive |
| `phev2mqtt client set <reg>:<value> ...` | Set registers (hex), e.g. `0b:02` sets register 0x0b to 0x02. `--wait_duration` (default 10s) waits after connecting, `--send_interval` (default 1s) between registers |
| `phev2mqtt client register` / `unregister` | Register or unregister this machine's Wifi MAC address with the car |
| `phev2mqtt decode hex <hex> ...` | Decode raw messages given as hex strings |
| `phev2mqtt decode file <file>` | Decode raw messages from a file of hex (whitespace is ignored) |
| `phev2mqtt decode pcap <file>` | Decode a packet capture (needs a `-tags pcap` build, see below) |

### Sniffing the official client

Further development of this library can be done with a packet dump of the official
Mistubishi app.

A number of sniffer apps for phones are available for this. Two that the author have
used are *Packet Capture* and *PCAP Remote*. These do not require root access, yet
can successfully sniff the traffic into PCAP files for further analysis.

*Packet Capture* can save the PCAP files to your local phone storage which you can
then extract off the phone.

*PCAP Remote* is a little more involved, but allows for live sniffing of the traffic.

Once you have downloaded the PCAP file(s) from the phone, you can analyse them with
the command *phev2mqtt decode pcap <filename>*. First build a `phev2mqtt` with pcap features:
`go build -tags pcap`; you will need libpcap for this. Adjust the verbosity level (`-v`) between
`info`, `debug` and `trace` for more details.

Additionally, the flag `--latency` will use the PCAP packet timestamps to decode
the packets with original timings which can help pinpoint app events.

You can also specify *tcp:<host>:<port>* which will connect to that host/port
over TCP and decode that traffic - useful when live sniffing to a TCP service.
