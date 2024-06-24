package main

import (
	"testing"
)

var t_outgoing1 = WeatherData{
	Time:          "2024-06-17 19:16:31",
	Model:         "Acurite-606TX",
	Message_type:  0,
	Id:            237,
	Channel:       "A",
	Sequence_num:  0,
	Battery_ok:    1,
	Wind_avg_mi_h: 0,
	Temperature_F: 75.5,
	Humidity:      40.0,
	Mic:           "CHECKSUM",
	Station:       "home",
}

func TestMaps(t *testing.T) {
	var ok bool
	expectedKey := "home:Acurite-606TX:237:A"
	// badKey := "home:Nonexisting:20:B"
	skey := t_outgoing1.BuildSensorKey()
	if skey != expectedKey {
		t.Errorf("Expected Key(%s) is not same as"+
			" actual key (%s)", expectedKey, skey)
	}
	if _, ok = visibleSensors[skey]; !ok {
		// Sensor not in map. Add it.
		sens := t_outgoing1.GetSensorFromData() // Create Sensor record
		visibleSensors[skey] = sens             // Add it to the available sensors
	}
	t_model := visibleSensors[expectedKey].Model
	if t_model != "Acurite-606TX" {
		t.Errorf("Expected value(%s) is not same as"+
			" actual value (%s)", "Acurite-606TX", t_model)
	}
}
