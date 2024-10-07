package event

import (
	"encoding/json"
	"fmt"
	"github.com/go-counter-backend/shared/common"
	"time"
)

type IStreamConn interface {
	Write([]byte) (int, error)
}

type EventClient struct {
	kconn IStreamConn
}

type msg struct {
	Event string    `json:"event"` //increment or reset
	Ts    time.Time `json:",omitempty"`
}

func InitEventClient(kconn IStreamConn) *EventClient {
	return &EventClient{kconn: kconn}
}

func (s *EventClient) WriteMsg(rawMsg []byte) (err error) {
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

	common.LogInfo(fmt.Sprintf("successfully sent message to stream client; %v", m.Event))
	return
}
