package deviceModel

import (
	"encoding/json"
)

type Device struct {
	Uuid       string   `json:"Uuid"`
	Identifier string   `json:"Identifier"`
	Model      string   `json:"Model"`
	Type       string   `json:"Type"`
	Name       string   `json:"Name"`
	Properties Property `json:"Properties"`
}

type Property struct {
	ThermostatOn        string `json:"ThermostatOn"`
	HvacOn              string `json:"HvacOn"`
	Program             string `json:"Program"`
	OperationMode       string `json:"OperationMode"`
	AmbientTemperature  string `json:"AmbientTemperature"`
	SetpointTemperature string `json:"SetpointTemperature"`
	FanSpeed            string `json:"FanSpeed"`
	OverruleActive      string `json:"OverruleActive"`
	OverruleSetpoint    string `json:"OverruleSetpoint"`
	OverruleTime        string `json:"OverruleTime"`
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
			"ThermostatOn":        &p.ThermostatOn,
			"HvacOn":              &p.HvacOn,
			"Program":             &p.Program,
			"OperationMode":       &p.OperationMode,
			"FanSpeed":            &p.FanSpeed,
			"OverruleTime":        &p.OverruleTime,
			"OverruleActive":      &p.OverruleActive,
			"AmbientTemperature":  &p.AmbientTemperature,
			"SetpointTemperature": &p.SetpointTemperature,
			"OverruleSetpoint":    &p.OverruleSetpoint,
		} {
			if v, ok := propMap[key]; ok {
				*value = v.(string)
			}
		}
	}
	return nil
}
