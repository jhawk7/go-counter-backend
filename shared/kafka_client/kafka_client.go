package kafka_client

/*
Ideally the consumer and producer would have used kafka to send messages in real time, but kafka
seemed a little too intense resource-wise for a button analytics counter on a k3s cluster. Leaving package for
future reference
*/

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
	sleep    = 15
	buffer   = 10e3
)

type IStreamConn interface {
	Write([]byte) (int, error)
	ReadBatch(int, int) *kafka.Batch
}

type KafkaClient struct {
	kconn IStreamConn
}

type Msg struct {
	Event string    `json:"event"` //increment or reset
	Ts    time.Time `json:"ts,omitempty"`
}

func InitKafkaClient(kconn IStreamConn) *KafkaClient {
	common.LogInfo("successfully established connection with Kafka")
	return &KafkaClient{kconn: kconn}
}

func (s *KafkaClient) PublishMsg(rawMsg []byte) (err error) {
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

func (s *KafkaClient) ReadMsg(ch chan<- Msg) {
	batch, b := s.configureBatch()

	for {
		n, rErr := batch.Read(b)
		if rErr != nil {
			checkBatchErr(rErr)
			batch.Close()                 //batch is no longer usable after an error is returned
			batch, b = s.configureBatch() //re-configure batch and byte array
			time.Sleep(sleep * time.Second)
			continue
		}

		common.LogInfo(fmt.Sprintf("Successfully read %v bytes from batch", n))
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

func (s *KafkaClient) configureBatch() (*kafka.Batch, []byte) {
	batch := s.kconn.ReadBatch(minBytes, maxBytes)
	b := make([]byte, buffer)
	return batch, b
}

// only log errors that are NOT EOF errors - meaning all messages were read
func checkBatchErr(rErr error) {
	if rErr == io.EOF {
		common.LogInfo("all messages read from batch")
	} else {
		err := fmt.Errorf("failed to process message batch: [error: %v]", rErr)
		common.LogError(err, false)
	}
}
