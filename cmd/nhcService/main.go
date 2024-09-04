package main

import (
	//"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nhc-hvac/internal/helpers"
	mqttClient "nhc-hvac/internal/mqtt"
	nhcModel "nhc-hvac/internal/nhc"
)

var lastUpdate time.Time

func main() {
	// Catch OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	abort := make(chan struct{})

	lastUpdate = time.Time{} // Oldest time in Go

	if err := helpers.LoadConfig("config.json"); err != nil {
		panic("Error loading config: " + err.Error())
	}

	ch := mqttClient.StartListeningThermostat(abort)
	mqttClient.GetHvacThData()

	nikoHvac := nhcModel.Device{
		Uuid:       "",
		Identifier: "",
		Model:      "",
		Type:       "",
		Name:       "",
		Properties: []nhcModel.Property{
			{
				ThermostatOn:        new(bool),
				HvacOn:              new(bool),
				Program:             new(string),
				OperationMode:       new(string),
				AmbientTemperature:  new(float64),
				SetpointTemperature: new(float64),
				FanSpeed:            new(nhcModel.FanSpeed),
				OverruleActive:      new(bool),
				OverruleSetpoint:    new(float64),
				OverruleTime:        new(string),
			},
		},
	}

	for {
		select {
		case s := <-sigs:
			fmt.Printf("<- Received signal: %s\n", s)
			close(abort)
			os.Exit(0)

		case d, ok := <-ch:
			if ok {
				updateFanSpeed(d, &nikoHvac)
			} else {
				// channel has been closed
				helpers.DebugLog("MQTT Channel closed", true)
				return
			}

		}
	}
}

// CalculateFanSpeed determines the fan speed based on the temperature difference.
func calculateFanSpeed(currentTemp, setPoint float64, current nhcModel.Device) nhcModel.FanSpeed {
	if *current.Properties[0].OperationMode == "Cooling" && currentTemp > setPoint ||
		*current.Properties[0].OperationMode == "Heating" && currentTemp < setPoint {

		diff := math.Abs(setPoint - currentTemp)

		switch {
		case diff <= helpers.ClientConfig.SmallTempDelta:
			return nhcModel.Low
		case diff <= helpers.ClientConfig.ModerateTempDelta:
			return nhcModel.Medium
		default:
			return nhcModel.High
		}

	} else {
		// TODO: check if fan should be off
		helpers.DebugLog("Fan should set to low because setpoint overshoot", true)
		return nhcModel.Low
	}
}

func updateFanSpeed(new nhcModel.Device, current *nhcModel.Device) {
	askFanUpdate := false
	forceFanUpdate := false

	// There is always and only one property in the array
	// TODO: move the comparaison to mqttClient and parse list of updates here only.
	if *current.Properties[0].AmbientTemperature != *new.Properties[0].AmbientTemperature {
		*current.Properties[0].AmbientTemperature = *new.Properties[0].AmbientTemperature
		helpers.DebugLog(fmt.Sprintf("Ambient temperature changed: %f", *new.Properties[0].AmbientTemperature), true)
		askFanUpdate = true
	}

	if *current.Properties[0].SetpointTemperature != *new.Properties[0].SetpointTemperature {
		*current.Properties[0].SetpointTemperature = *new.Properties[0].SetpointTemperature
		helpers.DebugLog(fmt.Sprintf("Setpoint temperature changed: %f", *new.Properties[0].SetpointTemperature), true)
		forceFanUpdate = true
	}

	if *current.Properties[0].FanSpeed != *new.Properties[0].FanSpeed {
		*current.Properties[0].FanSpeed = *new.Properties[0].FanSpeed
	}

	if *current.Properties[0].OperationMode != *new.Properties[0].OperationMode {
		*current.Properties[0].OperationMode = *new.Properties[0].OperationMode
		forceFanUpdate = true
	}

	if *current.Properties[0].Program != *new.Properties[0].Program {
		*current.Properties[0].Program = *new.Properties[0].Program
		forceFanUpdate = true
		// TODO: program handling programs are not yet implemented
	}

	if *current.Properties[0].OverruleActive != *new.Properties[0].OverruleActive {
		*current.Properties[0].OverruleActive = *new.Properties[0].OverruleActive
		helpers.DebugLog(fmt.Sprintf("Overrule active changed: %t", *new.Properties[0].OverruleActive), true)
		forceFanUpdate = true
	}

	if *current.Properties[0].OverruleSetpoint != *new.Properties[0].OverruleSetpoint {
		*current.Properties[0].OverruleSetpoint = *new.Properties[0].OverruleSetpoint
		helpers.DebugLog(fmt.Sprintf("Overrule setpoint changed: %f", *new.Properties[0].OverruleSetpoint), true)
		forceFanUpdate = true
	}

	if *current.Properties[0].OverruleTime != *new.Properties[0].OverruleTime {
		*current.Properties[0].OverruleTime = *new.Properties[0].OverruleTime
		helpers.DebugLog(fmt.Sprintf("Overrule time changed: %s", *new.Properties[0].OverruleTime), true)
		forceFanUpdate = true
	}

	if *current.Properties[0].HvacOn != *new.Properties[0].HvacOn {
		*current.Properties[0].HvacOn = *new.Properties[0].HvacOn
	}

	if *current.Properties[0].ThermostatOn != *new.Properties[0].ThermostatOn {
		*current.Properties[0].ThermostatOn = *new.Properties[0].ThermostatOn
		helpers.DebugLog(fmt.Sprintf("Thermostat Status changed: %t", *new.Properties[0].ThermostatOn), true)
		forceFanUpdate = true
	}

	// TODO: Create a pipe to avoid blasiting mqtt messages when too many force updated at the same time
	if (forceFanUpdate || askFanUpdate) && *current.Properties[0].ThermostatOn {
		// TODO: large hysteresis if high delta and low if small delta, or adapted to temperature change.
		if forceFanUpdate || (time.Since(lastUpdate) > time.Duration(helpers.ClientConfig.Hysteresis)*time.Minute) {
			var fanspeed nhcModel.FanSpeed
			if *current.Properties[0].OverruleActive { // OverruleActive is a string, easier to compare than tryparse
				fanspeed = calculateFanSpeed(*current.Properties[0].AmbientTemperature, *current.Properties[0].OverruleSetpoint, *current)
			} else {
				fanspeed = calculateFanSpeed(*current.Properties[0].AmbientTemperature, *current.Properties[0].SetpointTemperature, *current)
			}
			helpers.DebugLog(fmt.Sprintf("hysteresis reached (or forced), updating fan speed to %s", fanspeed), true)
			mqttClient.SetFanSpeed(fanspeed)
			lastUpdate = time.Now()
		} else {
			helpers.DebugLog("hysteresis not reached, skipping fan speed update", true)
		}
	}

}
