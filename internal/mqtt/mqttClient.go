package mqttClient

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	helpers "nhc-hvac/internal/helpers"
	nhcModel "nhc-hvac/internal/nhc"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client
var thermostatMessage = make(chan nhcModel.Device)

func onMessageReceived(client mqtt.Client, msg mqtt.Message) {
	// Unmarshal JSON data
	var response nhcModel.Response
	err := json.Unmarshal(msg.Payload(), &response)
	if err != nil {
		helpers.DebugLog(fmt.Sprint("Error decoding JSON:", err), true)
		return
	}

	// Access the Devices array within the Params object
	devices := response.Params[0].Devices

	// Find the HVAC Thermostat device
	var hvacThermostat nhcModel.Device
	found := false
	for _, device := range devices {
		// device.Model not received in devices.status messages
		if device.Uuid == helpers.ClientConfig.HVAC_UUID { // && device.Model == "hvacthermostat" {
			hvacThermostat = device
			found = true
			break
		}
	}

	// If HVAC Thermostat device is found, extract the desired properties
	if found {
		helpers.DebugLog("HVAC Thermostat device found", true)
		v := reflect.ValueOf(hvacThermostat.Properties)
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)
			helpers.DebugLog(fmt.Sprintf("%s: %v", field.Name, value), false)
		}
		helpers.DebugLog("\n", false)
		thermostatMessage <- hvacThermostat
	} else {
		helpers.DebugLog("Other Message Received", true)
		helpers.DebugLog(string(msg.Payload())+"\n", false)
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
			helpers.DebugLog(fmt.Sprintf("Subscribed to topic: %s\n", topic), true)

		}
	}
}

func publish(client mqtt.Client, topic, message string) {
	token := client.Publish(topic, 0, false, message)
	token.Wait()
	// helpers.DebugLog(fmt.Sprintf("Published message: %s to topic: %s\n", message, topic), true)
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
