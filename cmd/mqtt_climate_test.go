package cmd

import (
	"bytes"
	"testing"
)

func TestParseClimate(t *testing.T) {
	tests := []struct {
		topicMode, payload string
		mode, duration     byte
		wantErr            bool
	}{
		{"heat", "10", 0x2, 0x0, false},
		{"heat", "20", 0x2, 0x1, false},
		{"heat", "30", 0x2, 0x2, false},
		{"cool", "on", 0x1, 0x2, false},
		{"windscreen", "on", 0x3, 0x2, false},
		{"heat", "off", 0x0, 0x0, false},
		{"off", "off", 0x0, 0x0, false},
		{"mode", "cool", 0x1, 0x2, false},
		{"heat", "45", 0, 0, true},
		{"bogus", "on", 0, 0, true},
	}
	for _, tc := range tests {
		mode, dur, err := parseClimate(tc.topicMode, tc.payload)
		if (err != nil) != tc.wantErr {
			t.Errorf("parseClimate(%q,%q) err=%v, wantErr=%v", tc.topicMode, tc.payload, err, tc.wantErr)
			continue
		}
		if err == nil && (mode != tc.mode || dur != tc.duration) {
			t.Errorf("parseClimate(%q,%q) = %#x,%#x want %#x,%#x", tc.topicMode, tc.payload, mode, dur, tc.mode, tc.duration)
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
