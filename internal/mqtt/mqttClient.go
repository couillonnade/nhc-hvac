package mqttClient

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"

	helpers "nhc-hvac/internal/helpers"
	nhcModel "nhc-hvac/internal/nhc"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client
var thermostatMessage = make(chan nhcModel.Device)
var NikoHvac nhcModel.Device

func onMessageReceived(client mqtt.Client, msg mqtt.Message) {
	// Unmarshal JSON data to interfacec because some somes values might be missing or empty
	var updateMap map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &updateMap); err != nil {
		helpers.DebugLog(fmt.Sprint("Error decoding JSON:", err), true)
		return
	}
	// Apply the updates to the correct device based on UUID
	if NikoHvac.UpdateDeviceByUuid(updateMap) {
		thermostatMessage <- NikoHvac
	}

}

func connect(broker, username, password string) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("go-mqtt-client")
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetTLSConfig(&tls.Config{
		InsecureSkipVerify: true, // Disable certificate verification
	})
	opts.SetDefaultPublishHandler(onMessageReceived)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		helpers.DebugLog(fmt.Sprintf("Error connecting to MQTT broker: %v\n", token.Error()), true)
		os.Exit(1)
	}

	return client
}

func subscribe(client mqtt.Client, topics []string) {
	for _, topic := range topics {
		if token := client.Subscribe(topic, 0, onMessageReceived); token.Wait() && token.Error() != nil {
			helpers.DebugLog(fmt.Sprintf("Error subscribing to topic %s: %v\n", topic, token.Error()), true)
		} else {
			helpers.DebugLog(fmt.Sprintf("Subscribed to topic: %s", topic), true)

		}
	}
}

func publish(client mqtt.Client, topic, message string) {
	token := client.Publish(topic, 0, false, message)
	token.Wait()
}

func GetHvacThData() {
	if !client.IsConnected() {
		helpers.DebugLog("MQTT Client is not connected", true)
		return
	}

	jsonMessage := `{
		"Method": "devices.list"
	}`

	publish(client, "hobby/control/devices/cmd", jsonMessage)
}

func StartListeningThermostat(abort <-chan struct{}) <-chan nhcModel.Device {

	// Initial device state, set the UUID to get the HVAC Thermostat device response messages
	NikoHvac = nhcModel.Device{
		Uuid:       helpers.ClientConfig.HVAC_UUID,
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

	client = connect(helpers.ClientConfig.Broker,
		helpers.ClientConfig.Username,
		helpers.ClientConfig.Password)

	topics := []string{"hobby/control/devices/evt",
		"hobby/control/devices/rsp",
		"hobby/control/devices/err"}
	subscribe(client, topics)

	ch := make(chan nhcModel.Device)

	go func() {
		defer close(ch)

		for {
			select {
			case <-abort:
				helpers.DebugLog("Disconnecting MQTT Client...", true)
				client.Disconnect(250)
				return

			case t := <-thermostatMessage:
				ch <- t
			}
		}
	}()

	return ch
}

func SetFanSpeed(speed nhcModel.FanSpeed) {
	if !client.IsConnected() {
		helpers.DebugLog("MQTT Client is not connected", true)
		return
	}

	jsonMessage := `{
		"Method": "devices.control",
		"Params": [{
			"Devices": [{
				"Properties": [{
					"FanSpeed": "` + speed.String() + `"
				}],
				"Uuid": "` + helpers.ClientConfig.HVAC_UUID + `"
			}]
		}]
	}`
	publish(client, "hobby/control/devices/cmd", jsonMessage)
}
