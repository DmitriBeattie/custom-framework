package provider

import (
	"github.com/DmitriBeattie/custom-framework/interfaces/app"
	"fmt"
	"sync"
	"time"

	stan "github.com/nats-io/stan.go"
)

type ConnectionState uint8

const (
	IsNotActive ConnectionState = iota
	ConnectionFailed
	NothingToRead
	Reading
	Disconnected
)

type QueueState uint8

type Messages struct {
	sync.RWMutex
	msgs map[string]map[uint64]*stan.Msg
}

type NATS struct {
	url        string
	client     string
	cluster    string
	subSetting map[string][]stan.SubscriptionOption
	sub        []stan.Subscription
	conn       stan.Conn
	log        app.Logger
	msgs       map[string]map[uint64]*stan.Msg
	processed  map[string]map[uint64]*stan.Msg
	queueLock  map[string]*sync.RWMutex
	subError   map[string]error
	sync.RWMutex
	ConnectionState
	disconnectSign chan bool
}

func CreateNATSConnection(_url string, _client string, _cluster string, _subSetting map[string][]stan.SubscriptionOption, _log app.Logger) *NATS {
	return &NATS{
		url:            _url,
		client:         _client,
		cluster:        _cluster,
		subSetting:     _subSetting,
		log:            _log,
		msgs:           make(map[string]map[uint64]*stan.Msg),
		processed:      make(map[string]map[uint64]*stan.Msg),
		queueLock:      make(map[string]*sync.RWMutex),
		disconnectSign: make(chan bool),
		subError:       make(map[string]error),
	}
}

func (n *NATS) GetState() ConnectionState {
	n.RLock()
	defer n.RUnlock()

	return n.getState()
}

func (n *NATS) Url() string {
	return n.url
}

func (n *NATS) getState() ConnectionState {
	return n.ConnectionState
}

func (n *NATS) Open() {
	n.Lock()

	if n.ConnectionState != Disconnected && n.ConnectionState != ConnectionFailed && n.ConnectionState != IsNotActive {
		n.log.Error(fmt.Errorf("Невозможно открыть соединение. Статус: %d", n.ConnectionState))

		n.Unlock()

		return
	}

	var err error

	//log_std.Println(q.Cfg)
	n.conn, err = stan.Connect(n.cluster, n.client, stan.NatsURL(n.url), stan.SetConnectionLostHandler(n.Reconnect))

	if err != nil {
		if n.conn != nil {
			n.conn.Close()
		}

		n.log.Error(fmt.Errorf("Can't connect: %s.\nMake sure a NATS Streaming Server is running at: %s", err.Error(), n.url))

		n.ConnectionState = ConnectionFailed

		n.Unlock()

		return
	}

	for subject, settings := range n.subSetting {
		mu := sync.RWMutex{}

		n.queueLock[subject] = &mu
		n.processed[subject] = make(map[uint64]*stan.Msg)
		n.msgs[subject] = make(map[uint64]*stan.Msg)

		sub, err := n.conn.Subscribe(
			subject,
			n.HandleMessages(subject),
			settings...,
		)
		if sub != nil {
			n.sub = append(n.sub, sub)
		}
		if err != nil {
			n.log.Error(fmt.Errorf("Subscribe error %s: %s", subject, err))

			n.subError[subject] = err
		}
	}

	if len(n.sub) == 0 {
		n.log.Error("Nothing to read from nats " + n.url)

		n.ConnectionState = NothingToRead
	} else {
		n.ConnectionState = Reading
	}

	n.Unlock()

	<-n.disconnectSign
}

func (n *NATS) Reconnect(c stan.Conn, err error) {
	n.log.Error(fmt.Errorf("Reconnection to nats, because %s", err))

	n.CloseWithTimeout(10 * time.Second)
	n.Open()
}

func (n *NATS) CloseWithTimeout(d time.Duration) {
	n.Lock()
	defer n.Unlock()

	select {
	case n.disconnectSign <- true:
	case <-time.After(d):
	}

	n.close()

	n.ConnectionState = Disconnected
}

func (n *NATS) Ack(m map[uint64]*stan.Msg, subject string) {
	mu := n.queueLock[subject]
	mu.Lock()
	defer mu.Unlock()

	n.processed[subject] = m

	curMsgs := n.msgs[subject]

	for sequence := range m {
		delete(curMsgs, sequence)
	}

	n.msgs[subject] = curMsgs

	for _, msg := range m {
		msg.Ack()
	}
}

func (n *NATS) IsQueueActive(subject string) (bool, error) {
	st := n.GetState()
	if st != Reading {
		return false, nil
	}

	if err := n.subError[subject]; err != nil {
		return false, err
	}

	if _, isExists := n.queueLock[subject]; !isExists {
		return false, fmt.Errorf("Subscription %s not exists", subject)
	}

	return true, nil
}

func (n *NATS) HandleMessages(subject string) stan.MsgHandler {
	return func(m *stan.Msg) {
		mu := n.queueLock[subject]
		mu.Lock()
		defer mu.Unlock()

		processed := n.processed[subject]

		if _, found := processed[m.Sequence]; found {
			return
		}

		msg := n.msgs[subject]

		msg[m.Sequence] = m

		n.msgs[subject] = msg
	}
}

func (n *NATS) close() {
	for i := range n.sub {
		n.sub[i].Close()
	}

	n.msgs = make(map[string]map[uint64]*stan.Msg)
	n.processed = make(map[string]map[uint64]*stan.Msg)
	n.subError = make(map[string]error)
	n.queueLock = make(map[string]*sync.RWMutex)

	if n.conn != nil {
		n.conn.Close()
	}
}

func (n *NATS) GetMessages(subject string) (map[uint64]*stan.Msg, error) {
	isActive, err := n.IsQueueActive(subject)
	if err != nil {
		return nil, err
	}

	if !isActive {
		return nil, nil
	}

	mu := n.queueLock[subject]
	mu.RLock()
	defer mu.RUnlock()

	copiedMsg := make(map[uint64]*stan.Msg, len(n.msgs[subject]))

	for sequence, msg := range n.msgs[subject] {
		copiedMsg[sequence] = msg
	}

	return copiedMsg, nil
}

func (n *NATS) Publish(subject string, msg []byte) error {
	n.RLock()
	defer n.RUnlock()

	st := n.getState()

	if st < NothingToRead || st > Reading {
		return fmt.Errorf("bad state for publish: %d", st)
	}

	return n.conn.Publish(subject, msg)
}
