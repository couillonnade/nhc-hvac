package mqttClient

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Message struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

func onMessageReceived(client mqtt.Client, msg mqtt.Message) {
	message := Message{
		Topic:   msg.Topic(),
		Payload: string(msg.Payload()),
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Error converting message to JSON: %v\n", err)
		return
	}

	fmt.Println(string(jsonMessage))
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
		fmt.Printf("Error connecting to MQTT broker: %v\n", token.Error())
		os.Exit(1)
	}

	return client
}

func subscribe(client mqtt.Client, topics []string) {
	for _, topic := range topics {
		if token := client.Subscribe(topic, 0, onMessageReceived); token.Wait() && token.Error() != nil {
			fmt.Printf("Error subscribing to topic %s: %v\n", topic, token.Error())
		} else {
			fmt.Printf("Subscribed to topic: %s\n", topic)
		}
	}
}

func publish(client mqtt.Client, topic, message string) {
	token := client.Publish(topic, 0, false, message)
	token.Wait()
	fmt.Printf("Published message: %s to topic: %s\n", message, topic)
}

func StartListening() {
	broker := "mqtts://192.168.123.79:8884"
	username := "hobby"
	password := "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJob2JieSIsImlhdCI6MTcyNDQzMzQ5MCwiZXhwIjoxNzU1OTAwMDAwLCJyb2xlIjpbImhvYmJ5Il0sImF1ZCI6IkZQMDAxMTJBMjMzMTlGIiwiaXNzIjoibmhjLWNvcmUiLCJqdGkiOiI0ZmU1YTUzNy1mZjVmLTRiM2UtODI3Ny0zZTJmOWJkYzJjOTEifQ.YgFVN4rgj2etXbrYMlJgIHcTOhPmc_7VdS6FQwyBba2xcwueBneESNqT-c47oEjYQRy5risKXEHvUl2pe7kmA6GicNyWAyxvaUQftRgzuBo26vDakp9vSzADIGc2QCrs6cJto38lzGmHNKew10lqx6og8N1AeIJDiJ9TSD1wrzpkAZxdyQErRBeCLUjYjO7lKnuBZYK-lF6Zs-Nkqm4CnMNqKZ9Rr9iLGIk7u8ppnWhooiiuiM2EnRXQ2cGYmXSbEOitd6Uz8Dm1_7gP3DZZ5hL6GPcjaoVIPZJHd3KDkhnSMobCpQQHvnABy_gL6LXLsRtUAiofmrJql8Jlo6x4dQ"

	client := connect(broker, username, password)
	defer client.Disconnect(250)

	topics := []string{"hobby/control/devices/evt",
		"hobby/control/devices/rsp",
		"hobby/control/devices/err"}
	subscribe(client, topics)

	// Medium / High / Low
	jsonMessage := `{
						"Method": "devices.control",
						"Params": [{
							"Devices": [{
								"Properties": [{
									"FanSpeed": "Low"
								}],
								"Uuid": "abf89e98-d48d-4c9a-87f3-afd5758768be"
							}]
						}]
					}`
	publish(client, "hobby/control/devices/cmd", jsonMessage)

	// Keep the program running to receive messages
	select {}
}
