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
 * A Home Assistant select to choose how long climate runs (10, 20 or 30 minutes)
 * Updated dependencies with CI vulnerability checks
 * A Home Assistant add-on that builds from this fork (see [addon/README.md](addon/README.md))

The vehicle emulator from the original project has been removed.

Inspired by https://github.com/phev-remote/ but written entirely in Go.

For further hacking, read the [protocol documentation](protocol/README.md).

Tested against a MY18 vehicle and my own MY20.

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

 * Go 1.27.4 or newer
 * A Wifi interface that can connect to the car

On a Raspberry Pi, skip this and follow the [Quick start](#quick-start-raspberry-pi-step-by-step) below.

## Quick start: Raspberry Pi (step by step)

This is the recommended way to run phev2mqtt: a small Raspberry Pi that sits near the car,
joins the car's Wifi, and passes everything on to your MQTT broker (and so to Home
Assistant). It is written for the **current Raspberry Pi OS (Debian 13 "trixie")**. For older
versions see [Older Raspberry Pi OS versions](#older-raspberry-pi-os-versions).

> These steps were put together from the project's scripts and standard Raspberry Pi OS
> usage. Check each command's output as you go, and open an issue if something doesn't
> match.

### How it fits together

```
Car (Wifi hotspot) ~~wifi~~> Raspberry Pi --ethernet--> home network --> MQTT broker --> Home Assistant
```

 * The Pi's **Wifi** is used only to talk to the car. It cannot be on your home Wifi at the same time.
 * The Pi's **Ethernet** port connects it to your home network, so it can reach your MQTT broker.
 * The Pi must be within Wifi range of the car (usually parked at home, with the Pi indoors
   near a window or in the garage).

### What you need

 * A Raspberry Pi with built-in Wifi: Pi 3B+, 4, 5 or Zero 2 W. (To boot from USB, as
   recommended below, use a 3B+ or newer. The Zero 2 W has no Ethernet port.)
 * A good-quality **USB stick** (recommended), or a good SD card
 * A decent power supply for your Pi model, and an Ethernet cable to your router
 * The car's Wifi name (SSID, like `REMOTE45xxxx`) and password, from the car's Wifi
   information screen (see the [Mitsubishi instructions](https://www.mitsubishi-motors.com/en/products/outlander_phev/app/remote/))
 * The address of your MQTT broker, e.g. `192.168.1.50` (the Mosquitto add-on in Home
   Assistant works; your Home Assistant IP address is often the right one)
 * A computer to prepare the drive and log in to the Pi

### Step 1: Put Raspberry Pi OS on a USB drive (or SD card)

The Pi writes logs constantly. Because of this, an SD card slowly wears out and may
corrupt after months of 24/7 use. A good USB stick is much more reliable, so this guide
recommends booting from USB. It is the same procedure as for an SD card.

1. Install [Raspberry Pi Imager](https://www.raspberrypi.com/software/) on your computer.
2. Plug in your USB drive (or SD card). **Everything on it will be erased.**
3. In Imager choose your Pi model, then **Raspberry Pi OS (other) -> Raspberry Pi OS Lite (64-bit)**,
   then select your USB drive as the storage.
4. When asked to customise, set a hostname (e.g. `phev`), a username and password, and
   enable **SSH**. You don't need to enter Wifi details: the Pi uses Ethernet.
5. Write the drive.

**Booting from USB:**

 * **Pi 3B+, 4 and 5:** with a recent firmware the Pi boots from USB automatically when
   there's **no SD card inserted**. Just plug in the USB drive and the network cable, then power on.
 * If the Pi doesn't start from USB, its bootloader is probably old or the boot order
   prefers SD. Boot it once from any SD card with Raspberry Pi OS, then run
   `sudo raspi-config`, choose **Advanced Options -> Bootloader Version -> Latest**
   and then **Advanced Options -> Boot Order -> NVMe/USB Boot**, reboot, then shut down and
   move your system to the USB drive. (Menu names can change; see the official docs.)
 * On a Pi 5, use the official power supply: with a weaker one the Pi limits the power to
   USB devices, which can stop a stick from working.

More details are in the official documentation:
[Raspberry Pi hardware / boot options](https://www.raspberrypi.com/documentation/computers/raspberry-pi.html)
and [raspi-config](https://www.raspberrypi.com/documentation/computers/configuration.html).

**Which stick?** (General advice; I haven't tested specific models with this project.)

 * Pick a quality stick from a known brand with decent write speeds, e.g. the SanDisk
   Extreme or Extreme Pro, or the Samsung Bar Plus or Fit Plus. Avoid cheap no-name sticks.
 * **Size:** 32-64 GB is plenty. Bigger gives no benefit here.
 * Plug it into a blue USB 3 port on a Pi 4 or 5.
 * Beware of "too good to be true" capacities and prices: fake flash drives are common.

If you stay on an SD card, use a good brand (a high-endurance card for dashcams works well).

### Step 2: Log in and update the Pi

Find the Pi's IP address in your router (or try `ping phev.local`), then log in from your computer:

```bash
ssh <your_username>@<your pi's-IP-address>
```

Update everything and install git:

```bash
sudo apt update
sudo apt full-upgrade -y
sudo apt install -y git
```

phev2mqtt needs Go 1.27.4 or newer. The `golang-go` package from `apt` is too old, so
install Go from [go.dev/dl](https://go.dev/dl/). Check your CPU type with `uname -m`
(`aarch64` means the `linux-arm64` download, `armv7l` means `linux-armv6l`). On Raspberry Pi OS you need the linux-arm64 version and can copy the command below:

```bash
# Replace the file name with the current version for your CPU from https://go.dev/dl/
wget https://go.dev/dl/go1.27.4.linux-arm64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.27.4.linux-arm64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
go version
```

You should see something like `go version go1.27.4 linux/arm64`. If you previously
installed `golang-go` with `apt`, remove it (`sudo apt remove golang-go`) so the old
version isn't used by mistake.

### Step 3: Download and build phev2mqtt

```bash
cd ~
git clone https://github.com/ntigch/phev2mqtt.git
cd phev2mqtt
go build
sudo install -m 755 phev2mqtt /usr/local/bin/phev2mqtt
cd
```

The build takes a minute or two on a Pi and prints nothing when it works. Check:

```bash
phev2mqtt -h
```

You should see the help text.

### Step 4: Choose the Wifi MAC address for the car

The car only talks to Wifi devices it has been introduced to, identified by their **MAC
address**. You need to pick the MAC address the Pi will use on the car's Wifi:

 * **Easiest if you've never connected to the car before:** make one up, in the form `02:xx:xx:xx:xx:xx` (using hex digits `0-9` and
   `a-f`), e.g. `02:11:22:33:44:55`. Pick something unique and write it down. You will
   register it with the car in step 6.
 * Or reuse the MAC address of the phone that already works with the Mitsubishi app. If you have registered one or two phones before you need to reuse the MAC address of either of those because the car will not allow for extra MAC addresses to be registered (When reusing a phone MAC address
   you can skip the registration in step 6.)

### Step 5: Connect the Pi's Wifi to the car

The car must be switched on and the Pi must be within Wifi range. Check that your Pi uses
NetworkManager (it should on current Raspberry Pi OS):

```bash
systemctl is-active NetworkManager
```

This should print `active`. Now run the helper script with the car's Wifi name, its
password, the MAC address from step 4, the Wifi interface (`wlan0` on a Pi) and your
two-letter country code (for example `AU`, `GB` or `NL`). Keep the quotes around the password:

```bash
cd ~/phev2mqtt
sudo scripts/setup-car-wifi-nm.sh '<car_ssid>' '<car_password>' <mac_address> wlan0 <country>
```

For example `sudo scripts/setup-car-wifi-nm.sh 'REMOTE45abcd' 'mypassword123' 02:11:22:33:44:55 wlan0 AU`.

Raspberry Pi OS keeps Wifi switched off until a Wifi country is set. If you didn't enter
Wifi details in Imager, none is set, so the script sets it for you. You can leave out
`wlan0 <country>` if a country is already set.

The script creates a NetworkManager connection with the settings the car's old Wifi needs,
restarts NetworkManager and connects to the car. The car's Wifi never becomes the Pi's
default route, so the Pi keeps reaching your home network and MQTT broker over Ethernet.
To reconnect later by hand, run `sudo nmcli connection up '<car_ssid>'`.

Check that it worked. The Pi should have the address `192.168.8.47`, and the car should answer:

```bash
ip addr show wlan0
ping -c 3 192.168.8.46
```

If you can't reach the car, see [Troubleshooting](#troubleshooting).

### Step 6: Register the Pi with the car

(Skip this if you reused your phone's MAC address in step 4.)

1. Follow the [Mitsubishi instructions](https://www.mitsubishi-motors.com/en/products/outlander_phev/app/remote/)
   and put the car into registration mode ("Setup Your Vehicle"). You may need to run
   `sudo nmcli connection up '<car_ssid>'` again afterwards.
2. Run:

   ```bash
   phev2mqtt client register
   ```

   After a short time you should see a message that registration worked.

Now check that data flows. The Mitsubishi app on your phone must be disconnected from the car:

```bash
phev2mqtt client watch
```

You should see a stream of messages from the car. Press `Ctrl+C` to stop.

### Step 7: Run it automatically as a service

Create the service file. **Change `192.168.1.50` to your MQTT broker's address**, and
`<car_ssid>` (in the `ExecStart` line, all the way to the right) to the car's Wifi name. If your broker needs a login, add
`--mqtt_username <user> --mqtt_password <password>` right after the --mqtt_server address.

```bash
sudo nano /etc/systemd/system/phev2mqtt.service
```

Paste this, then save with `Ctrl+O`, `Enter`, and exit with `Ctrl+X`:

```ini
[Unit]
Description=phev2mqtt service script
After=network-online.target
Wants=network-online.target

[Service]
Type=exec
# Change the MQTT address (and add the username/password if your broker needs them).
# The wifi_restart options make it reconnect to the car's Wifi if the car hasn't been seen
# for 2 minutes. Replace <car_ssid> with the car's Wifi name.
ExecStart=/usr/local/bin/phev2mqtt --config=/dev/null client mqtt --mqtt_server tcp://192.168.1.50:1883 --wifi_restart_time=2m --wifi_restart_retry_time=1m --wifi_restart_command="nmcli device disconnect wlan0; sleep 2; nmcli connection up '<car_ssid>'" -v=debug

# Restart the program if it stops, after waiting 30 seconds
Restart=always
RestartSec=30s

# Tag the log lines. View with: sudo journalctl -f -u phev2mqtt -o cat
SyslogIdentifier=phev2mqtt
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

The service runs as root, so it can run `nmcli` without extra setup. Then start it, and
make it start at every boot:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now phev2mqtt.service
```

Watch the log (press `Ctrl+C` to stop watching; the service keeps running):

```bash
sudo journalctl -f -u phev2mqtt -o cat
```

You should see it connecting to the car and to MQTT, and then register values scrolling by.
While the car is out of Wifi range you'll see it retrying; that's normal.

Finally, **restart the Pi** (`sudo reboot`) and check that everything comes back by itself.

### Step 8: See the car in Home Assistant

With Home Assistant connected to the same MQTT broker, the car appears automatically once the Pi
has connected to the car and read its VIN. Look in **Settings -> Devices & services -> MQTT**,
or search for "phev" in your entities. See [Home Assistant discovery](#home-assistant-discovery)
for details.

### Troubleshooting

| Problem | What to try |
|---|---|
| `ping 192.168.8.46` fails | Is the car on and in range? Run `sudo nmcli connection up '<car_ssid>'`, then check `ip addr show wlan0` shows `192.168.8.47`. Check `sudo journalctl -u NetworkManager -n 50` for errors. |
| `No suitable device found for this connection (device eth0 not available … mismatching interface name)` | The message names eth0, but the real problem is that wlan0 can't be used. Usually no Wifi country is set: run the step 5 script again with your country code at the end (or `sudo raspi-config nonint do_wifi_country AU`). Also try `sudo rfkill unblock wifi` and `sudo nmcli radio wifi on`, and check `nmcli device` shows `wlan0` as `wifi` and not `unavailable`. If your Wifi interface has another name, pass it as the 4th argument. |
| `ASSOC-REJECT` in the logs | Run the step 5 script again; it sets the plain WPA2 settings the car needs. |
| Weak Wifi signal to the car (`ASSOC-REJECT status_code=16`, `ssid-not-found`, or connected but no `192.168.8.47` address and dropped after ~20 s) | Check the signal while connected with `sudo iw dev wlan0 link`; it should be -70 dBm or better (e.g. -60). If moving the Pi closer isn't possible, use a USB Wifi adapter with an external antenna: plug it in, run `nmcli device` and note its name (usually `wlan1`). Run the step 5 script again with that name as the 4th argument, e.g. `sudo scripts/setup-car-wifi-nm.sh '<car_ssid>' '<car_password>' <mac_address> wlan1 <country>`. Then change `wlan0` to the new name everywhere else: in the step 7 service file (`--wifi_restart_command`, then `sudo systemctl daemon-reload && sudo systemctl restart phev2mqtt`) and in the `ip addr show` / `ip link show` checks. To stop using the built-in Wifi, add `dtoverlay=disable-wifi` to `/boot/firmware/config.txt` and reboot. The adapter may then be called `wlan0`, so check `nmcli device` again. |
| `client register` fails | Put the car in registration mode again and run it straight away. Make sure `ip link show wlan0` shows the MAC address from step 4 (after `link/ether`). |
| `client watch` shows nothing | Disconnect the phone app from the car's Wifi. |
| The log says it can't reach MQTT | Check the broker address and login in the service file. Test with `ping <broker address>` from the Pi. |
| The car never appears in Home Assistant | Check the MQTT integration in Home Assistant uses the same broker, and look at the Pi's log for MQTT errors. |
| The Pi doesn't boot from USB | Remove any SD card, try another USB port (blue USB 3 ports are best on a Pi 4/5), and check the boot order (step 1). |

### Updating phev2mqtt

```bash
cd ~/phev2mqtt
git pull
go build
sudo install -m 755 phev2mqtt /usr/local/bin/phev2mqtt
sudo systemctl restart phev2mqtt.service
```

## Licence, etc

Licenced under the GPLv3 (or later) by the original developer:

Copyright 2021 Ben Buxton <bbuxton@gmail.com>

Modifications in this fork:

Copyright 2026 ntigch

Contributions and PRs are welcome.

## Manual setup and reference

The [Raspberry Pi quick start](#quick-start-raspberry-pi-step-by-step) is easier if you
have a Pi. These are the general steps for any Linux machine with a Wifi interface.

### Compiling

#### Install Go

 * Download and install the latest [Go compiler](https://golang.org/dl/) (Go 1.27.4 or newer is required)
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
with the car (see below). If you have connected your phone to use the Mitsubishi RemoteCtrl app you can just use the MAC address of your phone.

On a Raspberry Pi, follow the [Quick start](#quick-start-raspberry-pi-step-by-step) instead
of this section: it covers the Wifi setup, registration and auto-start in one go.
On other systems, use the steps below.

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
| phev/climate/duration | Minutes climate runs when switched on without a duration: *10*, *20* or *30* |
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
| phev/set/climate/[mode] | Set climate mode (cool/heat/windscreen/off). Payload is the duration in minutes: *10*, *20*, *30*, *on* (uses the duration chosen with `phev/set/climate/duration`, default 30), or *off* |
| phev/set/climate/mode | Set climate mode from the payload (*cool*, *heat*, *windscreen* or *off*), for the duration chosen with `phev/set/climate/duration` (default 30 minutes) |
| phev/set/climate/duration | Choose how long climate runs when switched on with *on* or via `phev/set/climate/mode`: *10*, *20* or *30*. Works while the car is offline; publish it retained to keep it across restarts (the Home Assistant select does this) |
| phev/set/climate/state | `[payload]=reset` clears "terminated" state |
| phev/connection | Change car connection state to (on/off/restart) |
| phev/settings/dump | Log the car settings received so far, then clear them |

#### Home Assistant discovery

The client supports [Home Assistant MQTT Discovery](https://www.home-assistant.io/docs/mqtt/discovery/) by default.

The entities appear once phev2mqtt has connected to the car and read its VIN. They
are published again when Home Assistant or the MQTT broker restarts. You can
search for "phev" in your entity list. Your car should also appear as a device
in the Devices tab.

The "climate state" select only knows *off*, *heat*, *cool* and *windscreen*, so it shows
*off* when the car reports *terminated* or *unknown*. The read-only "climate status" sensor
shows the full state, including *terminated* and *unknown*.

You can disable this with `--ha_discovery=false` or change the discovery prefix, the default is `--ha_discovery_prefix=homeassistant`.


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

## Older Raspberry Pi OS versions

The [Quick start](#quick-start-raspberry-pi-step-by-step) is written for the current
Raspberry Pi OS (Debian 13 "trixie"). Use the notes below only if you run an older version.

### Bookworm (Debian 12)

Bookworm also uses NetworkManager, and the Quick start already installs Go from go.dev,
so the Quick start steps work unchanged.

<details>
<summary>Bullseye and older (dhcpcd + wpa_supplicant)</summary>

Only use this on older systems that do not run NetworkManager; do not combine it with
the NetworkManager setup above. Use these two steps instead of Quick start step 5, then
continue with step 6. In step 7 drop the `--wifi_restart_*` options, since they use `nmcli`.

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
</details>