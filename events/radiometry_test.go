package events

import (
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/types"
)

func TestIsRadiometryTopic(t *testing.T) {
	tests := []struct {
		topic string
		want  bool
	}{
		{"tns1:Thermal/RadiometryAlarm", true},
		{"tns1:RuleEngine/Thermal/RadiometryAlarm", true},
		{"Thermal/RadiometryAlarm", true},
		{"tns1:thermal/radiometryalarm", true},
		{"tns1:Thermal/Temperature", false},
		{"tns1:VideoSource/MotionAlarm", false},
		{"", false},
	}
	for _, tc := range tests {
		if got := IsRadiometryTopic(tc.topic); got != tc.want {
			t.Errorf("IsRadiometryTopic(%q) = %v, want %v", tc.topic, got, tc.want)
		}
	}
}

func TestParseRadiometryAlarm(t *testing.T) {
	tests := []struct {
		name   string
		msg    NotificationMessage
		want   RadiometryAlarm
		wantOK bool
	}{
		{
			name: "spec shape Active true with Temperature and Source",
			msg: motionMsg("tns1:Thermal/RadiometryAlarm",
				[]types.SimpleItem{item("Source", "ThermalSource_1")},
				[]types.SimpleItem{item("Active", "true"), item("Temperature", "87.5")}),
			want:   RadiometryAlarm{Active: true, Temperature: 87.5, Source: "ThermalSource_1"},
			wantOK: true,
		},
		{
			name: "RuleEngine-qualified vendor variant",
			msg: motionMsg("tns1:RuleEngine/Thermal/RadiometryAlarm",
				nil,
				[]types.SimpleItem{item("Active", "true")}),
			want:   RadiometryAlarm{Active: true},
			wantOK: true,
		},
		{
			name: "vendor variance State false",
			msg: motionMsg("Thermal/RadiometryAlarm",
				nil,
				[]types.SimpleItem{item("State", "false")}),
			want:   RadiometryAlarm{},
			wantOK: true,
		},
		{
			name: "Active wins when both Active and State are present",
			msg: motionMsg("tns1:Thermal/RadiometryAlarm",
				nil,
				[]types.SimpleItem{item("Active", "false"), item("State", "true")}),
			want:   RadiometryAlarm{},
			wantOK: true,
		},
		{
			name: "case-insensitive topic and 1/0 boolean forms",
			msg: motionMsg("tns1:thermal/RADIOMETRYALARM",
				nil,
				[]types.SimpleItem{item("Active", "1")}),
			want:   RadiometryAlarm{Active: true},
			wantOK: true,
		},
		{
			name: "unparseable Temperature tolerated",
			msg: motionMsg("tns1:Thermal/RadiometryAlarm",
				nil,
				[]types.SimpleItem{item("Active", "true"), item("Temperature", "n/a")}),
			want:   RadiometryAlarm{Active: true},
			wantOK: true,
		},
		{
			name: "non-radiometry topic is not our contract",
			msg: motionMsg("tns1:VideoSource/MotionAlarm",
				nil,
				[]types.SimpleItem{item("Active", "true")}),
			want:   RadiometryAlarm{},
			wantOK: false,
		},
		{
			name: "missing state item",
			msg: motionMsg("tns1:Thermal/RadiometryAlarm",
				nil,
				[]types.SimpleItem{item("Temperature", "42")}),
			want:   RadiometryAlarm{},
			wantOK: false,
		},
		{
			name: "non-boolean state",
			msg: motionMsg("tns1:Thermal/RadiometryAlarm",
				nil,
				[]types.SimpleItem{item("Active", "hot")}),
			want:   RadiometryAlarm{},
			wantOK: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseRadiometryAlarm(tc.msg)
			if ok != tc.wantOK {
				t.Fatalf("ParseRadiometryAlarm(%q) ok = %v, want %v", tc.msg.Topic, ok, tc.wantOK)
			}
			if got != tc.want {
				t.Errorf("ParseRadiometryAlarm(%q) = %+v, want %+v", tc.msg.Topic, got, tc.want)
			}
		})
	}
}
