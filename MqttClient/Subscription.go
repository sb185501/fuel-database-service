package MQTTClient

import (
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type subscriptionMessage struct {
	topic    string
	qos      byte
	callback mqtt.MessageHandler
}

type subscriptionHandler struct {
	client                  *mqtt.Client
	subscriptionMsgChan     chan subscriptionMessage
	maxSubscriptionWaitTime time.Duration
	mtx                     sync.Mutex
	subscriptions           []subscriptionMessage
	done                    <-chan interface{}
}

func newSubscriptionHandler(client *mqtt.Client) subscriptionHandler {
	handler := new(subscriptionHandler)
	handler.client = client
	handler.subscriptionMsgChan = make(chan subscriptionMessage)
	handler.maxSubscriptionWaitTime = 1 * time.Second
	return *handler
}

func (sh *subscriptionHandler) start(wg *sync.WaitGroup, done <-chan interface{}) {
	wg.Add(1)
	sh.done = done
	go sh.subscriptionRoutine(wg)
}

func (sh *subscriptionHandler) addNewSubscription(topic string, qos byte, callback mqtt.MessageHandler) {
	req := subscriptionMessage{
		topic:    topic,
		qos:      qos,
		callback: callback,
	}

	select {
	case sh.subscriptionMsgChan <- req:
	default:
		fmt.Println("Subscription message channel does not exist")
	}
}

func (sh *subscriptionHandler) subscriptionRoutine(wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(sh.subscriptionMsgChan)
	defer fmt.Println("Exiting subscription Go routine")
	fmt.Println("Starting subscription Go routine")
	for {
		select {
		case req, ok := <-sh.subscriptionMsgChan:
			if !ok {
				fmt.Println("subscription message channel is closed. Exiting")
				return
			}
			sh.mtx.Lock()
			token := (*sh.client).Subscribe(req.topic, req.qos, req.callback)
			if !token.WaitTimeout(sh.maxSubscriptionWaitTime) {
				fmt.Printf("Failed to subscribe to topic: %s. Connected:%t\n", req.topic, (*sh.client).IsConnectionOpen())

			} else {
				fmt.Printf("Subcription to topic %s succeed\n", req.topic)
			}
			sh.subscriptions = append(sh.subscriptions, subscriptionMessage{req.topic, req.qos, req.callback})
			sh.mtx.Unlock()
		case <-sh.done:
			fmt.Println("SubscriptionRoutine Received shutdown. Exiting")
			return
		}
	}
}

func (sh *subscriptionHandler) reAddSubscriptions() {
	// When the connection to the broker is lost and connected back.
	// We need to reprocess all the subscriptions to broker
	for _, sub := range sh.subscriptions {
		sh.mtx.Lock()
		token := (*sh.client).Subscribe(sub.topic, sub.qos, sub.callback)
		if !token.WaitTimeout(sh.maxSubscriptionWaitTime) {
			fmt.Printf("Failed to subscribe to topic: %s. Will try on next reconnect\n", sub.topic)

		} else {
			fmt.Printf("Subcription to topic %s succeed\n", sub.topic)
		}
		sh.mtx.Unlock()
	}
}
