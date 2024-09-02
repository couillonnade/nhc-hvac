package main

import (
	//"context"
	//"fmt"
	//"os"
	//"os/signal"
	//"syscall"

	mqttClient "nhc-hvac/internal/mqtt"
)

func main() {
	mqttClient.StartListening()
}
