// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

/*
Package amqp is an AMQP 0.9.1 client with RabbitMQ extensions

Understand the AMQP 0.9.1 messaging model by reviewing these links first.
Much of the terminology术语 in this library directly relates to AMQP concepts.

  Resources

  TODO http://www.rabbitmq.com/tutorials/amqp-concepts.html
  TODO http://www.rabbitmq.com/getstarted.html
  TODO http://www.rabbitmq.com/amqp-0-9-1-reference.html

Design

Most other broker clients publish to queues, but in AMQP, clients publish Exchanges instead. 其他都是发队列, amqp 是发交换机
AMQP is programmable, meaning that both the producers and consumers agree on the configuration of the broker, 生产者和消费者约定配置
instead of requiring an operator操作员 or system configuration系统配置 that declares the logical topology 逻辑结构 in the broker.

The routing between producers and consumer queues is via Bindings. 路由由绑定决定
These bindings form the logical topology of the broker. 这些绑定组成了代理间的逻辑拓扑

In this library, a message sent from publisher is called a "Publishing"发布 and a message received to a consumer is
called a "Delivery"递送.

The fields of Publishings and Deliveries are close接近 but not exact mappings to the underlying wire format 传输格式 to
maintain stronger types.

Many other libraries will combine message properties with message headers. 区分共有属性和header里的用户自定义属性
In this library, the message well known properties are strongly typed fields on the Publishings and Deliveries,
whereas the user defined headers are in the Headers field.

The method naming closely matches the protocol's method name with positional parameters mapping to named protocol message fields.
The motivation here is to present a comprehensive全面 view over all possible interactions with the server.

Generally, methods that map to protocol methods of the "basic" class will be elided省略 in this interface,
and "select" methods of various channel mode selectors will be elided 例如for example Channel.Confirm and Channel.Tx.

The library is intentionally designed to be synchronous同步, where responses for
each protocol message are required to be received in an RPC manner.

Some methods have a noWait parameter like Channel.QueueDeclare, and some methods are
asynchronous like Channel.Publish. 一些方法可选同步还是异步, 一些是异步方法

要检查错误The error values should still be checked for
these methods as they will indicate IO failures like when the underlying
connection closes.

Asynchronous Events

Clients of this library may be interested in receiving some of the protocol
messages other than Deliveries like basic.ack methods while a channel is in
confirm mode. 想接收除消息外的一些其他通知

通过Noticy开头的方法, 接收
The Notify* methods with Connection and Channel receivers model the pattern of
asynchronous events like closes due to exceptions, or messages that are sent out
of band from an RPC call like basic.ack or basic.flow.

Any asynchronous events, including Deliveries and Publishings must always have
a receiver until the corresponding chans are closed.  Without asynchronous
receivers, the sychronous methods will block.

Use Case

It's important as a client to an AMQP topology to ensure the state of the broker matches your expectations.
For both publish and consume use cases,
make sure you declare the queues, exchanges and bindings you expect to exist
prior先于 to calling Channel.Publish or Channel.Consume.

  // Connections start with amqp.Dial() typically from a command line argument
  // or environment variable.
  connection, err := amqp.Dial(os.Getenv("AMQP_URL"))

  // To cleanly shutdown by flushing kernel buffers, make sure to close and
  // wait for the response.
  defer connection.Close()

  // Most operations happen on a channel.
  // If any error is returned on a channel, the channel will no longer be valid, throw it away and try with
  // a different channel.  channel出错无法再使用, 必须换一个
  // If you use many channels, it's useful for the server to
  channel, err := connection.Channel()

  // Declare your topology here, if it doesn't exist, it will be created, if
  // it existed already and is not what you expect, then that's considered an
  // error. 不一样, 会报错

  // Use your connection on this topology with either Publish or Consume, or
  // inspect your queues with QueueInspect.  It's unwise to mix Publish and
  // Consume to let TCP do its job well. 发布和消费最好是分开的连接, 不要混用

SSL/TLS - Secure connections

When Dial encounters an amqps:// scheme, it will use the zero value零值 of a tls.Config.
This will only perform server certificate and host verification.

Use DialTLS when you wish to provide a client certificate (recommended),
include a private certificate authority's certificate in the cert chain for
server validity, or run insecure by not verifying the server certificate dial
your own connection.
DialTLS will use the provided预设的 tls.Config when it encounters an amqps:// scheme
and will dial a plain明文 connection when it encounters an amqp:// scheme. DialTLS 自动适配两种协议

SSL/TLS in RabbitMQ is documented here: TODO http://www.rabbitmq.com/ssl.html

*/
package amqp
