package api

import (
	"github.com/pkg/errors"

	"github.com/streadway/amqp"
)

type MQClient interface {
	Publish(exchangeName, routingKey string, body []byte)
	Consume(queueName string, callback func([]byte) error)
}

type MQConn struct {
	conn        amqp.Connection
	publishChan amqp.Channel
}

func NewConn() (*MQConn, error) {
	mqConn := &MQConn{}

	defaultUrl := "amqp://guest:guest@localhost:5672/"
	url := defaultUrl
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, errors.Wrapf(err, "Dial %s", url)
	}

}

func (mq *MQConn) Publish(exchangeName, routingKey string, body []byte) {

}

func (mq *MQConn) Consume(queueName string, callback func([]byte) error) {

}
