package cmd

import "fmt"

var climateModes = map[string]byte{"off": 0x0, "OFF": 0x0, "cool": 0x1, "heat": 0x2, "windscreen": 0x3, "mode": 0x4}

// Maps the MQTT payload to a climate duration code (10, 20, 30 minutes).
var climateDurations = map[string]byte{"10": 0x0, "20": 0x1, "30": 0x2, "off": 0x0}

// The duration used for "on" until one is chosen with set/climate/duration.
const defaultClimateDuration = byte(0x2)

// parseClimate turns the last topic part (set/climate/<mode>) and the
// lowercased payload into a climate mode and duration code. A payload of
// "on" (and set/climate/mode) uses onDuration.
func parseClimate(topicMode, payload string, onDuration byte) (mode, duration byte, err error) {
	mode, ok := climateModes[topicMode]
	if !ok {
		return 0, 0, fmt.Errorf("Unknown climate mode: %s", topicMode)
	}
	if mode == 0x4 { // set/climate/mode -> "heat"
		mode, ok = climateModes[payload]
		if !ok {
			return 0, 0, fmt.Errorf("Unknown climate mode: %s", payload)
		}
		payload = "on"
	}
	if payload == "off" {
		mode = 0x0
	}
	if payload == "on" {
		return mode, onDuration, nil
	}
	duration, ok = climateDurations[payload]
	if mode != 0x0 && !ok {
		return 0, 0, fmt.Errorf("Unknown climate duration: %s", payload)
	}
	return mode, duration, nil
}

// parseClimateDuration parses a set/climate/duration payload (minutes).
func parseClimateDuration(payload string) (byte, error) {
	if payload != "off" {
		if d, ok := climateDurations[payload]; ok {
			return d, nil
		}
	}
	return 0, fmt.Errorf("Unknown climate duration: %s (use 10, 20 or 30)", payload)
}

// climateDurationMinutes returns the minutes for a duration code, as used in
// the /climate/duration state topic.
func climateDurationMinutes(d byte) string {
	return fmt.Sprintf("%d", (int(d)+1)*10)
}

// climatePayloadMY14 builds the AC mode register payload for MY14 cars.
// The duration is encoded in the high nibble (10min=0x00, 20min=0x10, 30min=0x20).
// NOTE: this layout has not been confirmed on an MY14 car.
func climatePayloadMY14(mode, duration byte) []byte {
	payload := make([]byte, 15)
	for i := range payload {
		payload[i] = 0xff
	}
	payload[0] = 0x0
	payload[1] = 0x0
	payload[6] = mode | (duration << 4)
	return payload
}

// climatePayloadMY18 builds the AC mode register payload for MY18 and MY24 cars.
func climatePayloadMY18(mode, duration byte) []byte {
	state := byte(0x02)
	if mode == 0x0 {
		state = 0x1
	}
	return []byte{state, mode, duration, 0x0}
}
