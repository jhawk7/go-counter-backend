package event

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/go-counter-backend/shared/common"
	"github.com/segmentio/kafka-go"
)

const (
	minBytes = 200
	maxBytes = 1e6
	sleep    = 10
	buffer   = 10e3
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
	common.LogInfo("successfully established connection with Kafka")
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

	bcount, wErr := s.kconn.Write(rawBytes)
	if wErr != nil {
		err = fmt.Errorf("failed to write msg to stream client; [error: %v]", wErr)
		return
	}

	common.LogInfo(fmt.Sprintf("successfully sent message of %v bytes to stream client; %v", bcount, m.Event))
	return
}

func (s *EventClient) ReadMsg(ch chan<- Msg) {
	batch := s.kconn.ReadBatch(minBytes, maxBytes)
	b := make([]byte, buffer)
	for {
		n, rErr := batch.Read(b)

		//only log errors that are NOT EOF errors - meaning all messages were read
		if rErr != nil && rErr != io.EOF {
			err := fmt.Errorf("failed to process message batch: [error: %v]", rErr)
			common.LogError(err, false)
		}

		if n == 0 {
			time.Sleep(sleep * time.Second)
			continue
		}

		common.LogInfo(fmt.Sprintf("Successfully read %v messages from batch", n))
		common.LogInfo(fmt.Sprintf("read msg %v", string(b[:n])))

		var msg Msg
		if jErr := json.Unmarshal(b[:n], &msg); jErr != nil {
			err := fmt.Errorf("failed to parse raw message into struct; [error: %v]", jErr)
			common.LogError(err, false)
		}

		common.LogInfo(fmt.Sprintf("processing message with ts: %v", msg.Ts.String()))
		ch <- msg
	}
}
