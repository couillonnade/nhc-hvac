package helpers

import (
	"fmt"
	"time"
)

func DebugLog(message string, dateTime bool) {
	if dateTime {
		fmt.Println(time.Now().Local().Format(time.RFC3339) + " -> " + message)
	} else {
		fmt.Println(message)
	}
}
