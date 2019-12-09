package certstream

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
)

type JSONReader interface {
	ReadJSON(v interface{}) error
}

type Event struct {
	AllDomains []string
	Data       json.RawMessage
}

func EventStream(ctx context.Context, c JSONReader, ch chan Event) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// no-op
		}

		msg := rawCertStreamMessage{}
		err := c.ReadJSON(&msg)
		if err != nil {
			return errors.Wrap(err, "Error decoding json frame!")
		}

		res := msg.MessageType
		switch res {
		case "heartbeat":
			continue
		case "certificate_update":
			certData := data{}
			err = json.Unmarshal(msg.Data, &certData)
			if err != nil {
				return errors.Wrap(err, "Error decoding json frame!")
			}

			ch <- Event{
				AllDomains: certData.LeafCert.AllDomains,
				Data:       msg.Data,
			}
		default:
			return errors.Errorf("unknown message type %s", res)
		}
	}
}

type rawCertStreamMessage struct {
	MessageType string          `json:"message_type"`
	Data        json.RawMessage `json:"data"`
}

type data struct {
	UpdateType string `json:"update_type"`
	LeafCert   struct {
		AllDomains []string `json:"all_domains"`
	} `json:"leaf_cert"`
}
