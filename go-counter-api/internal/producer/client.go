package producer

import (
	"encoding/json"
	"fmt"
	"time"
)

type IStreamConn interface {
	Write([]byte) (int, error)
}

type ProducerClient struct {
	kconn IStreamConn
}

type msg struct {
	Event string    `json:"event"` //increment or reset
	Ts    time.Time `json:",omitempty"`
}

func InitProducer(kconn IStreamConn) *ProducerClient {
	return &ProducerClient{kconn: kconn}
}

func (s *ProducerClient) WriteMsg(rawMsg []byte) (err error) {
	var m msg
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

	return
}
