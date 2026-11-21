package cmd

import "fmt"

var climateModes = map[string]byte{"off": 0x0, "OFF": 0x0, "cool": 0x1, "heat": 0x2, "windscreen": 0x3, "mode": 0x4}

// Maps the MQTT payload to a climate duration code (10, 20, 30 minutes).
// "on" defaults to the longest duration.
var climateDurations = map[string]byte{"10": 0x0, "20": 0x1, "30": 0x2, "on": 0x2, "off": 0x0}

// parseClimate turns the last topic part (set/climate/<mode>) and the
// lowercased payload into a climate mode and duration code.
func parseClimate(topicMode, payload string) (mode, duration byte, err error) {
	mode, ok := climateModes[topicMode]
	if !ok {
		return 0, 0, fmt.Errorf("Unknown climate mode: %s", topicMode)
	}
	if mode == 0x4 { // set/climate/mode -> "heat"
		mode = climateModes[payload]
		payload = "on"
	}
	if payload == "off" {
		mode = 0x0
	}
	duration, ok = climateDurations[payload]
	if mode != 0x0 && !ok {
		return 0, 0, fmt.Errorf("Unknown climate duration: %s", payload)
	}
	return mode, duration, nil
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
