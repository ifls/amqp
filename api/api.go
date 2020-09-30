package api

type MQConn struct {
}

func (mq *MQConn) Publish(exchangeName, routingKey string, body []byte) {

}

func (mq *MQConn) Consume(queueName string, callback func([]byte) error) {

}
