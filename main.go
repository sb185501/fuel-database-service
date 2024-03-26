package main

import (
	MQTTClient "DatabaseService/MqttClient"
	SaleStorage "DatabaseService/SaleStorage"
	MQTTMessages "DatabaseService/Types"
	"encoding/json"
	"strconv"
	"strings"

	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func defaultMQTTClientConfig() MQTTClient.MQTTOptions {
	var connInfo MQTTClient.MQTTOptions
	connInfo.BrokerIP = "localhost"
	connInfo.BrokerPort = 1883
	connInfo.CleanSession = true
	connInfo.ClientName = "MQTTClient"
	connInfo.KeepAliveSeconds = 15
	return connInfo
}

var mqttService MQTTClient.MqttInterface
var storage SaleStorage.CompletedSalesStorage

// Using a dummy structure for now
type COMPLETED_SALE struct {
	SaleNumber int64  `json:"saleNumber"`
	Data       string `json:"data"` // dummy data for now
}

func OnMQTTMessage(topic string, payload []byte, qos byte) {
	fmt.Printf("Topic %s, payload %s, qos %d\n", topic, payload, qos)

	topicTree := strings.Split(topic, "/")

	if len(topicTree) != 5 {
		fmt.Println("Not an expected topic format - %w", topic)
		return
	}

	// client/{clientId}/datastore/{pumpId}/
	clientID := topicTree[1]
	pumpNumber, _ := strconv.Atoi(topicTree[3])

	switch {
	case topicTree[0] == "client" && topicTree[2] == "datastore" && topicTree[4] == "store-completed-sale":
		var completedSaleMsg MQTTMessages.CompletedFuelSale

		err := json.Unmarshal(payload, &completedSaleMsg)
		if err != nil {
			fmt.Println("Error unmarshalling store-completed-sale topic payload: %w", err)
			PublishResponse(clientID, completedSaleMsg.Metadata.EventId, MQTTMessages.ERR_INVALID_SYNTAX)
		}
		// temporarily storing the payload as is in the database
		_, err = storage.StoreCompletedSale(int64(pumpNumber), int64(completedSaleMsg.Completion.RecordNumber), string(payload))
		if err != nil {
			fmt.Println("Failed to store completed sale: %w", err)
			PublishResponse(clientID, completedSaleMsg.Metadata.EventId, MQTTMessages.ERR_FAILED)
		}
		PublishResponse(clientID, completedSaleMsg.Metadata.EventId, MQTTMessages.ERR_NONE)
	case topicTree[0] == "client" && topicTree[2] == "datastore" && topicTree[4] == "delete-completed-sale":
		var msg MQTTMessages.DeleteCompletedSale

		err := json.Unmarshal(payload, &msg)
		if err != nil {
			fmt.Println("Error unmarshalling store-completed-sale topic payload: %w", err)
			PublishResponse(clientID, msg.Metadata.EventId, MQTTMessages.ERR_INVALID_SYNTAX)
		}

		_, err = storage.DeleteCompletedSale(int64(pumpNumber), *msg.SaleNumber)
		if err != nil {
			fmt.Println("Failed to delete completed sale: %w", err)
			PublishResponse(clientID, msg.Metadata.EventId, MQTTMessages.ERR_FAILED)
		}
		PublishResponse(clientID, msg.Metadata.EventId, MQTTMessages.ERR_NONE)
	}
}

func PublishResponse(clientId string, cause string, response MQTTMessages.ResponseCode) {
	payloadObj := MQTTMessages.FuelResponse{
		Cause:        cause,
		ResponseCode: response,
	}

	topic := fmt.Sprintf("datastore/client/%s/response", clientId)

	Publish(topic, 1, payloadObj, false)
}

func Publish(topic string, qos byte, object any, retained bool) {
	payload, err := json.Marshal(object)
	if err != nil {
		fmt.Println("Failed to marshal JSON: %w", err)
	}

	mqttService.Publish(topic, 1, string(payload), false)
}

func main() {
	fmt.Println("Database Service started")
	var wg sync.WaitGroup
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
	done := make(chan interface{})
	mqttService = MQTTClient.NewService(defaultMQTTClientConfig())
	mqttService.Connect()
	mqttService.SetOnMessageCallback(OnMQTTMessage)
	mqttService.Start(&wg, done)
	mqttService.Subscribe("client/+/datastore/#", 1)

	mqttService.Publish("datastore/online", 1, "", true)

	// create new sale storage
	var err error
	storage, err = SaleStorage.NewSaleStorage("SALE_DATABASE")
	if err != nil {
		fmt.Println("failed to initialize sale storage. Error - %w", err)
	}

	//wait for any interrupt signals
	<-interruptChan
	fmt.Println("Main - Received system interrupt - Shutting down")
	close(interruptChan)
	close(done)

	//wait for routines to finish
	wg.Wait()
	fmt.Println("Main - Exiting - Shutdown complete")
}
