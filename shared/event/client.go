package event

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-counter-backend/shared/common"
	"github.com/segmentio/kafka-go"
)

const (
	minBytes = 10
	maxBytes = 1e6
)

type IStreamConn interface {
	Write([]byte) (int, error)
	ReadBatch(int, int) *kafka.Batch
}

type EventClient struct {
	kconn IStreamConn
}

type Msg struct {
	Event string    `json:"event"` //increment or reset
	Ts    time.Time `json:"ts,omitempty"`
}

func InitEventClient(kconn IStreamConn) *EventClient {
	return &EventClient{kconn: kconn}
}

func (s *EventClient) WriteMsg(rawMsg []byte) (err error) {
	var m Msg
	if mErr := json.Unmarshal(rawMsg, &m); mErr != nil {
		err = fmt.Errorf("failed to parse incoming message; [raw: %v], [error: %v]", rawMsg, mErr)
		return
	}

	m.Ts = time.Now()
	rawBytes, bErr := json.Marshal(m)
	if bErr != nil {
		err = fmt.Errorf("failed to convert msg to bytes; [error: %v]", bErr)
		return
	}

	if _, wErr := s.kconn.Write(rawBytes); wErr != nil {
		err = fmt.Errorf("failed to write msg to stream client; [error: %v]", wErr)
		return
	}

	common.LogInfo(fmt.Sprintf("successfully sent message to stream client; %v", m.Event))
	return
}

func (s *EventClient) ReadMsg(ch chan<- Msg) {
	batch := s.kconn.ReadBatch(minBytes, maxBytes)
	b := make([]byte, 20)
	for {
		n, rErr := batch.Read(b)
		if rErr != nil {
			err := fmt.Errorf("failed to process message batch: [error: %v]", rErr)
			common.LogError(err, false)
		}
		common.LogInfo(fmt.Sprintf("Successfully read %v messages from batch", n))
		common.LogInfo(fmt.Sprintf("read msg %v", string(b[:n])))

		var msg Msg
		if jErr := json.Unmarshal(b[:n], &msg); jErr != nil {
			err := fmt.Errorf("failed to parse raw message into struct; [error: %v]", jErr)
			common.LogError(err, false)
		}
		ch <- msg
	}
}
