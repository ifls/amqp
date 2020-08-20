package amqp

import (
	"log"
	"testing"
)

func Init() {
	url := "amqp://ivmservice:Ivscloud@123@100.94.63.0:5672/"
	con, err := Dial(url)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := con.Channel()

	message := []byte("sss")
	err = ch.Publish(
		"mq.alarmExchange",   // exchange类型（fanout,direct,topic,topic）,默认为direct
		"mq.alarmRoutingKey", // routing key(这里用的队列名称)
		false,                // mandatory
		false,                // immediate
		Publishing{
			ContentType: "application/json", // 传输类型
			Body:        message,            // 需要发布的消息
			Type:        "",                 // 自定义消息类型
		})

	if err != nil {
		log.Fatal(err)
	}
}

func TestAmqp(t *testing.T) {
	Init()
}
