package MQTTClient

import (
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type messageCallback func(topic string, payload []byte, qos byte)

type outgoingMessage struct {
	topic   string
	payload string
	qos     byte
	retain  bool
}

type incomingMessage struct {
	topic   string
	payload []byte
	qos     byte
}

type messageProcessor struct {
	onMsgCallback   messageCallback
	client          *mqtt.Client
	incomingMsgChan chan incomingMessage
	outgoingMsgChan chan outgoingMessage
	mtx             sync.Mutex
	publishWaitTime time.Duration
	done            <-chan interface{}
}

func newMessageProcessor(client *mqtt.Client) messageProcessor {
	msgProcessor := new(messageProcessor)
	msgProcessor.client = client
	msgProcessor.outgoingMsgChan = make(chan outgoingMessage)
	msgProcessor.publishWaitTime = 1 * time.Second
	msgProcessor.incomingMsgChan = make(chan incomingMessage)
	return *msgProcessor
}

func (mp *messageProcessor) start(wg *sync.WaitGroup, done <-chan interface{}) {
	wg.Add(2)
	mp.done = done
	go mp.publishRoutine(wg)
	go mp.onMessageRoutine(wg)
}

func (mp *messageProcessor) addOutgoingMessage(topic string, qos byte, payload string, retained bool) {
	req := outgoingMessage{
		topic:   topic,
		qos:     qos,
		retain:  retained,
		payload: payload,
	}
	select {
	case mp.outgoingMsgChan <- req:
	default:
		fmt.Println("Outgoing message channel does not exist")
	}

}

func (mp *messageProcessor) publishPendingMessages() {

}

func (mp *messageProcessor) publishRoutine(wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(mp.outgoingMsgChan)
	defer fmt.Println("Exiting Publish Go routine")
	fmt.Println("Starting Publish Go routine")
	for {
		select {
		case req, ok := <-mp.outgoingMsgChan:
			if !ok {
				fmt.Println("Outgoing message Channel closed. Exiting")
				return
			}
			mp.mtx.Lock()
			token := (*mp.client).Publish(req.topic, req.qos, req.retain, req.payload)
			if !token.WaitTimeout(mp.publishWaitTime) {
				fmt.Printf("Failed to publish topic: %s. Connected:%t\n", req.topic, (*mp.client).IsConnectionOpen())
				//TODO: add persistence in case of publish fails due to broker not active
			} else {
				fmt.Printf("Publish topic %s succeed\n", req.topic)
			}
			mp.mtx.Unlock()
		case <-mp.done:
			fmt.Println("publishRoutine Shutdown received. Exiting")
			return
		}
	}
}

func (mp *messageProcessor) setReceiveMessageCallback(callback messageCallback) {
	mp.onMsgCallback = callback
}

func (mp *messageProcessor) addIncomingMessage(message mqtt.Message) {
	req := incomingMessage{
		topic:   message.Topic(),
		qos:     message.Qos(),
		payload: message.Payload(),
	}
	select {
	case mp.incomingMsgChan <- req:
	default:
		fmt.Println("Incoming message channel does not exist")
	}
}

func (mp *messageProcessor) onMessageRoutine(wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(mp.incomingMsgChan)
	defer fmt.Println("Exiting onMessage Go routine")
	fmt.Println("Starting receiving message Go routine")
	for {
		select {
		case req, ok := <-mp.incomingMsgChan:
			if !ok {
				fmt.Println("Incoming message Channel closed. Exiting")
				return
			}
			mp.mtx.Lock()
			mp.onMsgCallback(req.topic, req.payload, req.qos)
			mp.mtx.Unlock()
		case <-mp.done:
			fmt.Println("onMessageRoutine Shutdown received. Exiting")
			return
		}
	}
}
