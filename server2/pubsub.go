package stserver

import (
	"errors"
	"sync"
)

type PubSubMessage struct {
	Topic   string
	Message interface{}
}

type PubSub struct {
	sync.Mutex
	topics      map[string]map[chan interface{}]struct{}
	publish     chan *PubSubMessage
	OnSubscribe func(string, chan interface{})
}

func NewPubSub() *PubSub {
	pb := PubSub{
		topics:  make(map[string]map[chan interface{}]struct{}),
		publish: make(chan *PubSubMessage),
	}
	go pb.listen()
	return &pb
}

func (p *PubSub) listen() {
	for {
		select {
		case m := <-p.publish:
			if subscribers, ok := p.topics[m.Topic]; ok {
				for subscriber, _ := range subscribers {
					subscriber <- m.Message
				}
			}
		}
	}
}

func (p *PubSub) Subscribe(topic string, subscription chan interface{}) {
	p.Lock()
	defer p.Unlock()
	if _, ok := p.topics[topic]; !ok {
		p.topics[topic] = make(map[chan interface{}]struct{})
	}
	p.topics[topic][subscription] = struct{}{}

	p.OnSubscribe(topic, subscription)
}

func (p *PubSub) Unsubscribe(subscription chan interface{}) error {
	p.Lock()
	defer p.Unlock()
	for _, topic := range p.topics {
		for subscriber, _ := range topic {
			if subscriber == subscription {
				delete(topic, subscription)
				return nil
			}
		}
	}
	return errors.New("could not delete channel, does not exist")
}

func (p *PubSub) Publish(topic string, message interface{}) {
	p.publish <- &PubSubMessage{
		Topic:   topic,
		Message: message,
	}
}
