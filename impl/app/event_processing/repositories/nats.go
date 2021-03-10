package repositories

import (
	"errors"
	"fmt"
	"delivery-report/framework/interfaces/app"
	"delivery-report/framework/provider"
	"github.com/nats-io/stan.go"
	"strconv"
	"strings"
	"sync"
)

type natsRepo struct {
	conn *provider.NATS
	q map[string]string
}

func NatsRepository(n *provider.NATS, q map[string]string) *natsRepo {
	return &natsRepo{n, q}
}

func (n *natsRepo) GetNew(conf *app.TaskConfig, consumerName string) error {
	queue, ok := n.q[conf.Name]
	if !ok {
		return fmt.Errorf("Not found queue in nats for event %s", conf.Name)
	}

	indexedMsgs, err := n.conn.GetMessages(queue)
	if err != nil {
		return err
	}

	conf.AllocateMemForEvents(len(indexedMsgs))

	for id, msg := range indexedMsgs {
		conf.WriteEvent(id, msg.Data)
	}

	return nil
}

func (n *natsRepo) ConfirmAck(conf *app.TaskConfig, consumerName string) error {
	queue, ok := n.q[conf.Name]
	if !ok {
		return fmt.Errorf("Not found queue in nats for event %s", conf.Name)
	}

	acknowledgedMsgs := make(map[uint64]*stan.Msg, conf.Len())

	msgs, err := n.conn.GetMessages(queue)
	if err != nil {
		return err
	}

	for id, err := range conf.ShowConsumigResult() {
		if err != nil {
			continue
		}

		acknowledgedMsgs[id.(uint64)] = msgs[id.(uint64)]
	}

	n.conn.Ack(acknowledgedMsgs, queue)

	return nil
}

type natsCluster struct {
	conn []*provider.NATS
	q map[string]string
	log app.Logger
}

func msgID(clusterID int, sequenceID uint64) string {
	return fmt.Sprintf("%d;%d", clusterID, sequenceID)
}

func parseMsgID(msgID string) (int, uint64, error) {
	ind := strings.Index(msgID, ";")
	if ind <= 0 {
		return 0, 0, errors.New("Not an msgID")
	}

	clusterID, err := strconv.ParseInt(msgID[:ind], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	sequenceID, err := strconv.ParseUint(msgID[ind+1:], 10, 64)

	return int(clusterID), sequenceID, err
}

func (n *natsCluster) GetNew(conf *app.TaskConfig, consumerName string) error {
	queue, ok := n.q[conf.Name]
	if !ok {
		return fmt.Errorf("Not found queue in nats for event %s", conf.Name)
	}

	msgChan := make(chan map[string]*stan.Msg, len(n.conn))
	errChan := make(chan error, len(n.conn))

	parallelReaderWork := sync.WaitGroup{}
	parallelReaderWork.Add(len(n.conn))

	for i := range n.conn {
		go func(connInd int) {
			defer parallelReaderWork.Done()

			indexedMsg, err := n.conn[connInd].GetMessages(queue)
			if err != nil {
				errChan <- fmt.Errorf("Reading from %s: %s", n.conn[connInd].Url(), err.Error())

				return
			}

			resMsg := make(map[string]*stan.Msg, len(indexedMsg))

			for seqID, msg := range indexedMsg {
				resMsg[msgID(connInd, seqID)] = msg
			}

			msgChan <- resMsg
		}(i)
	}

	parallelReaderWork.Wait()
	close(errChan)
	close(msgChan)

	msgFromAllConn := make(map[string]*stan.Msg)

	for msg := range msgChan {
		for id, m := range msg {
			msgFromAllConn[id] = m
		}
	}

	var lastErr error
	for err := range errChan {
		if err != nil {
			n.log.Error(err)

			lastErr = err
		}
	}

	if lastErr != nil && len(msgFromAllConn) == 0 {
		return lastErr
	}

	conf.AllocateMemForEvents(len(msgFromAllConn))

	for mID, msg := range msgFromAllConn {
		conf.WriteEvent(mID, msg.Data)
	}

	return nil
}

func (n *natsCluster) ConfirmAck(conf *app.TaskConfig, consumerName string) error {
	queue, ok := n.q[conf.Name]
	if !ok {
		return fmt.Errorf("Not found queue in nats for event %s", conf.Name)
	}

	ackResByConn := make(map[int][]uint64)

	for id, err := range conf.ShowConsumigResult() {
		if err != nil {
			continue
		}

		connID, sequenceID, err := parseMsgID(id.(string))
		if err != nil {
			n.log.Error(fmt.Errorf("Unable to parse msg %s", id.(string)))

			continue
		}

		ackResByConn[connID] = append(ackResByConn[connID], sequenceID)
	}

	if len(ackResByConn) == 0 {
		return nil
	}

	parallelAck := sync.WaitGroup{}
	parallelAck.Add(len(ackResByConn))

	for srvID, seqIDs := range ackResByConn {
		go func(serverID int, sequences []uint64) {
			defer parallelAck.Done()

			rawMsg, err := n.conn[serverID].GetMessages(queue)
			if err != nil {
				n.log.Error(fmt.Errorf("Reading from %s: %s", n.conn[serverID].Url(), err.Error()))

				return
			}

			ackedMsg := make(map[uint64]*stan.Msg, len(sequences))

			for i := range sequences {
				if msg, ok := rawMsg[sequences[i]]; ok {
					ackedMsg[sequences[i]] = msg
				} else {
					n.log.Error(fmt.Errorf("Not found msg with id %d in server %d", sequences[i]))
				}
			}

			if len(ackedMsg) == 0 {
				return
			}

			n.conn[serverID].Ack(ackedMsg, queue)
		}(srvID, seqIDs)
	}

	parallelAck.Wait()

	return nil
}

func NewNatsCluster(q map[string]string, log app.Logger, conn ...*provider.NATS) *natsCluster {
	return &natsCluster{
		conn: conn,
		q:    q,
		log: log,
	}
}