package amqp

import (
	"log"
	"testing"
)

var (
	testCh   *Channel
	testConn *Connection

	exchange   string
	publishKey string
	msg        = []byte("hhh+222")
)

func init() {
	url := "amqp://ivmservice:Ivscloud@123@100.85.231.246:5672/"
	con, err := Dial(url)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := con.Channel()
	if err != nil {
		log.Fatal(err)
	}

	err = ch.Confirm(false)
	if err != nil {
		log.Fatal(err)
	}

	returnCh := ch.NotifyReturn(make(chan Return))
	go handleReturn(returnCh)
	testConn = con
	testCh = ch
}

func handleReturn(returnCh chan Return) {
	for ret := range returnCh {
		log.Printf("handleReturn %#v\n", ret)
	}
}
func TestAmqpPublish(t *testing.T) {
	err := testCh.Publish(
		exchange,   // exchange类型（fanout,direct,topic,topic）,默认为direct
		publishKey, // routing key(这里用的队列名称)
		false,      // mandatory
		false,      // immediate
		Publishing{
			ContentType: "application/json", // 传输类型
			Body:        msg,                // 需要发布的消息
			Type:        "",                 // 自定义消息类型
		})

	if err != nil {
		log.Fatal(err)
	}
}

func TestAmqpPublishMandatory(t *testing.T) {
	err := testCh.Publish(
		exchange,   // exchange名字
		publishKey, // routing key
		true,       // mandatory 如果不能立刻路由到队列,会return
		false,      // immediate
		Publishing{
			ContentType: "application/json", // 传输类型
			Body:        msg,                // 需要发布的消息
			Type:        "",                 // 自定义消息类型
		})

	if err != nil {
		log.Fatal(err)
	}

	select {}
}

func TestAmqpPublishImmediate(t *testing.T) {
	err := testCh.Publish(
		exchange,   // exchange类型（fanout,direct,topic,topic）,默认为direct
		publishKey, // routing key(这里用的队列名称)
		false,      // mandatory
		true,       // immediate // 如果不能立刻路由给消费者,会return
		Publishing{
			ContentType: "application/json", // 传输类型
			Body:        msg,                // 需要发布的消息
			Type:        "",                 // 自定义消息类型
		})

	if err != nil {
		log.Fatal(err)
	}

	select {}
}
