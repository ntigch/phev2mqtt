package cmd

import (
	"bytes"
	"testing"

	"github.com/ntigch/phev2mqtt/protocol"
)

func TestParseClimate(t *testing.T) {
	tests := []struct {
		topicMode, payload string
		onDuration         byte
		mode, duration     byte
		wantErr            bool
	}{
		{"heat", "10", 0x2, 0x2, 0x0, false},
		{"heat", "20", 0x2, 0x2, 0x1, false},
		{"heat", "30", 0x2, 0x2, 0x2, false},
		{"cool", "on", 0x2, 0x1, 0x2, false},
		{"windscreen", "on", 0x2, 0x3, 0x2, false},
		{"heat", "off", 0x2, 0x0, 0x0, false},
		{"off", "off", 0x2, 0x0, 0x0, false},
		{"mode", "cool", 0x2, 0x1, 0x2, false},
		// "on" and set/climate/mode follow the chosen duration.
		{"heat", "on", 0x0, 0x2, 0x0, false},
		{"mode", "heat", 0x1, 0x2, 0x1, false},
		{"mode", "windscreen", 0x0, 0x3, 0x0, false},
		// An explicit duration overrides the chosen one.
		{"heat", "20", 0x0, 0x2, 0x1, false},
		{"heat", "45", 0x2, 0, 0, true},
		{"bogus", "on", 0x2, 0, 0, true},
		{"mode", "bogus", 0x2, 0, 0, true},
	}
	for _, tc := range tests {
		mode, dur, err := parseClimate(tc.topicMode, tc.payload, tc.onDuration)
		if (err != nil) != tc.wantErr {
			t.Errorf("parseClimate(%q,%q,%#x) err=%v, wantErr=%v", tc.topicMode, tc.payload, tc.onDuration, err, tc.wantErr)
			continue
		}
		if err == nil && (mode != tc.mode || dur != tc.duration) {
			t.Errorf("parseClimate(%q,%q,%#x) = %#x,%#x want %#x,%#x", tc.topicMode, tc.payload, tc.onDuration, mode, dur, tc.mode, tc.duration)
		}
	}
}

func TestParseClimateDuration(t *testing.T) {
	for payload, want := range map[string]byte{"10": 0x0, "20": 0x1, "30": 0x2} {
		got, err := parseClimateDuration(payload)
		if err != nil || got != want {
			t.Errorf("parseClimateDuration(%q) = %#x,%v want %#x", payload, got, err, want)
		}
		if m := climateDurationMinutes(got); m != payload {
			t.Errorf("climateDurationMinutes(%#x) = %q, want %q", got, m, payload)
		}
	}
	for _, payload := range []string{"45", "on", "off", ""} {
		if _, err := parseClimateDuration(payload); err == nil {
			t.Errorf("parseClimateDuration(%q) should fail", payload)
		}
	}
}

func TestClimatePayloadMY14(t *testing.T) {
	for _, tc := range []struct {
		mode, duration, want byte
	}{
		{0x2, 0x0, 0x02}, {0x2, 0x1, 0x12}, {0x2, 0x2, 0x22},
		{0x1, 0x2, 0x21}, {0x3, 0x1, 0x13}, {0x0, 0x0, 0x00},
	} {
		got := climatePayloadMY14(tc.mode, tc.duration)
		if len(got) != 15 || got[0] != 0 || got[1] != 0 || got[6] != tc.want {
			t.Errorf("climatePayloadMY14(%#x,%#x) = %x, want byte 6 = %#x", tc.mode, tc.duration, got, tc.want)
		}
		if !bytes.Equal(got[2:6], []byte{0xff, 0xff, 0xff, 0xff}) || !bytes.Equal(got[7:], bytes.Repeat([]byte{0xff}, 8)) {
			t.Errorf("climatePayloadMY14 padding wrong: %x", got)
		}
	}
}

func TestClimatePayloadMY18(t *testing.T) {
	if got := climatePayloadMY18(0x2, 0x1); !bytes.Equal(got, []byte{0x2, 0x2, 0x1, 0x0}) {
		t.Errorf("heat/20: got %x", got)
	}
	if got := climatePayloadMY18(0x0, 0x0); !bytes.Equal(got, []byte{0x1, 0x0, 0x0, 0x0}) {
		t.Errorf("off: got %x", got)
	}
}

func TestClimateMqttStates(t *testing.T) {
	for _, tc := range []struct {
		state protocol.PreACState
		mode  string
		want  string
		heat  string
	}{
		{protocol.PreACOff, "heat", "off", "off"},
		{protocol.PreACOffMY18, "heat", "off", "off"}, // 010000 on MY18/MY20
		{protocol.PreACOn, "heat", "heat", "on"},      // 02ce02
		{protocol.PreACTerminated, "heat", "terminated", "off"},
		{protocol.PreACState(7), "heat", "unknown", "off"},
	} {
		c := new(climate)
		c.setMode(tc.mode)
		c.setState(tc.state)
		got := c.mqttStates()
		if got["/climate/state"] != tc.want || got["/climate/heat"] != tc.heat {
			t.Errorf("state %d: got state=%q heat=%q, want %q/%q", tc.state, got["/climate/state"], got["/climate/heat"], tc.want, tc.heat)
		}
	}
}
