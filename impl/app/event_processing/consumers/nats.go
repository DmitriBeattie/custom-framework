package consumers

import (
	"encoding/json"
	"fmt"
	"github.com/DmitriBeattie/custom-framework/interfaces/app"
	"github.com/DmitriBeattie/custom-framework/provider"
)

type queueDefinition map[string]string

type nats struct {
	conn      *provider.NATS
	def    	  queueDefinition
	name      string
}

func Nats(_conn *provider.NATS, _def map[string]string, _name string) *nats {
	return &nats{
		conn: _conn,
		def: _def,
		name: _name,
	}
}

func (n *nats) Name() string {
	return n.name
}

func consume(conn *provider.NATS, def queueDefinition, taskConfig *app.TaskConfig) (error, errCode) {
	queue, ok := def[taskConfig.Name]
	if !ok {
		return nil, ErrCodeNotFoundQueue
	}

	msgs := make([]interface{}, 0, taskConfig.Len())

	for _, val := range taskConfig.ShowEvents() {
		msgs = append(msgs, val)
	}

	msgByte, err := json.Marshal(msgs)
	if err != nil {
		return err, ErrCodeBadMessage
	}

	if err := conn.Publish(queue, msgByte); err != nil {
		return err, ErrCodeFailedToPublish
	}

	return nil, OK
}

func (n *nats) Consume(taskConfig *app.TaskConfig) error {
	err, code := consume(n.conn, n.def, taskConfig)
	if code == ErrCodeFailedToPublish {
		taskConfig.SetErrorToAll(err)
	}

	if code == ErrCodeNotFoundQueue {
		return fmt.Errorf("Not found queue for event %s in %s", taskConfig.Name, n.name)
	}

	return err
}
