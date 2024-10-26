package mqtt_client

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-counter-backend/shared/common"
)

var (
	topic string
	qos   string
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

	mChan <- msg
}

func InitConn(config *common.Config) IMQTTConn {
	//set client options
	broker := config.MQTTServer
	port := config.MQTTPort
	user := config.MQTTUser
	pass := config.MQTTPass
	topic = config.MQTTTopic

	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%v:%v", broker, port))
	opts.SetClientID("mqtt-sub-client")
	opts.SetUsername(user)
	opts.SetPassword(pass)
	opts.SetKeepAlive(time.Second * 10)
	opts.OnConnect = connHandler
	opts.OnConnectionLost = lostHandler
	mqttConn := mqtt.NewClient(opts)
	if token := mqttConn.Connect(); token.Wait() && token.Error() != nil {
		err := fmt.Errorf("mqtt connection failed; %v", token.Error())
		common.LogError(err, true)
	}

	//mChan = make(chan mqtt.Message, 1)
	//c := &client{mqttConn: mqttConn}
	return mqttConn
}

func InitClient(conn IMQTTConn) *MQTTClient {
	return &MQTTClient{
		conn: conn,
	}
}

func (m *MQTTClient) PublishMsg(rawMsg []byte) {
	var m Msg

}

func (m *MQTTClient) ReadMessages() {

}
