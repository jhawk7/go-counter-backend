package mqtt_client

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-counter-backend/shared/common"
)

var (
	topic   string
	qos     byte // 0 (at most once), 1 (atleast once), 2 (only once)
	msgChan chan []byte
)

type IMQTTConn interface {
	Publish(string, byte, bool, interface{}) mqtt.Token
	Subscribe(string, byte, mqtt.MessageHandler) mqtt.Token
}

type MQTTClient struct {
	conn IMQTTConn
}

type Msg struct {
	Event string    `json:"event"` //increment or reset
	Ts    time.Time `json:"ts,omitempty"`
}

var connHandler mqtt.OnConnectHandler = func(mclient mqtt.Client) {
	common.LogInfo("successfully connected to mqtt server")
}

var lostHandler mqtt.ConnectionLostHandler = func(mclient mqtt.Client, err error) {
	connErr := fmt.Errorf("lost connection to mqtt server; %v", err)
	common.LogError(connErr, false)
}

var msgHandler mqtt.MessageHandler = func(mclient mqtt.Client, msg mqtt.Message) {
	info := fmt.Sprintf("messaged received [id: %d] [topic: %v], [payload: %s]", msg.MessageID(), msg.Topic(), msg.Payload())
	common.LogInfo(info)

	msgChan <- msg.Payload()
}

func InitConn(config *common.Config, clientId string) (IMQTTConn, error) {
	//set client options
	broker := config.MQTTServer
	port := config.MQTTPort
	user := config.MQTTUser
	pass := config.MQTTPass
	topic = config.MQTTTopic
	i, _ := strconv.Atoi(config.MQTTQos)
	qos = byte(i)

	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%v:%v", broker, port))
	opts.SetClientID(clientId)
	opts.SetCleanSession(false) //disabling so that messages are resumed on reconnect
	opts.SetUsername(user)
	opts.SetPassword(pass)
	opts.SetKeepAlive(time.Second * 10)
	opts.OnConnect = connHandler
	opts.OnConnectionLost = lostHandler
	mqttConn := mqtt.NewClient(opts)
	if token := mqttConn.Connect(); token.Wait() && token.Error() != nil {
		err := fmt.Errorf("mqtt connection failed; %v", token.Error())
		return nil, err
	}

	return mqttConn, nil
}

func InitClient(conn IMQTTConn) *MQTTClient {
	return &MQTTClient{
		conn: conn,
	}
}

func (m *MQTTClient) PublishMsg(rawMsg []byte) (err error) {
	var msg Msg
	if mErr := json.Unmarshal(rawMsg, &msg); mErr != nil {
		err = fmt.Errorf("failed to parse incoming message; [raw: %v], [error: %v]", rawMsg, mErr)
		return
	}

	msg.Ts = time.Now()
	msgBytes, bErr := json.Marshal(msg)
	if bErr != nil {
		err = fmt.Errorf("failed to convert msg to bytes; [error: %v]", bErr)
		return
	}

	if token := m.conn.Publish(topic, qos, true, msgBytes); token.Wait() && token.Error() != nil {
		err = fmt.Errorf("failed to publish msg to mqtt; [error: %v]", token.Error())
		return
	}

	common.LogInfo(fmt.Sprintf("successfully sent message of %v bytes to stream client; %v", len(msgBytes), msg.Event))
	return
}

func (m *MQTTClient) ReadMsg(ch chan<- Msg) {
	if token := m.conn.Subscribe(topic, qos, msgHandler); token.Wait() && token.Error() != nil {
		common.LogError(fmt.Errorf("failed to subscribe to mqtt topics; %v", token.Error()), true)
	}
	msgChan = make(chan []byte, 5) //populated by msgHandler

	for {
		for msgBytes := range msgChan {
			var msg Msg
			if jErr := json.Unmarshal(msgBytes, &msg); jErr != nil {
				err := fmt.Errorf("failed to parse raw mqtt message into struct; [error: %v] [raw: %v]", jErr, string(msgBytes))
				common.LogError(err, false)
				continue
			}

			common.LogInfo(fmt.Sprintf("processing message %v event [message ts: %v]", msg.Event, msg.Ts.String()))
			ch <- msg
		}
	}
}
