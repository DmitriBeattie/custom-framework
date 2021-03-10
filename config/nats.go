package config

import (
	"time"

	"github.com/nats-io/stan.go"
)

type Nats struct {
	Clusters    []Cluster            `json:"clusters"`
	ClusterName string               `json:"name"`
	Client      string               `json:"client"`
	Queue       map[string]QueueData `json:"queue"`
}

type QueueData struct {
	MaxInFligth     *uint16 `json:"maxInFlight,omitempty"`
	DurableName     string  `json:"durableName,omitempty"`
	StartAtSequence *uint64 `json:"startAtSequence,omitempty"`
	StartAtTime     *string `json:"startAtTime,omitempty"`
	AckWaitSeconds  *int    `json:"ackWaitSeconds"`
	IsManualAck     bool    `json:"isManualAck"`
}

func (n *Nats) InstanceKind() string {
	return "nats"
}

func (n *Nats) Url() string {
	var u string
	var div string

	for i := range n.Clusters {
		if u == "" {
			div = ""
		} else {
			div = ","
		}

		u += div + n.Clusters[i].Url.String()
	}

	return u
}

func (n *Nats) SubscriptionOptions() map[string][]stan.SubscriptionOption {
	result := make(map[string][]stan.SubscriptionOption, len(n.Queue))

	for subject, settings := range n.Queue {
		var opts []stan.SubscriptionOption

		opts = append(opts, stan.DurableName(settings.DurableName))

		if settings.MaxInFligth != nil {
			opts = append(opts, stan.MaxInflight(int(*settings.MaxInFligth)))
		}

		if settings.StartAtSequence != nil {
			opts = append(opts, stan.StartAtSequence(*settings.StartAtSequence))
		}

		if settings.StartAtTime != nil {
			if time, err := time.Parse("2006-01-02T15:04:05", *settings.StartAtTime); err == nil {
				opts = append(opts, stan.StartAtTime(time))
			}
		}

		if settings.AckWaitSeconds != nil {
			opts = append(opts, stan.AckWait(time.Second*time.Duration(*settings.AckWaitSeconds)))
		}

		if settings.IsManualAck {
			opts = append(opts, stan.SetManualAckMode())
		}

		result[subject] = opts
	}

	return result
}
