package MQTTClient

import (
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTOptions struct {
	ClientName       string `json:"clientName"`
	CleanSession     bool   `json:"cleanSession"`
	BrokerIP         string `json:"brokerIP"`
	BrokerPort       int    `json:"brokerPort"`
	KeepAliveSeconds int    `json:"keepAliveSeconds"`
}

type MqttInterface interface {
	Connect() bool
	Start(wg *sync.WaitGroup, done <-chan interface{})
	ConfigureTLS()
	Subscribe(topicFormat string, qos byte)
	Publish(topic string, qos byte, payload string, retained bool)
	SetOnMessageCallback(callback messageCallback)
}

type mqttService struct {
	client       mqtt.Client
	mtx          sync.Mutex
	messageMgr   messageProcessor
	subscription subscriptionHandler
}

func NewService(options MQTTOptions) MqttInterface {
	service := new(mqttService)
	opts := service.initMQTTOptions(options)
	service.client = mqtt.NewClient(&opts)
	service.messageMgr = newMessageProcessor(&service.client)
	service.subscription = newSubscriptionHandler(&service.client)
	return service
}

func (mc *mqttService) SetOnMessageCallback(callback messageCallback) {
	mc.messageMgr.setReceiveMessageCallback(callback)
}

func (mc *mqttService) Connect() bool {
	mc.mtx.Lock()
	defer mc.mtx.Unlock()
	connToken := mc.client.Connect()
	for attempt := 1; attempt < 4; attempt++ {
		if !connToken.WaitTimeout(5 * time.Second) {
			fmt.Printf("Failed to connect to broker. Retrying: Attempt: %d\n", attempt)
		} else {
			fmt.Println("Successfully connected to the broker")
			break
		}
	}
	return mc.client.IsConnected()
}

func (mc *mqttService) Start(wg *sync.WaitGroup, done <-chan interface{}) {
	mc.mtx.Lock()
	defer mc.mtx.Unlock()
	mc.messageMgr.start(wg, done)
	mc.subscription.start(wg, done)
	// TODO: need to look at why this sleep is needed.
	// The subscription seems to fail if this sleep is not present
	time.Sleep(1 * time.Second)
}

func (mc *mqttService) ConfigureTLS() {
	// TODO
}

func (mc *mqttService) Subscribe(topicFormat string, qos byte) {
	mc.mtx.Lock()
	defer mc.mtx.Unlock()
	mc.subscription.addNewSubscription(topicFormat, qos, mc.onMQTTMessage)
}

func (mc *mqttService) Publish(topic string, qos byte, payload string, retained bool) {
	mc.mtx.Lock()
	defer mc.mtx.Unlock()
	mc.messageMgr.addOutgoingMessage(topic, qos, payload, retained)
}

func (mc *mqttService) onMQTTMessage(client mqtt.Client, message mqtt.Message) {
	mc.mtx.Lock()
	defer mc.mtx.Unlock()
	mc.messageMgr.addIncomingMessage(message)
}

func (mc *mqttService) onConnectionLostHandler(client mqtt.Client, err error) {
	if err != nil {
		fmt.Printf("Lost Connection with MQTT Broker. Error: %s\n", err.Error())
	} else {
		fmt.Println("Lost Connection with MQTT Broker.")
	}
}

func (mc *mqttService) onConnectHandler(client mqtt.Client) {
	fmt.Println("Connection established with MQTT Broker")
	mc.messageMgr.publishPendingMessages()
	mc.subscription.reAddSubscriptions()
}

func (mc *mqttService) setEventHandlers(options *mqtt.ClientOptions) {
	options.SetConnectionLostHandler(mc.onConnectionLostHandler)
	options.SetOnConnectHandler(mc.onConnectHandler)
	options.SetReconnectingHandler(func(client mqtt.Client, options *mqtt.ClientOptions) {
		fmt.Printf("Reconnecting to MQTT broker with options %+v\n", options)
	})
}

func (mc *mqttService) initMQTTOptions(config MQTTOptions) mqtt.ClientOptions {
	options := mqtt.NewClientOptions()
	options.SetAutoReconnect(true)
	options.SetCleanSession(config.CleanSession)
	options.SetClientID(config.ClientName)
	options.SetKeepAlive(time.Duration(config.KeepAliveSeconds) * time.Second)
	options.SetConnectRetry(true)
	options.SetConnectRetryInterval(5 * time.Second)
	options.SetMaxReconnectInterval(5 * time.Second)
	options.AddBroker(fmt.Sprintf("tcp://%s:%d", config.BrokerIP, config.BrokerPort))

	mc.setEventHandlers(options)
	return *options
}
