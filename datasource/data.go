package datasource

import (
	"fmt"

	"encoding/json"
	"sync"
	"time"

	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/natemarks/puppers"
)

func addEc2InstanceMetadata(m map[string]string) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return
	}

	client := imds.NewFromConfig(cfg)
	instanceId, err := client.GetMetadata(context.TODO(), &imds.GetMetadataInput{
		Path: "instance-id",
	})
	if err == nil {
		m["Ec2InstanceId"] = instanceId
	}

	return
}

// GetEventFromMessage returns a JSON message string
func GetEventFromMessage(message string) (event string) {
	m := make(map[string]string)
	m["Version"] = puppers.Version
	m["Message"] = message
	addEc2Instance(m)
	marshalled, err := json.Marshal(m)
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
