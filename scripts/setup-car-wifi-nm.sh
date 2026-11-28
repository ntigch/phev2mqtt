#!/bin/sh
# Create/refresh a NetworkManager connection for the car's Wifi hotspot and
# connect to it.
#
# Usage: sudo scripts/setup-car-wifi-nm.sh <SSID> <PSK> <PAIRED_MAC> [ifname] [country]
#
# - SSID/PSK:    the car's hotspot credentials (SSID looks like REMOTE45bhds)
# - PAIRED_MAC:  the Wifi MAC address the car was registered with
# - ifname:      Wifi interface (default wlan0)
# - country:     two-letter Wifi country code (e.g. AU, GB, NL). Only needed if
#                none is set yet: Raspberry Pi OS keeps Wifi blocked until it is.
#
# Settings that matter for the car's old embedded access point:
#  * plain WPA2-PSK only (the car advertises RSN/CCMP/PSK, nothing else);
#    NetworkManager's default offer of SAE/FT/PMF can get rejected
#    (ASSOC-REJECT status_code=16).
#  * hidden SSID (scan_ssid), 2.4 GHz only (faster discovery), Wifi
#    power saving off
#  * fixed MAC address, no randomisation, also while scanning
#  * unlimited autoconnect retries, otherwise NetworkManager gives up for good
#    after a few failures while the car is away.
#  * never the default route and no DNS from the car, so Ethernet stays the
#    way to the home network and the MQTT broker.
set -eu

if [ $# -lt 3 ]; then
	echo "usage: $0 <SSID> <PSK> <PAIRED_MAC> [ifname] [country]" >&2
	exit 1
fi
SSID=$1
PSK=$2
MAC=$3
IFNAME=${4:-wlan0}
COUNTRY=${5:-}

die() {
	echo "Error: $*" >&2
	exit 1
}

if [ "$(id -u)" -ne 0 ]; then
	die "run this script with sudo"
fi

command -v nmcli >/dev/null 2>&1 || die "nmcli not found; is NetworkManager installed?"

# device_field <field> prints that field (TYPE or STATE) of $IFNAME, or nothing.
device_field() {
	nmcli -t -f DEVICE,"$1" device 2>/dev/null | awk -F: -v d="$IFNAME" '$1 == d { print $2; exit }'
}

# 1. The interface must exist and be a Wifi device.
TYPE=$(device_field TYPE)
if [ "$TYPE" != "wifi" ]; then
	WIFIS=$(nmcli -t -f DEVICE,TYPE device | awk -F: '$2 == "wifi" { print $1 }')
	if [ -z "$TYPE" ]; then
		echo "Error: NetworkManager has no device called '$IFNAME'." >&2
	else
		echo "Error: '$IFNAME' is a '$TYPE' device, not Wifi." >&2
	fi
	if [ -z "$WIFIS" ]; then
		echo "No Wifi devices found. Check 'nmcli device' and 'rfkill list'." >&2
	else
		echo "Wifi devices found: $(echo "$WIFIS" | paste -sd ' ' -)" >&2
		if [ "$(echo "$WIFIS" | wc -l)" -eq 1 ]; then
			echo "Pass it as the 4th argument, e.g.: $0 '$SSID' '<PSK>' $MAC $WIFIS" >&2
		fi
	fi
	exit 1
fi

# 2. Raspberry Pi OS blocks Wifi until a Wifi country is set.
if command -v raspi-config >/dev/null 2>&1; then
	if [ -n "$COUNTRY" ]; then
		raspi-config nonint do_wifi_country "$COUNTRY" || die "could not set Wifi country '$COUNTRY'"
	fi
	CURRENT=$(raspi-config nonint get_wifi_country 2>/dev/null || true)
	if [ -z "$CURRENT" ]; then
		echo "Error: no Wifi country is set, so Raspberry Pi OS keeps Wifi switched off." >&2
		echo "Rerun with your two-letter country code as the 5th argument, e.g.:" >&2
		echo "  sudo $0 '$SSID' '<PSK>' $MAC $IFNAME AU" >&2
		echo "or set it with: sudo raspi-config nonint do_wifi_country AU" >&2
		exit 1
	fi
elif [ -n "$COUNTRY" ] && command -v iw >/dev/null 2>&1; then
	iw reg set "$COUNTRY" || true
fi

# Make sure Wifi is switched on and managed by NetworkManager.
if command -v rfkill >/dev/null 2>&1; then
	rfkill unblock wifi || true
fi
nmcli radio wifi on
if [ "$(device_field STATE)" = "unmanaged" ]; then
	nmcli device set "$IFNAME" managed yes
fi

# Don't randomise the MAC while scanning.
mkdir -p /etc/NetworkManager/conf.d
printf '[device]\nwifi.scan-rand-mac-address=no\n' > /etc/NetworkManager/conf.d/wifi-scan.conf

# 3. Replace any existing connection of the same name.
nmcli connection delete "$SSID" >/dev/null 2>&1 || true

nmcli connection add type wifi ifname "$IFNAME" con-name "$SSID" ssid "$SSID" \
	802-11-wireless.hidden yes \
	802-11-wireless.band bg \
	802-11-wireless.powersave 2 \
	802-11-wireless.cloned-mac-address "$MAC" \
	802-11-wireless-security.key-mgmt wpa-psk \
	802-11-wireless-security.psk "$PSK" \
	802-11-wireless-security.proto rsn \
	802-11-wireless-security.pairwise ccmp \
	802-11-wireless-security.group ccmp \
	802-11-wireless-security.pmf disable \
	ipv4.method auto \
	ipv4.never-default yes \
	ipv4.ignore-auto-dns yes \
	ipv6.method disabled \
	connection.autoconnect yes \
	connection.autoconnect-retries 0

# 4. Restart NetworkManager (picks up the scan setting), wait for the Wifi
# device to become usable, then connect.
systemctl restart NetworkManager

i=0
while [ $i -lt 20 ]; do
	STATE=$(device_field STATE)
	case "$STATE" in
	"" | unavailable | unmanaged) ;;
	*) break ;;
	esac
	i=$((i + 1))
	sleep 1
done

if nmcli --wait 45 connection up "$SSID" ifname "$IFNAME"; then
	echo
	echo "Connected to the car. Address on $IFNAME:"
	ip -4 -brief addr show dev "$IFNAME" || true
	exit 0
fi

echo >&2
echo "The connection '$SSID' is saved, but connecting failed." >&2
echo "Is the car switched on and within Wifi range? Diagnostics:" >&2
echo "--- nmcli device" >&2
nmcli device >&2 || true
if command -v rfkill >/dev/null 2>&1; then
	echo "--- rfkill list wifi" >&2
	rfkill list wifi >&2 || true
fi
echo "--- NetworkManager log" >&2
journalctl -u NetworkManager -n 20 --no-pager >&2 || true
echo >&2
echo "Try again later with: sudo nmcli connection up '$SSID'" >&2
exit 1
