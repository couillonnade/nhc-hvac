package helpers

import (
	"encoding/json"
	"io"
	"os"
)

var ClientConfig Config

type Config struct {
	Broker            string  `json:"broker"`
	Username          string  `json:"username"`
	Password          string  `json:"password"`
	HVAC_UUID         string  `json:"HVAC-TH-Uuid"`
	SmallTempDelta    float64 `json:"smallTempDelta"`
	ModerateTempDelta float64 `json:"moderateTempDelta"`
	Hysteresis        int     `json:"hysteresis"`
}

func LoadConfig(filename string) error {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &config)
	ClientConfig = config
	return err
}
