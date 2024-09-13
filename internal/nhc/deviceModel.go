package deviceModel

import (
	"reflect"
	"strconv"
	"strings"
)

type Response struct {
	Params []Params `json:"Params"`
}

type Params struct {
	Devices []Device `json:"Devices"`
}
type Device struct {
	Uuid       string     `json:"Uuid"`
	Identifier string     `json:"Identifier"`
	Model      string     `json:"Model"`
	Type       string     `json:"Type"`
	Name       string     `json:"Name"`
	Properties []Property `json:"Properties"`
}

// Property struct containing all possible properties
type Property struct {
	ThermostatOn        *bool     `json:"ThermostatOn,omitempty"`
	HvacOn              *bool     `json:"HvacOn,omitempty"`
	Program             *string   `json:"Program,omitempty"`
	OperationMode       *string   `json:"OperationMode,omitempty"`
	AmbientTemperature  *float64  `json:"AmbientTemperature,omitempty"`
	SetpointTemperature *float64  `json:"SetpointTemperature,omitempty"`
	FanSpeed            *FanSpeed `json:"FanSpeed,omitempty"`
	OverruleActive      *bool     `json:"OverruleActive,omitempty"`
	OverruleSetpoint    *float64  `json:"OverruleSetpoint,omitempty"`
	OverruleTime        *string   `json:"OverruleTime,omitempty"`
	UpdatedProperties   PropertyFieldUpdated
}

type FanSpeed int

const (
	High FanSpeed = iota
	Medium
	Low
)

type PropertyFieldUpdated struct {
	ThermostatOn        bool
	HvacOn              bool
	Program             bool
	OperationMode       bool
	AmbientTemperature  bool
	SetpointTemperature bool
	FanSpeed            bool
	OverruleActive      bool
	OverruleSetpoint    bool
	OverruleTime        bool
}

func (fs FanSpeed) String() string {
	return [...]string{"High", "Medium", "Low"}[fs]
}

func (d *Device) UpdateDeviceByUuid(updateMap map[string]interface{}) bool {
	params := updateMap["Params"].([]interface{})[0].(map[string]interface{})
	devices := params["Devices"].([]interface{})

	for _, deviceData := range devices {
		deviceMap := deviceData.(map[string]interface{})
		if deviceUuid, ok := deviceMap["Uuid"].(string); ok && deviceUuid == d.Uuid {
			// Update the properties
			if props, ok := deviceMap["Properties"].([]interface{}); ok && len(props) > 0 {
				for _, propData := range props {
					if propMap, ok := propData.(map[string]interface{}); ok {
						d.Properties[0].ResetUpdatedProperties()
						d.Properties[0].UpdateDeviceProperties(propMap)
					}
				}
			}
			return true // Found the correct device
		}
	}
	return false
}

func (d *Property) ResetUpdatedProperties() {
	v := reflect.ValueOf(&d.UpdatedProperties).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Bool && field.CanSet() {
			field.SetBool(false)
		}
	}
}

func (p *Property) UpdateDeviceProperties(data map[string]interface{}) {
	for key, value := range data {
		switch key {
		case "ThermostatOn":
			if p.ThermostatOn != nil && p.ThermostatOn != parseBool(value) {
				p.ThermostatOn = parseBool(value)
				p.UpdatedProperties.ThermostatOn = true
			}
		case "HvacOn":
			if p.HvacOn != nil && p.HvacOn != parseBool(value) {
				p.HvacOn = parseBool(value)
				p.UpdatedProperties.HvacOn = true
			}
		case "Program":
			if p.Program != nil && p.Program != parseString(value) {
				p.Program = parseString(value)
				p.UpdatedProperties.Program = true
			}
		case "OperationMode":
			if p.OperationMode != nil && p.OperationMode != parseString(value) {
				p.OperationMode = parseString(value)
				p.UpdatedProperties.OperationMode = true
			}
		case "AmbientTemperature":
			if p.AmbientTemperature != nil && p.AmbientTemperature != parseFloat64(value) {
				p.AmbientTemperature = parseFloat64(value)
				p.UpdatedProperties.AmbientTemperature = true
			}
		case "SetpointTemperature":
			if p.SetpointTemperature != nil && p.SetpointTemperature != parseFloat64(value) {
				p.SetpointTemperature = parseFloat64(value)
				p.UpdatedProperties.SetpointTemperature = true
			}
		case "FanSpeed":
			if p.FanSpeed != nil && p.FanSpeed != parseFanSpeed(value) {
				p.FanSpeed = parseFanSpeed(value)
				p.UpdatedProperties.FanSpeed = true
			}
		case "OverruleActive":
			if p.OverruleActive != nil && p.OverruleActive != parseBool(value) {
				p.OverruleActive = parseBool(value)
				p.UpdatedProperties.OverruleActive = true
			}
		case "OverruleSetpoint":
			if p.OverruleSetpoint != nil && p.OverruleSetpoint != parseFloat64(value) {
				p.OverruleSetpoint = parseFloat64(value)
				p.UpdatedProperties.OverruleSetpoint = true
			}
		case "OverruleTime":
			if p.OverruleTime != nil && p.OverruleTime != parseString(value) {
				p.OverruleTime = parseString(value)
				p.UpdatedProperties.OverruleTime = true
			}
		}
	}
}

// Helper functions to parse the different types
func parseString(value interface{}) *string {
	if str, ok := value.(string); ok && str != "" {
		return &str
	}
	return nil
}

func parseBool(value interface{}) *bool {
	if str, ok := value.(string); ok {
		switch strings.ToLower(str) {
		case "true", "1", "on":
			val := true
			return &val
		case "false", "0", "off":
			val := false
			return &val
		}
	}
	return nil
}

func parseFloat64(value interface{}) *float64 {
	if f, ok := value.(float64); ok {
		return &f
	}
	if str, ok := value.(string); ok && str != "" {
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return &f
		}
	}
	return nil
}

func parseFanSpeed(value interface{}) *FanSpeed {
	if s, ok := value.(string); ok && s != "" {
		var fs FanSpeed
		switch strings.ToLower(s) {
		case "high":
			fs = High
		case "medium":
			fs = Medium
		case "low":
			fs = Low
		default:
			return nil
		}
		return &fs
	}
	return nil
}
