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

	var nikoHvac nhcModel.Device

	for {
		select {
		case s := <-sigs:
			fmt.Printf("<- Received signal: %s\n", s)
			close(abort)
			os.Exit(0)

		case d, ok := <-ch:
			if ok {
				updateThData(d, nikoHvac)
			} else {
				// channel has been closed
				helpers.DebugLog("MQTT Channel closed", true)
				return
			}

		}
	}
}

// CalculateFanSpeed determines the fan speed based on the temperature difference.
func calculateFanSpeed(currentTemp, setPoint float64, nikoHvac nhcModel.Device) nhcModel.FanSpeed {
	if nikoHvac.Properties.OperationMode == "Cooling" && currentTemp > setPoint ||
		nikoHvac.Properties.OperationMode == "Heating" && currentTemp < setPoint {

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

func updateThData(d nhcModel.Device, nikoHvac nhcModel.Device) {
	needFanUpdate := false
	// store a copy of the device data but check if not empty because
	// the device data is not always complete on update messages
	if d.Properties.AmbientTemperature != nhcModel.NilFloat {
		if nikoHvac.Properties.AmbientTemperature != d.Properties.AmbientTemperature {
			nikoHvac.Properties.AmbientTemperature = d.Properties.AmbientTemperature
			helpers.DebugLog(fmt.Sprintf("Ambient temperature changed: %f", d.Properties.AmbientTemperature), true)
			// ambiant has changed, calculate fan
			needFanUpdate = true
		}
	}
	if d.Properties.SetpointTemperature != nhcModel.NilFloat {
		if nikoHvac.Properties.SetpointTemperature != d.Properties.SetpointTemperature {
			nikoHvac.Properties.SetpointTemperature = d.Properties.SetpointTemperature
			helpers.DebugLog(fmt.Sprintf("Setpoint temperature changed: %f", d.Properties.SetpointTemperature), true)
			needFanUpdate = true
		}
	}
	if d.Properties.FanSpeed != "" {
		if nikoHvac.Properties.FanSpeed != d.Properties.FanSpeed {
			nikoHvac.Properties.FanSpeed = d.Properties.FanSpeed
		}
	}
	if d.Properties.OperationMode != "" {
		if nikoHvac.Properties.OperationMode != d.Properties.OperationMode {
			nikoHvac.Properties.OperationMode = d.Properties.OperationMode
		}
	}
	if d.Properties.Program != "" {
		if nikoHvac.Properties.Program != d.Properties.Program {
			nikoHvac.Properties.Program = d.Properties.Program
		}
		// TODO: program handling programs are not yet implemented
	}

	if nikoHvac.Properties.OverruleActive != d.Properties.OverruleActive {
		nikoHvac.Properties.OverruleActive = d.Properties.OverruleActive
		helpers.DebugLog(fmt.Sprintf("Overrule active changed: %t", d.Properties.OverruleActive), true)
		needFanUpdate = true
	}

	if d.Properties.OverruleSetpoint != nhcModel.NilFloat {
		if nikoHvac.Properties.OverruleSetpoint != d.Properties.OverruleSetpoint {
			nikoHvac.Properties.OverruleSetpoint = d.Properties.OverruleSetpoint
			helpers.DebugLog(fmt.Sprintf("Overrule setpoint changed: %f", d.Properties.OverruleSetpoint), true)
			needFanUpdate = true
		}
	}
	if d.Properties.OverruleTime != "" {
		if nikoHvac.Properties.OverruleTime != d.Properties.OverruleTime {
			nikoHvac.Properties.OverruleTime = d.Properties.OverruleTime
			helpers.DebugLog(fmt.Sprintf("Overrule time changed: %s", d.Properties.OverruleTime), true)
			needFanUpdate = true
		}
	}

	if d.Properties.HvacOn != "" {
		if nikoHvac.Properties.HvacOn != d.Properties.HvacOn {
			nikoHvac.Properties.HvacOn = d.Properties.HvacOn
		}
	}

	if d.Properties.ThermostatOn != "" {
		if nikoHvac.Properties.ThermostatOn != d.Properties.ThermostatOn {
			nikoHvac.Properties.ThermostatOn = d.Properties.ThermostatOn
			helpers.DebugLog(fmt.Sprintf("Thermostat Status changed: %s", d.Properties.ThermostatOn), true)
		}
	}

	if needFanUpdate && nikoHvac.Properties.ThermostatOn == "On" {
		// TODO: large hysteresis if high delta and low if small delta, or adapted to temperature change.
		if time.Since(lastUpdate) > time.Duration(helpers.ClientConfig.Hysteresis)*time.Minute {
			var fanspeed nhcModel.FanSpeed
			if nikoHvac.Properties.OverruleActive {
				fanspeed = calculateFanSpeed(nikoHvac.Properties.AmbientTemperature, nikoHvac.Properties.OverruleSetpoint, nikoHvac)
			} else {
				fanspeed = calculateFanSpeed(nikoHvac.Properties.AmbientTemperature, nikoHvac.Properties.SetpointTemperature, nikoHvac)
			}
			helpers.DebugLog(fmt.Sprintf("hysteresis reached, updating fan speed to %s", fanspeed), true)
			mqttClient.SetFanSpeed(fanspeed)
			lastUpdate = time.Now()
		} else {
			helpers.DebugLog("hysteresis not reached, skipping fan speed update", true)
		}
	}
}
