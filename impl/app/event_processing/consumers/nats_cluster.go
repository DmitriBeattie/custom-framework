package consumers

import (
	"fmt"
	"github.com/DmitriBeattie/custom-framework/interfaces/app"
	"github.com/DmitriBeattie/custom-framework/provider"
)

type natsCluster struct {
	conn []*provider.NATS
	currentActive int
	def    	  map[string]string
	name      string
	log app.Logger
}

func NewNatsCluster(def queueDefinition, name string, log app.Logger, conn ...*provider.NATS) *natsCluster {
	return &natsCluster{
		conn: conn,
		def: def,
		name: name,
		log: log,
	}
}

func (nC *natsCluster) Name() string {
	return nC.name
}

func (nC *natsCluster) Consume(taskConfig *app.TaskConfig) error {
	var err error
	var code errCode

	activeConn := nC.currentActive

	for i := 0; i < len(nC.conn); i++ {
		err, code = consume(nC.conn[activeConn], nC.def, taskConfig)
		if code == OK {
			nC.currentActive = activeConn

			break
		}

		if code == ErrCodeNotFoundQueue {
			return fmt.Errorf("Not found queue for event %s in cluster %s", taskConfig.Name, nC.name)
		}

		if code == ErrCodeFailedToPublish {
			nC.log.Error(fmt.Errorf("Err while publish to %s: %s", nC.conn[activeConn].Url(), err.Error()))
		}

		if activeConn + 1 == len(nC.conn) {
			activeConn = 0
		} else {
			activeConn += 1
		}
	}

	if code == ErrCodeFailedToPublish {
		taskConfig.SetErrorToAll(err)

		return nil
	}

	return err
}