#!/bin/sh
# Create/refresh a NetworkManager connection for the car's Wifi hotspot.
#
# Usage: sudo scripts/setup-car-wifi-nm.sh <SSID> <PSK> <PAIRED_MAC> [ifname]
#
# - SSID/PSK:    the car's hotspot credentials (SSID looks like REMOTE45bhds)
# - PAIRED_MAC:  the Wifi MAC address the car was registered with
# - ifname:      Wifi interface (default wlan0)
#
# Settings that matter for the car's old embedded access point:
#  * plain WPA2-PSK only (the car advertises RSN/CCMP/PSK, nothing else);
#    NetworkManager's default offer of SAE/FT/PMF can get rejected
#    (ASSOC-REJECT status_code=16).
#  * hidden SSID (scan_ssid)
#  * fixed MAC address, no randomisation, also while scanning
#  * unlimited autoconnect retries, otherwise NetworkManager gives up for good
#    after a few failures while the car is away.
set -eu

if [ $# -lt 3 ]; then
	echo "usage: $0 <SSID> <PSK> <PAIRED_MAC> [ifname]" >&2
	exit 1
fi
SSID=$1
PSK=$2
MAC=$3
IFNAME=${4:-wlan0}

# Don't randomise the MAC while scanning.
mkdir -p /etc/NetworkManager/conf.d
printf '[device]\nwifi.scan-rand-mac-address=no\n' > /etc/NetworkManager/conf.d/wifi-scan.conf

# Replace any existing connection of the same name.
nmcli connection delete "$SSID" >/dev/null 2>&1 || true

nmcli connection add type wifi ifname "$IFNAME" con-name "$SSID" ssid "$SSID" \
	802-11-wireless.hidden yes \
	802-11-wireless.cloned-mac-address "$MAC" \
	802-11-wireless-security.key-mgmt wpa-psk \
	802-11-wireless-security.psk "$PSK" \
	802-11-wireless-security.proto rsn \
	802-11-wireless-security.pairwise ccmp \
	802-11-wireless-security.group ccmp \
	802-11-wireless-security.pmf disable \
	connection.autoconnect yes \
	connection.autoconnect-retries 0

systemctl restart NetworkManager
echo "Done. Connect now with: nmcli connection up '$SSID'"
