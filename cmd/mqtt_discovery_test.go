package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// State topics published by publishRegister (and climate.mqttStates).
var publishedStateTopics = map[string]bool{
	"/available":            true,
	"/vin":                  true,
	"/registrations":        true,
	"/ecuversion":           true,
	"/charge/charging":      true,
	"/charge/remaining":     true,
	"/charge/plug":          true,
	"/door/locked":          true,
	"/door/rear_left":       true,
	"/door/rear_right":      true,
	"/door/front_right":     true,
	"/door/driver":          true,
	"/door/front_left":      true,
	"/door/front_passenger": true,
	"/door/bonnet":          true,
	"/door/boot":            true,
	"/battery/level":        true,
	"/lights/head":          true,
	"/lights/parking":       true,
	"/lights/interior":      true,
	"/lights/hazard":        true,
	"/climate/duration":     true,
}

// Command topics handled by handleIncomingMqtt.
var handledCommandTopics = map[string]bool{
	"/connection":             true,
	"/set/parkinglights":      true,
	"/set/headlights":         true,
	"/set/cancelchargetimer":  true,
	"/set/climate/heat":       true,
	"/set/climate/cool":       true,
	"/set/climate/windscreen": true,
	"/set/climate/mode":       true,
	"/set/climate/duration":   true,
}

func TestDiscoveryTopicsExist(t *testing.T) {
	for topic := range new(climate).mqttStates() {
		publishedStateTopics[topic] = true
	}

	const prefix = "phev"
	configs := discoveryConfigs("homeassistant", "VIN123", prefix, "Phev")
	if len(configs) == 0 {
		t.Fatal("no discovery configs generated")
	}
	for topic, payload := range configs {
		if !strings.HasPrefix(topic, "homeassistant/") || !strings.Contains(topic, "VIN123") {
			t.Errorf("unexpected discovery topic %q", topic)
		}
		var cfg map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &cfg); err != nil {
			t.Errorf("%s: invalid JSON: %v", topic, err)
			continue
		}
		// A select's state must be one of its options, or Home Assistant
		// rejects it. A state topic that can carry other values (like
		// /climate/state) needs a value_template mapping them to an option.
		if opts, ok := cfg["options"].([]interface{}); ok && cfg["state_topic"] == "~/climate/state" {
			tmpl, _ := cfg["value_template"].(string)
			if tmpl == "" {
				t.Errorf("%s: select on ~/climate/state needs a value_template", topic)
			}
			for _, o := range opts {
				if !strings.Contains(tmpl, fmt.Sprintf("'%v'", o)) {
					t.Errorf("%s: value_template does not cover option %v", topic, o)
				}
			}
		}
		if cfg["~"] != prefix {
			t.Errorf("%s: ~ = %v, want %q", topic, cfg["~"], prefix)
		}
		for key, known := range map[string]map[string]bool{
			"state_topic":   publishedStateTopics,
			"command_topic": handledCommandTopics,
			"avty_t":        publishedStateTopics,
		} {
			v, ok := cfg[key].(string)
			if !ok {
				continue
			}
			if !strings.HasPrefix(v, "~/") {
				t.Errorf("%s: %s %q does not use the ~ prefix", topic, key, v)
				continue
			}
			if !known[strings.TrimPrefix(v, "~")] {
				t.Errorf("%s: %s %q is not a topic phev2mqtt uses", topic, key, v)
			}
		}
	}
}
