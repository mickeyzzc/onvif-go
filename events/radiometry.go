// RadiometryAlarm semantics of the ONVIF Thermal Service: the thermal
// counterpart of the motion topics. Camera-side temperature rules
// (spot/region thresholds, fire-point detection) evaluate on the device;
// the event only carries the verdict.

package events

import (
	"strconv"
	"strings"

	"github.com/mickeyzzc/onvif-go/v2/types"
)

// RadiometryAlarm is the fire/over-temperature state carried by the Thermal
// Service's RadiometryAlarm notification. Thermal rules run on the camera;
// consumers turn the verdict into triggers, the same role MotionAlarm plays
// for visible-light motion.
type RadiometryAlarm struct {
	Active      bool
	Temperature float64 // °C when the device reports a reading; 0 otherwise
	Source      string  // Source SimpleItem value, e.g. a thermal source token
}

// radiometryTopics maps each recognized radiometry topic, normalized by
// normalizeTopic, to the Data items that may carry its boolean state, in
// preference order: the spec shape is Active; State is accepted as vendor
// variance (the same leniency the motion topics need in the field).
var radiometryTopics = map[string][]string{
	"thermal/radiometryalarm":            {"Active", "State"},
	"ruleengine/thermal/radiometryalarm": {"Active", "State"},
}

// IsRadiometryTopic reports whether topic is a radiometry topic
// ParseRadiometryAlarm understands. Matching is case-insensitive and ignores
// namespace prefixes.
func IsRadiometryTopic(topic string) bool {
	_, ok := radiometryTopics[normalizeTopic(topic)]

	return ok
}

// ParseRadiometryAlarm extracts the radiometry state from a notification.
// The spec topic is tns1:Thermal/RadiometryAlarm; RuleEngine-qualified
// vendor variants compare equal. The state is read from the Data slot only —
// "Active" (spec shape) first, "State" as vendor variance; when both are
// present Active wins. Temperature is an optional numeric Data item (°C); a
// missing or unparseable reading is 0. Source follows the same rules as
// ParseMotionAlarm.
//
// Any other topic, or a missing or unparseable state item, returns
// (RadiometryAlarm{}, false).
func ParseRadiometryAlarm(msg NotificationMessage) (RadiometryAlarm, bool) {
	keys, ok := radiometryTopics[normalizeTopic(msg.Topic)]
	if !ok {
		return RadiometryAlarm{}, false
	}

	raw, ok := firstSimpleItemValue(msg.Message.Data, keys)
	if !ok {
		return RadiometryAlarm{}, false
	}

	active, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return RadiometryAlarm{}, false
	}

	alarm := RadiometryAlarm{Active: active, Source: motionSource(msg.Message)}

	if s, ok := simpleItemValue(msg.Message.Data, "Temperature"); ok {
		if temp, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
			alarm.Temperature = temp
		}
	}

	return alarm, true
}

// firstSimpleItemValue returns the value of the first name present among
// names, in order.
func firstSimpleItemValue(items []types.SimpleItem, names []string) (string, bool) {
	for _, name := range names {
		if v, ok := simpleItemValue(items, name); ok {
			return v, true
		}
	}

	return "", false
}
