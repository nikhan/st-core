package stserver

import "sync"

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
			//if m.Topic != nil {
			if subscribers, ok := p.topics[m.Topic]; ok {
				for subscriber, _ := range subscribers {
					subscriber <- m.Message
				}
			}
			/*} else {
				for _, topic := range p.topics {
					for subscriber, _ := range topic {
						subscriber <- m.Message
					}
				}
			}*/
			//}
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

func (p *PubSub) Unsubscribe(topic string, subscription chan interface{}) {
	p.Lock()
	defer p.Unlock()

	delete(p.topics[topic], subscription)
}

func (p *PubSub) UnsubscribeAll(subscription chan interface{}) {
	p.Lock()
	defer p.Unlock()

	for _, topic := range p.topics {
		delete(topic, subscription)
	}
}

func (p *PubSub) Publish(topic string, message interface{}) {
	p.publish <- &PubSubMessage{
		Topic:   topic,
		Message: message,
	}
}
