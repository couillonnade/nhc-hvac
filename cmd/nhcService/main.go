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

	for {
		select {
		case s := <-sigs:
			fmt.Printf("<- Received signal: %s\n", s)
			close(abort)
			os.Exit(0)

		case d, ok := <-ch:
			if ok {
				updateFanSpeed(d)
			} else {
				// channel has been closed
				helpers.DebugLog("MQTT Channel closed", true)
				return
			}

		}
	}
}

// CalculateFanSpeed determines the fan speed based on the temperature difference.
func calculateFanSpeed(currentTemp, setPoint float64, operationMode string) nhcModel.FanSpeed {
	if operationMode == "Cooling" && currentTemp > setPoint ||
		operationMode == "Heating" && currentTemp < setPoint {

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

func updateFanSpeed(HvacTh nhcModel.Device) {
	askFanUpdate := false
	forceFanUpdate := false

	// There is always and only one property in the array
	if HvacTh.Properties[0].UpdatedProperties.AmbientTemperature {
		helpers.DebugLog(fmt.Sprintf("Ambient temperature changed: %f", *HvacTh.Properties[0].AmbientTemperature), true)
		askFanUpdate = true
	}

	if HvacTh.Properties[0].UpdatedProperties.SetpointTemperature {
		helpers.DebugLog(fmt.Sprintf("Setpoint temperature changed: %f", *HvacTh.Properties[0].SetpointTemperature), true)
		forceFanUpdate = true
	}

	if HvacTh.Properties[0].UpdatedProperties.OperationMode {
		forceFanUpdate = true
	}

	if HvacTh.Properties[0].UpdatedProperties.Program {
		forceFanUpdate = true
		// TODO: program handling programs are not yet implemented
	}

	if HvacTh.Properties[0].UpdatedProperties.OverruleActive {
		helpers.DebugLog(fmt.Sprintf("Overrule active changed: %t", *HvacTh.Properties[0].OverruleActive), true)
		forceFanUpdate = true
	}

	if HvacTh.Properties[0].UpdatedProperties.OverruleSetpoint {
		helpers.DebugLog(fmt.Sprintf("Overrule setpoint changed: %f", *HvacTh.Properties[0].OverruleSetpoint), true)
		forceFanUpdate = true
	}

	if HvacTh.Properties[0].UpdatedProperties.OverruleTime {
		helpers.DebugLog(fmt.Sprintf("Overrule time changed: %s", *HvacTh.Properties[0].OverruleTime), true)
		forceFanUpdate = true
	}

	if HvacTh.Properties[0].UpdatedProperties.ThermostatOn {
		helpers.DebugLog(fmt.Sprintf("Thermostat Status changed: %t", *HvacTh.Properties[0].ThermostatOn), true)
		forceFanUpdate = true
	}

	// TODO: Create a pipe to avoid blasiting mqtt messages when too many force updated at the same time
	if (forceFanUpdate || askFanUpdate) && HvacTh.Properties[0].UpdatedProperties.ThermostatOn {
		// TODO: large hysteresis if high delta and low if small delta, or adapted to temperature change.
		if forceFanUpdate || (time.Since(lastUpdate) > time.Duration(helpers.ClientConfig.Hysteresis)*time.Minute) {
			var fanspeed nhcModel.FanSpeed
			if *HvacTh.Properties[0].OverruleActive { // OverruleActive is a string, easier to compare than tryparse
				fanspeed = calculateFanSpeed(*HvacTh.Properties[0].AmbientTemperature, *HvacTh.Properties[0].OverruleSetpoint, *HvacTh.Properties[0].OperationMode)
			} else {
				fanspeed = calculateFanSpeed(*HvacTh.Properties[0].AmbientTemperature, *HvacTh.Properties[0].SetpointTemperature, *HvacTh.Properties[0].OperationMode)
			}
			var mode string
			if askFanUpdate {
				mode = "Hysteresis reached"
			} else {
				mode = "Forced update"
			}

			helpers.DebugLog(fmt.Sprintf(mode+", setting fan speed to %s", fanspeed), true)
			mqttClient.SetFanSpeed(fanspeed)
			lastUpdate = time.Now()
		} else {
			helpers.DebugLog("Hysteresis not reached, skipping fan speed update", true)
		}
	}

}
