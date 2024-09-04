package deviceModel

import (
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
}

type FanSpeed int

const (
	High FanSpeed = iota
	Medium
	Low
)

func (fs FanSpeed) String() string {
	return [...]string{"High", "Medium", "Low"}[fs]
}

func (d *Device) UpdateDeviceByUuid(updateMap map[string]interface{}) {
	params := updateMap["Params"].([]interface{})[0].(map[string]interface{})
	devices := params["Devices"].([]interface{})

	for _, deviceData := range devices {
		deviceMap := deviceData.(map[string]interface{})
		if deviceUuid, ok := deviceMap["Uuid"].(string); ok && deviceUuid == d.Uuid {
			// Update the properties
			if props, ok := deviceMap["Properties"].([]interface{}); ok && len(props) > 0 {
				for _, propData := range props {
					if propMap, ok := propData.(map[string]interface{}); ok {
						d.Properties[0].UpdateDeviceProperties(propMap)
					}
				}
			}
			break // Found the correct device
		}
	}
}

func (p *Property) UpdateDeviceProperties(data map[string]interface{}) {
	for key, value := range data {
		switch key {
		case "ThermostatOn":
			p.ThermostatOn = parseBool(value)
		case "HvacOn":
			p.HvacOn = parseBool(value)
		case "Program":
			p.Program = parseString(value)
		case "OperationMode":
			p.OperationMode = parseString(value)
		case "AmbientTemperature":
			p.AmbientTemperature = parseFloat64(value)
		case "SetpointTemperature":
			p.SetpointTemperature = parseFloat64(value)
		case "FanSpeed":
			p.FanSpeed = parseFanSpeed(value)
		case "OverruleActive":
			p.OverruleActive = parseBool(value)
		case "OverruleSetpoint":
			p.OverruleSetpoint = parseFloat64(value)
		case "OverruleTime":
			p.OverruleTime = parseString(value)
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

// // Function to update the device properties based on the update map
// func (d *Device) ApplyUpdate(updateMap map[string]interface{}) {
// 	// Access "Params" key
// 	if params, ok := updateMap["Params"].([]interface{}); ok {
// 		for _, paramData := range params {
// 			if paramMap, ok := paramData.(map[string]interface{}); ok {
// 				// Access "Devices" key within Params
// 				if devices, ok := paramMap["Devices"].([]interface{}); ok {
// 					for _, deviceData := range devices {
// 						if deviceMap, ok := deviceData.(map[string]interface{}); ok {
// 							if deviceUuid, ok := deviceMap["Uuid"].(string); ok && deviceUuid == d.Uuid {
// 								// Update the properties
// 								if props, ok := deviceMap["Properties"].([]interface{}); ok && len(props) > 0 {
// 									for _, propData := range props {
// 										if propMap, ok := propData.(map[string]interface{}); ok {
// 											d.Properties[0].ApplyUpdate(propMap)
// 										}
// 									}
// 								}
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// }
