package datasource

import (
	"fmt"

	"encoding/json"
	"sync"
	"time"

	"github.com/natemarks/puppers/version"
)

// LogEvent Define the structure of the generated JSON log messages
type LogEvent struct {
	Version string `json:"version"`
	Message string `json:"message"`
}

// GetEventFromMessage returns a JSON message string
func GetEventFromMessage(message string) (event string) {
	marshalled, err := json.Marshal(
		LogEvent{Version: version.Version, Message: message})
	if err != nil {
		panic(err)
	}
	return fmt.Sprint(string(marshalled))
}

// GenerateData Generates a message and appends it to the queue (arr)
func GenerateData(arr *[]string, lock *sync.Mutex) {
	for {
		lock.Lock()
		*arr = append(*arr, GetEventFromMessage("this damn message"))
		lock.Unlock()

		time.Sleep(time.Second * 2)
	}
}
