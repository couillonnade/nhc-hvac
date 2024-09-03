package deviceModel

import (
	"encoding/json"
	"strconv"
)

const NilFloat = -999

type Device struct {
	Uuid       string   `json:"Uuid"`
	Identifier string   `json:"Identifier"`
	Model      string   `json:"Model"`
	Type       string   `json:"Type"`
	Name       string   `json:"Name"`
	Properties Property `json:"Properties"`
}

type Property struct {
	ThermostatOn        string  `json:"ThermostatOn"`
	HvacOn              string  `json:"HvacOn"`
	Program             string  `json:"Program"`
	OperationMode       string  `json:"OperationMode"`
	AmbientTemperature  float64 `json:"AmbientTemperature"`
	SetpointTemperature float64 `json:"SetpointTemperature"`
	FanSpeed            string  `json:"FanSpeed"`
	OverruleActive      bool    `json:"OverruleActive"`
	OverruleSetpoint    float64 `json:"OverruleSetpoint"`
	OverruleTime        string  `json:"OverruleTime"`
}

type Params struct {
	Devices []Device `json:"Devices"`
}

type Response struct {
	Params []Params `json:"Params"`
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

func (p *Property) UnmarshalJSON(data []byte) error {
	var properties []map[string]interface{}
	if err := json.Unmarshal(data, &properties); err != nil {
		return err
	}

	for _, propMap := range properties {
		for key, value := range map[string]*string{
			"ThermostatOn":  &p.ThermostatOn,
			"HvacOn":        &p.HvacOn,
			"Program":       &p.Program,
			"OperationMode": &p.OperationMode,
			"FanSpeed":      &p.FanSpeed,
			"OverruleTime":  &p.OverruleTime,
		} {
			if v, ok := propMap[key]; ok {
				*value = v.(string)
			}
		}
		if ambientTemperature, ok := propMap["AmbientTemperature"]; ok {
			if floatValue, err := strconv.ParseFloat(ambientTemperature.(string), 64); err == nil {
				p.AmbientTemperature = floatValue
			} else {
				p.SetpointTemperature = NilFloat
			}
		}
		if setpointTemperature, ok := propMap["SetpointTemperature"]; ok {
			if floatValue, err := strconv.ParseFloat(setpointTemperature.(string), 64); err == nil {
				p.SetpointTemperature = floatValue
			} else {
				p.SetpointTemperature = NilFloat
			}
		}
		if overruleActive, ok := propMap["OverruleActive"]; ok {
			// bool default is false, ignore error
			if boolValue, err := strconv.ParseBool(overruleActive.(string)); err == nil {
				p.OverruleActive = boolValue
			}
		}
		if overruleSetpoint, ok := propMap["OverruleSetpoint"]; ok {
			if floatValue, err := strconv.ParseFloat(overruleSetpoint.(string), 64); err == nil {
				p.OverruleSetpoint = floatValue
			} else {
				p.SetpointTemperature = NilFloat
			}
		}

		if operationMode, ok := propMap["OverruleActive"]; ok {
			// bool default is false, ignore error
			boolValue, _ := strconv.ParseBool(operationMode.(string))
			p.OverruleActive = boolValue
		}
	}
	return nil
}
