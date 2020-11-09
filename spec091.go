// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

/* GENERATED FILE - DO NOT EDIT */
/* Rebuild from the spec/gen.go tool */

package amqp

// 协议特定, 请求结构体, 响应结构体

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Error codes that can be sent from the server during a connection or
// channel exception or used by the client to indicate a class of error like
// ErrCredentials.  The text of the error is likely more interesting than
// these constants.
const (
	frameMethod        = 1
	frameHeader        = 2
	frameBody          = 3
	frameHeartbeat     = 8
	frameMinSize       = 4096
	frameEnd           = 206 // 0xce
	replySuccess       = 200
	ContentTooLarge    = 311
	NoRoute            = 312
	NoConsumers        = 313
	ConnectionForced   = 320
	InvalidPath        = 402
	AccessRefused      = 403
	NotFound           = 404
	ResourceLocked     = 405
	PreconditionFailed = 406
	FrameError         = 501
	SyntaxError        = 502
	CommandInvalid     = 503
	ChannelError       = 504
	UnexpectedFrame    = 505
	ResourceError      = 506
	NotAllowed         = 530
	NotImplemented     = 540
	InternalError      = 541
)

var methodMap map[string]int

// 参考协议 https://www.rabbitmq.com/amqp-0-9-1-reference.html#exchange.declare.internal
func init() {
	methodMap = make(map[string]int)
	// connection
	// S->C 建立连接 开始连接协商过程
	methodMap["connectionStart"] = 10<<8 + 10
	// C->S 选择SASL安全机制和locale
	methodMap["connectionStartOk"] = 10<<8 + 11
	// S->C 要求客户端提供更多信息,以认证
	methodMap["connectionSecure"] = 10<<8 + 20
	// C->S 发送认证信息
	methodMap["connectionSecureOk"] = 10<<8 + 21
	// S->C 发送建议的连接级别的配置
	methodMap["connectionTune"] = 10<<8 + 30 // 5
	// C->S 配置协商结果反馈
	methodMap["connectionTuneOk"] = 10<<8 + 31

	// C->S 打开一个到虚拟节点的连接
	methodMap["connectionOpen"] = 10<<8 + 40
	// S->C 通知客户端 连接可用
	methodMap["connectionOpenOk"] = 10<<8 + 41

	// C->S or S->C 表示想要关闭连接
	methodMap["connectionClose"] = 10<<8 + 50
	// C->S or S->C 确认连接关闭, 接收者可以释放资源关闭socket连接
	methodMap["connectionCloseOk"] = 10<<8 + 51 // 10

	// 拓展协议 https://www.rabbitmq.com/connection-blocked.html
	// S->C 指示连接被阻塞了, 不接受新的发布, 服务器通知客户端, 资源太少, 需要被阻塞
	methodMap["connectionBlocked"] = 10<<8 + 60

	// S->C 指示连接阻塞解除了, 接受新的发布
	methodMap["connectionUnblocked"] = 10<<8 + 61

	// channel
	// C->S 打开一个对服务器的channel
	methodMap["channelOpen"] = 20<<8 + 10
	// S->C 通知客户端 channel可用
	methodMap["channelOpenOk"] = 20<<8 + 11

	// 流控双方都可以发, 服务器限制客户端发布速度, 或者客户端限制服务器consume后的deliver推送
	// 流控启用时, 这个Connection的状态每秒在blocked和unblocked之间来回切换数次，这样可以将消息发送的速率控制在服务器能够支撑的范围之内。
	// 客户端应该主动减少发布速度, 流控设置active=false才开启流量控制, 服务器如果不支持, 会返回504, 未实现
	// C->S or S->C 通知对端停止或者重启  流空, 避免过量, 不影响get获取, 只影响consumer和publish
	methodMap["channelFlow"] = 20<<8 + 20 // 15
	// C->S or S->C 向对方确认 已收到或者已处理
	methodMap["channelFlowOk"] = 20<<8 + 21

	// C->S or S->C 请求关闭一个Channel, 可能是因为被进程强制关闭或者发生了一个错误
	methodMap["channelClose"] = 20<<8 + 40
	// C->S or S->C 确认关闭, 通知接收者可以释放channel的资源
	methodMap["channelCloseOk"] = 20<<8 + 41

	// exchange
	// C->S 创建如果不存在, 如果已存在, 验证交换机是正确的以及期望的类型
	methodMap["exchangeDeclare"] = 40<<8 + 10
	// S->C 确认, 返回名字
	methodMap["exchangeDeclareOk"] = 40<<8 + 11 // 20

	// C->S 删除交换机
	methodMap["exchangeDelete"] = 40<<8 + 20
	// S->C 确认删除
	methodMap["exchangeDeleteOk"] = 40<<8 + 21

	// C->S 绑定交换机到交换机
	methodMap["exchangeBind"] = 40<<8 + 30
	// S->C 返回绑定成功
	methodMap["exchangeBindOk"] = 40<<8 + 31

	// C->S 解除绑定 交换机到交换机
	methodMap["exchangeUnbind"] = 40<<8 + 40
	// S->C 返回解绑成功
	methodMap["exchangeUnbindOk"] = 40<<8 + 51

	// queue
	// C->S 创建或者检查队列,
	methodMap["queueDeclare"] = 50<<8 + 10
	// S->C 确认, 确认返回名字
	methodMap["queueDeclareOk"] = 50<<8 + 11

	// C->S 队列绑定交换机
	methodMap["queueBind"] = 50<<8 + 20
	// S->C 确认绑定成功
	methodMap["queueBindOk"] = 50<<8 + 21 // 30

	// C->S 队列解除绑定交换机
	methodMap["queueUnbind"] = 50<<8 + 50
	// S->C 确认解除绑定成功
	methodMap["queueUnbindOk"] = 50<<8 + 51

	// C->S 删除所有非等待消费确认的消息, 此队列中 正在消费的消息不能删除
	methodMap["queuePurge"] = 50<<8 + 30
	// S->C 确认删除成功
	methodMap["queuePurgeOk"] = 50<<8 + 31

	// C->S 删除队列, 消息会被发送死信队列(如果定义配置了), 所有消费者都会被取消
	methodMap["queueDelete"] = 50<<8 + 40
	// S->C 确认队列的删除
	methodMap["queueDeleteOk"] = 50<<8 + 41

	// basic
	// C->S 请求特定服务质量, 可以针对channel和connection两种级别
	methodMap["basicQos"] = 60<<8 + 10
	// S->C
	methodMap["basicQosOk"] = 60<<8 + 11

	// C->S 要求服务器开始一个消费者, 生命周期直到channel结束或者cancel掉
	methodMap["basicConsume"] = 60<<8 + 20
	// S->C 给客户端一个consumerTag, 用于之后的消费者相关的调用
	methodMap["basicConsumeOk"] = 60<<8 + 21 // 40

	// C->S 取消一个消费者, 不影响已经递送的消息
	methodMap["basicCancel"] = 60<<8 + 30
	// S->C 确认取消消费者完成
	methodMap["basicCancelOk"] = 60<<8 + 31

	// C->S 发送消息 到指定交换机, 消息将被路由到一个队列
	methodMap["basicPublish"] = 60<<8 + 40

	// This method returns an undeliverable message that was published with the "immediate" flag set, or an
	// unroutable message published with the "mandatory" flag set
	// S->C 返回 一个未交付的消息(立刻immediate发布的消息或者无法路由的消息mandatory)
	methodMap["basicReturn"] = 60<<8 + 50

	// S->C 递送一个消息给客户端
	methodMap["basicDeliver"] = 60<<8 + 60

	// C->S 请求同步获取一条消息, 性能低
	methodMap["basicGet"] = 60<<8 + 70
	// S->C 响应消息
	methodMap["basicGetOk"] = 60<<8 + 71
	// S->C 响应无消息
	methodMap["basicGetEmpty"] = 60<<8 + 72

	// C->S 确认消费, rabbitmq 拓展, S->C 发送发布确认
	methodMap["basicAck"] = 60<<8 + 80

	// C->S 拒绝消费
	methodMap["basicReject"] = 60<<8 + 90 // 50

	// C->S 要求服务器重新递送, 在特定的channel上, 目前已弃用
	methodMap["basicRecoverAsync"] = 60<<8 + 100

	// C->S 要求服务器重新递送未确认的消息, 在特定的channel上
	methodMap["basicRecover"] = 60<<8 + 110
	// S->C 确认Recover
	methodMap["basicRecoverOk"] = 60<<8 + 111

	// C->S 拒绝一个或者多个消息
	methodMap["basicNack"] = 60<<8 + 120

	// tx
	// C->S 开启事务模式, 在使用commit or rollback 之前必须调用一次 channel级别
	methodMap["txSelect"] = 90<<8 + 10
	// S->C 确认此channel开启事务模式
	methodMap["txSelectOk"] = 90<<8 + 11

	// C->S 提交当前事务所有发布和确认, 之后新的事务立刻开启
	methodMap["txCommit"] = 90<<8 + 20
	// S->C 确认提交成功, 如果提交失败, 服务器触发一个channel级的异常
	methodMap["txCommitOk"] = 90<<8 + 21

	// C->S 丢弃所有, 新事务开启, 未确认的消息不会自动重发, 需要 主动recover
	methodMap["txRollback"] = 90<<8 + 30
	// S->C 确认rollback成功, 如果rollback失败, 服务器触发一个channel级的异常
	methodMap["txRollbackOk"] = 90<<8 + 31 // 60

	// confirm 进入发布确认模式(与事务模式无法并存), 服务器必须确认收到的所有消息
	// C->S rabbitmq 实现的 amqp0-9-1规范的拓展 https://www.rabbitmq.com/extensions.html
	methodMap["confirmSelect"] = 85<<8 + 10
	methodMap["confirmSelectOk"] = 85<<8 + 11
}

func isSoftExceptionCode(code int) bool {
	switch code {
	case 311: // body太大
		return true
	case 312: //
		return true
	case 313:
		return true
	case 403:
		return true
	case 404:
		return true
	case 405:
		return true
	case 406:
		return true

	}
	return false
}

type connectionStart struct {
	VersionMajor     byte   // amqp 协议大版本
	VersionMinor     byte   // amqp 协议小版本
	ServerProperties Table  // 服务器属性
	Mechanisms       string // 可用的安全机制
	Locales          string // 可用的locale
}

// classID,methodID
func (msg *connectionStart) id() (uint16, uint16) {
	return 10, 10
}

func (msg *connectionStart) wait() bool {
	return true
}

func (msg *connectionStart) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.VersionMajor); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.VersionMinor); err != nil {
		return
	}

	if err = writeTable(w, msg.ServerProperties); err != nil {
		return
	}

	if err = writeLongstr(w, msg.Mechanisms); err != nil {
		return
	}
	if err = writeLongstr(w, msg.Locales); err != nil {
		return
	}

	return
}

func (msg *connectionStart) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.VersionMajor); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.VersionMinor); err != nil {
		return
	}

	if msg.ServerProperties, err = readTable(r); err != nil {
		return
	}

	if msg.Mechanisms, err = readLongstr(r); err != nil {
		return
	}
	if msg.Locales, err = readLongstr(r); err != nil {
		return
	}

	return
}

type connectionStartOk struct {
	ClientProperties Table  // 客户端属性
	Mechanism        string // 选择的安全机制
	Response         string // 安全响应数据
	Locale           string // 选择的locale
}

func (msg *connectionStartOk) id() (uint16, uint16) {
	return 10, 11
}

func (msg *connectionStartOk) wait() bool {
	return true
}

func (msg *connectionStartOk) write(w io.Writer) (err error) {

	if err = writeTable(w, msg.ClientProperties); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Mechanism); err != nil {
		return
	}

	if err = writeLongstr(w, msg.Response); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Locale); err != nil {
		return
	}

	return
}

func (msg *connectionStartOk) read(r io.Reader) (err error) {

	if msg.ClientProperties, err = readTable(r); err != nil {
		return
	}

	if msg.Mechanism, err = readShortstr(r); err != nil {
		return
	}

	if msg.Response, err = readLongstr(r); err != nil {
		return
	}

	if msg.Locale, err = readShortstr(r); err != nil {
		return
	}

	return
}

// S->C
type connectionSecure struct {
	Challenge string // 请求客户端安全认证, 安全要求数据
}

func (msg *connectionSecure) id() (uint16, uint16) {
	return 10, 20
}

func (msg *connectionSecure) wait() bool {
	return true
}

func (msg *connectionSecure) write(w io.Writer) (err error) {

	if err = writeLongstr(w, msg.Challenge); err != nil {
		return
	}

	return
}

func (msg *connectionSecure) read(r io.Reader) (err error) {

	if msg.Challenge, err = readLongstr(r); err != nil {
		return
	}

	return
}

type connectionSecureOk struct {
	Response string // 安全响应数据
}

func (msg *connectionSecureOk) id() (uint16, uint16) {
	return 10, 21
}

func (msg *connectionSecureOk) wait() bool {
	return true
}

func (msg *connectionSecureOk) write(w io.Writer) (err error) {

	if err = writeLongstr(w, msg.Response); err != nil {
		return
	}

	return
}

func (msg *connectionSecureOk) read(r io.Reader) (err error) {

	if msg.Response, err = readLongstr(r); err != nil {
		return
	}

	return
}

type connectionTune struct {
	ChannelMax uint16 // 建议的最大channel数量
	FrameMax   uint32 // 建议的最大帧大小
	Heartbeat  uint16 // 渴望的心跳延迟
}

func (msg *connectionTune) id() (uint16, uint16) {
	return 10, 30
}

func (msg *connectionTune) wait() bool {
	return true
}

func (msg *connectionTune) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.FrameMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.Heartbeat); err != nil {
		return
	}

	return
}

func (msg *connectionTune) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.FrameMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.Heartbeat); err != nil {
		return
	}

	return
}

type connectionTuneOk struct {
	ChannelMax uint16 // 协商的最大channels
	FrameMax   uint32 // 协商的最大帧大小
	Heartbeat  uint16 // 渴望的心跳延迟
}

func (msg *connectionTuneOk) id() (uint16, uint16) {
	return 10, 31
}

func (msg *connectionTuneOk) wait() bool {
	return true
}

func (msg *connectionTuneOk) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.FrameMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.Heartbeat); err != nil {
		return
	}

	return
}

func (msg *connectionTuneOk) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.FrameMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.Heartbeat); err != nil {
		return
	}

	return
}

type connectionOpen struct {
	VirtualHost string // 虚拟主机名
	reserved1   string
	reserved2   bool
}

func (msg *connectionOpen) id() (uint16, uint16) {
	return 10, 40
}

func (msg *connectionOpen) wait() bool {
	return true
}

func (msg *connectionOpen) write(w io.Writer) (err error) {
	var bits byte

	if err = writeShortstr(w, msg.VirtualHost); err != nil {
		return
	}
	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	if msg.reserved2 {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *connectionOpen) read(r io.Reader) (err error) {
	var bits byte

	if msg.VirtualHost, err = readShortstr(r); err != nil {
		return
	}
	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.reserved2 = (bits&(1<<0) > 0)

	return
}

type connectionOpenOk struct {
	reserved1 string
}

func (msg *connectionOpenOk) id() (uint16, uint16) {
	return 10, 41
}

func (msg *connectionOpenOk) wait() bool {
	return true
}

func (msg *connectionOpenOk) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

func (msg *connectionOpenOk) read(r io.Reader) (err error) {

	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	return
}

type connectionClose struct {
	ReplyCode uint16
	ReplyText string
	ClassId   uint16 // 指示哪个类比的哪个方法导致了异常(如果是因为执行异常关闭)
	MethodId  uint16 // 如上
}

func (msg *connectionClose) id() (uint16, uint16) {
	return 10, 50
}

func (msg *connectionClose) wait() bool {
	return true
}

func (msg *connectionClose) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ReplyCode); err != nil {
		return
	}

	if err = writeShortstr(w, msg.ReplyText); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.ClassId); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.MethodId); err != nil {
		return
	}

	return
}

func (msg *connectionClose) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ReplyCode); err != nil {
		return
	}

	if msg.ReplyText, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.ClassId); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.MethodId); err != nil {
		return
	}

	return
}

type connectionCloseOk struct {
}

func (msg *connectionCloseOk) id() (uint16, uint16) {
	return 10, 51
}

func (msg *connectionCloseOk) wait() bool {
	return true
}

func (msg *connectionCloseOk) write(w io.Writer) (err error) {

	return
}

func (msg *connectionCloseOk) read(r io.Reader) (err error) {

	return
}

type connectionBlocked struct {
	Reason string // 阻塞的理由
}

func (msg *connectionBlocked) id() (uint16, uint16) {
	return 10, 60
}

func (msg *connectionBlocked) wait() bool {
	return false
}

func (msg *connectionBlocked) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.Reason); err != nil {
		return
	}

	return
}

func (msg *connectionBlocked) read(r io.Reader) (err error) {

	if msg.Reason, err = readShortstr(r); err != nil {
		return
	}

	return
}

type connectionUnblocked struct {
}

func (msg *connectionUnblocked) id() (uint16, uint16) {
	return 10, 61
}

func (msg *connectionUnblocked) wait() bool {
	return false
}

func (msg *connectionUnblocked) write(w io.Writer) (err error) {

	return
}

func (msg *connectionUnblocked) read(r io.Reader) (err error) {

	return
}

type channelOpen struct {
	reserved1 string
}

func (msg *channelOpen) id() (uint16, uint16) {
	return 20, 10
}

func (msg *channelOpen) wait() bool {
	return true
}

func (msg *channelOpen) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

func (msg *channelOpen) read(r io.Reader) (err error) {

	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	return
}

type channelOpenOk struct {
	reserved1 string
}

func (msg *channelOpenOk) id() (uint16, uint16) {
	return 20, 11
}

func (msg *channelOpenOk) wait() bool {
	return true
}

func (msg *channelOpenOk) write(w io.Writer) (err error) {

	if err = writeLongstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

func (msg *channelOpenOk) read(r io.Reader) (err error) {

	if msg.reserved1, err = readLongstr(r); err != nil {
		return
	}

	return
}

type channelFlow struct {
	Active bool // true 表示, 开始发送. false表示停止发送
}

func (msg *channelFlow) id() (uint16, uint16) {
	return 20, 20
}

func (msg *channelFlow) wait() bool {
	return true
}

func (msg *channelFlow) write(w io.Writer) (err error) {
	var bits byte

	if msg.Active {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *channelFlow) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Active = (bits&(1<<0) > 0)

	return
}

type channelFlowOk struct {
	Active bool // 当前的流设置
}

func (msg *channelFlowOk) id() (uint16, uint16) {
	return 20, 21
}

func (msg *channelFlowOk) wait() bool {
	return false
}

func (msg *channelFlowOk) write(w io.Writer) (err error) {
	var bits byte

	if msg.Active {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *channelFlowOk) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Active = (bits&(1<<0) > 0)

	return
}

type channelClose struct {
	ReplyCode uint16
	ReplyText string
	ClassId   uint16 // 指示哪个类比的哪个方法导致了异常(如果是因为执行异常关闭)
	MethodId  uint16 // 如上
}

func (msg *channelClose) id() (uint16, uint16) {
	return 20, 40
}

func (msg *channelClose) wait() bool {
	return true
}

func (msg *channelClose) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ReplyCode); err != nil {
		return
	}

	if err = writeShortstr(w, msg.ReplyText); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.ClassId); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.MethodId); err != nil {
		return
	}

	return
}

func (msg *channelClose) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ReplyCode); err != nil {
		return
	}

	if msg.ReplyText, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.ClassId); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.MethodId); err != nil {
		return
	}

	return
}

type channelCloseOk struct {
}

func (msg *channelCloseOk) id() (uint16, uint16) {
	return 20, 41
}

func (msg *channelCloseOk) wait() bool {
	return true
}

func (msg *channelCloseOk) write(w io.Writer) (err error) {

	return
}

func (msg *channelCloseOk) read(r io.Reader) (err error) {

	return
}

type exchangeDeclare struct {
	reserved1  uint16
	Exchange   string // 交换机名字
	Type       string // 交换机类型
	Passive    bool   // 不要创建交换机, 只是检查
	Durable    bool   // 创建持久交换机
	AutoDelete bool   // 绑定的队列解绑后, 如果没有一个队列与之绑定, 则删除此交换机
	Internal   bool   // 内部交换机, 不能直接发布消息到此交换机
	NoWait     bool   // true, 表示不等待服务器回复, 服务器不会回复此请求, 如果服务器无法完成, 会抛出异常
	Arguments  Table  // 额外参数, 取决于服务器实现, 支持 死信队列, ttl
}

func (msg *exchangeDeclare) id() (uint16, uint16) {
	return 40, 10
}

func (msg *exchangeDeclare) wait() bool {
	return true && !msg.NoWait
}

func (msg *exchangeDeclare) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Type); err != nil {
		return
	}

	if msg.Passive {
		bits |= 1 << 0
	}

	if msg.Durable {
		bits |= 1 << 1
	}

	if msg.AutoDelete {
		bits |= 1 << 2
	}

	if msg.Internal {
		bits |= 1 << 3
	}

	if msg.NoWait {
		bits |= 1 << 4
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *exchangeDeclare) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.Type, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Passive = (bits&(1<<0) > 0)
	msg.Durable = (bits&(1<<1) > 0)
	msg.AutoDelete = (bits&(1<<2) > 0)
	msg.Internal = (bits&(1<<3) > 0)
	msg.NoWait = (bits&(1<<4) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type exchangeDeclareOk struct {
}

func (msg *exchangeDeclareOk) id() (uint16, uint16) {
	return 40, 11
}

func (msg *exchangeDeclareOk) wait() bool {
	return true
}

func (msg *exchangeDeclareOk) write(w io.Writer) (err error) {

	return
}

func (msg *exchangeDeclareOk) read(r io.Reader) (err error) {

	return
}

type exchangeDelete struct {
	reserved1 uint16
	Exchange  string // 交换机名字
	IfUnused  bool   // true, 表示只在未使用时删除
	NoWait    bool   // true, 表示不等待服务器回复, 服务器不会回复此请求, 如果服务器无法完成, 会抛出异常
}

func (msg *exchangeDelete) id() (uint16, uint16) {
	return 40, 20
}

func (msg *exchangeDelete) wait() bool {
	return true && !msg.NoWait
}

func (msg *exchangeDelete) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}

	if msg.IfUnused {
		bits |= 1 << 0
	}

	if msg.NoWait {
		bits |= 1 << 1
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *exchangeDelete) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.IfUnused = (bits&(1<<0) > 0)
	msg.NoWait = (bits&(1<<1) > 0)

	return
}

type exchangeDeleteOk struct {
}

func (msg *exchangeDeleteOk) id() (uint16, uint16) {
	return 40, 21
}

func (msg *exchangeDeleteOk) wait() bool {
	return true
}

func (msg *exchangeDeleteOk) write(w io.Writer) (err error) {

	return
}

func (msg *exchangeDeleteOk) read(r io.Reader) (err error) {

	return
}

type exchangeBind struct {
	reserved1   uint16
	Destination string
	Source      string
	RoutingKey  string
	NoWait      bool
	Arguments   Table
}

func (msg *exchangeBind) id() (uint16, uint16) {
	return 40, 30
}

func (msg *exchangeBind) wait() bool {
	return true && !msg.NoWait
}

func (msg *exchangeBind) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Destination); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Source); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *exchangeBind) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Destination, err = readShortstr(r); err != nil {
		return
	}
	if msg.Source, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type exchangeBindOk struct {
}

func (msg *exchangeBindOk) id() (uint16, uint16) {
	return 40, 31
}

func (msg *exchangeBindOk) wait() bool {
	return true
}

func (msg *exchangeBindOk) write(w io.Writer) (err error) {

	return
}

func (msg *exchangeBindOk) read(r io.Reader) (err error) {

	return
}

type exchangeUnbind struct {
	reserved1   uint16
	Destination string
	Source      string
	RoutingKey  string
	NoWait      bool
	Arguments   Table
}

func (msg *exchangeUnbind) id() (uint16, uint16) {
	return 40, 40
}

func (msg *exchangeUnbind) wait() bool {
	return true && !msg.NoWait
}

func (msg *exchangeUnbind) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Destination); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Source); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *exchangeUnbind) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Destination, err = readShortstr(r); err != nil {
		return
	}
	if msg.Source, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type exchangeUnbindOk struct {
}

func (msg *exchangeUnbindOk) id() (uint16, uint16) {
	return 40, 51
}

func (msg *exchangeUnbindOk) wait() bool {
	return true
}

func (msg *exchangeUnbindOk) write(w io.Writer) (err error) {

	return
}

func (msg *exchangeUnbindOk) read(r io.Reader) (err error) {

	return
}

type queueDeclare struct {
	reserved1  uint16
	Queue      string // 队列名
	Passive    bool   // 不创建队列
	Durable    bool   // 创建持久队列
	Exclusive  bool   // 只能由当前连接访问, 连接中断, 队列就删除了
	AutoDelete bool   // 所有消费者结束使用了就自动删除此队列
	NoWait     bool   // true, 表示不等待服务器回复, 服务器不会回复此请求, 如果服务器无法完成, 会抛出异常
	Arguments  Table  // 额外参数, 取决于服务器实现
}

func (msg *queueDeclare) id() (uint16, uint16) {
	return 50, 10
}

func (msg *queueDeclare) wait() bool {
	return true && !msg.NoWait
}

func (msg *queueDeclare) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.Passive {
		bits |= 1 << 0
	}

	if msg.Durable {
		bits |= 1 << 1
	}

	if msg.Exclusive {
		bits |= 1 << 2
	}

	if msg.AutoDelete {
		bits |= 1 << 3
	}

	if msg.NoWait {
		bits |= 1 << 4
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *queueDeclare) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Passive = (bits&(1<<0) > 0)
	msg.Durable = (bits&(1<<1) > 0)
	msg.Exclusive = (bits&(1<<2) > 0)
	msg.AutoDelete = (bits&(1<<3) > 0)
	msg.NoWait = (bits&(1<<4) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type queueDeclareOk struct {
	Queue         string // 队列名
	MessageCount  uint32 // 队列里的消息数量
	ConsumerCount uint32 // 队列当前的消费者数量
}

func (msg *queueDeclareOk) id() (uint16, uint16) {
	return 50, 11
}

func (msg *queueDeclareOk) wait() bool {
	return true
}

func (msg *queueDeclareOk) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.ConsumerCount); err != nil {
		return
	}

	return
}

func (msg *queueDeclareOk) read(r io.Reader) (err error) {

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.ConsumerCount); err != nil {
		return
	}

	return
}

type queueBind struct {
	reserved1  uint16
	Queue      string // 队列名, 如果队列名为空, 服务器使用channel上最后声明的队列
	Exchange   string // 交换机名
	RoutingKey string // 队列 路由键 绑定键
	NoWait     bool   // true, 表示不等待服务器回复, 服务器不会回复此请求, 如果服务器无法完成, 会抛出异常
	Arguments  Table  // 参数, 服务器实现特定
}

func (msg *queueBind) id() (uint16, uint16) {
	return 50, 20
}

func (msg *queueBind) wait() bool {
	return true && !msg.NoWait
}

func (msg *queueBind) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *queueBind) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}
	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type queueBindOk struct {
}

func (msg *queueBindOk) id() (uint16, uint16) {
	return 50, 21
}

func (msg *queueBindOk) wait() bool {
	return true
}

func (msg *queueBindOk) write(w io.Writer) (err error) {

	return
}

func (msg *queueBindOk) read(r io.Reader) (err error) {

	return
}

type queueUnbind struct {
	reserved1  uint16
	Queue      string // 队列名
	Exchange   string // 交换机名
	RoutingKey string // 路由键
	Arguments  Table  // 参数
}

func (msg *queueUnbind) id() (uint16, uint16) {
	return 50, 50
}

func (msg *queueUnbind) wait() bool {
	return true
}

func (msg *queueUnbind) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *queueUnbind) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}
	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type queueUnbindOk struct {
}

func (msg *queueUnbindOk) id() (uint16, uint16) {
	return 50, 51
}

func (msg *queueUnbindOk) wait() bool {
	return true
}

func (msg *queueUnbindOk) write(w io.Writer) (err error) {

	return
}

func (msg *queueUnbindOk) read(r io.Reader) (err error) {

	return
}

type queuePurge struct {
	reserved1 uint16
	Queue     string // 队列名
	NoWait    bool   // true, 表示不等待服务器回复, 服务器不会回复此请求, 如果服务器无法完成, 会抛出异常
}

func (msg *queuePurge) id() (uint16, uint16) {
	return 50, 30
}

func (msg *queuePurge) wait() bool {
	return true && !msg.NoWait
}

func (msg *queuePurge) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *queuePurge) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	return
}

type queuePurgeOk struct {
	MessageCount uint32 // 有多少个消息被清洗了
}

func (msg *queuePurgeOk) id() (uint16, uint16) {
	return 50, 31
}

func (msg *queuePurgeOk) wait() bool {
	return true
}

func (msg *queuePurgeOk) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}

	return
}

func (msg *queuePurgeOk) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}

	return
}

type queueDelete struct {
	reserved1 uint16
	Queue     string // 队列名
	IfUnused  bool   // 未使用才删除
	IfEmpty   bool   // 没有消息才删除
	NoWait    bool   // true, 表示不等待服务器回复, 服务器不会回复此请求, 如果服务器无法完成, 会抛出异常
}

func (msg *queueDelete) id() (uint16, uint16) {
	return 50, 40
}

func (msg *queueDelete) wait() bool {
	return true && !msg.NoWait
}

func (msg *queueDelete) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.IfUnused {
		bits |= 1 << 0
	}

	if msg.IfEmpty {
		bits |= 1 << 1
	}

	if msg.NoWait {
		bits |= 1 << 2
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *queueDelete) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.IfUnused = (bits&(1<<0) > 0)
	msg.IfEmpty = (bits&(1<<1) > 0)
	msg.NoWait = (bits&(1<<2) > 0)

	return
}

type queueDeleteOk struct {
	MessageCount uint32 // 有多少个消息被删除了
}

func (msg *queueDeleteOk) id() (uint16, uint16) {
	return 50, 41
}

func (msg *queueDeleteOk) wait() bool {
	return true
}

func (msg *queueDeleteOk) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}

	return
}

func (msg *queueDeleteOk) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}

	return
}

type basicQos struct { // 当前只对服务器端有意义
	PrefetchSize  uint32 // 预取窗口 字节数
	PrefetchCount uint16 // 预取窗口 消息数
	Global        bool   // 是否是全connection配置还是channel配置
}

func (msg *basicQos) id() (uint16, uint16) {
	return 60, 10
}

func (msg *basicQos) wait() bool {
	return true
}

func (msg *basicQos) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.PrefetchSize); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.PrefetchCount); err != nil {
		return
	}

	if msg.Global {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicQos) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.PrefetchSize); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.PrefetchCount); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Global = (bits&(1<<0) > 0)

	return
}

type basicQosOk struct {
}

func (msg *basicQosOk) id() (uint16, uint16) {
	return 60, 11
}

func (msg *basicQosOk) wait() bool {
	return true
}

func (msg *basicQosOk) write(w io.Writer) (err error) {

	return
}

func (msg *basicQosOk) read(r io.Reader) (err error) {

	return
}

type basicConsume struct {
	reserved1   uint16
	Queue       string // 队列名
	ConsumerTag string // 指定消费者标志, 只属于channel级别, 如果为"", 服务器会生成一个tag
	NoLocal     bool   // 不会发送回给此连接上, 包括所有的channel, 也就是不会转回来 https://www.rabbitmq.com/amqp-0-9-1-reference.html#domain.
	// no-local
	NoAck     bool // 告诉服务器不需要等待ack,reject,nack 之类的回复
	Exclusive bool // 消费的排他性, 只有此消费者, 能访问此服务
	NoWait    bool // 服务器不会发送请求的响应, 有异常
	Arguments Table
}

func (msg *basicConsume) id() (uint16, uint16) {
	return 60, 20
}

func (msg *basicConsume) wait() bool {
	return true && !msg.NoWait
}

func (msg *basicConsume) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	if msg.NoLocal {
		bits |= 1 << 0
	}

	if msg.NoAck {
		bits |= 1 << 1
	}

	if msg.Exclusive {
		bits |= 1 << 2
	}

	if msg.NoWait {
		bits |= 1 << 3
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *basicConsume) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}
	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoLocal = (bits&(1<<0) > 0)
	msg.NoAck = (bits&(1<<1) > 0)
	msg.Exclusive = (bits&(1<<2) > 0)
	msg.NoWait = (bits&(1<<3) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type basicConsumeOk struct {
	ConsumerTag string // 返回确定的tag, 用于取消消费指定消费者
}

func (msg *basicConsumeOk) id() (uint16, uint16) {
	return 60, 21
}

func (msg *basicConsumeOk) wait() bool {
	return true
}

func (msg *basicConsumeOk) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	return
}

func (msg *basicConsumeOk) read(r io.Reader) (err error) {

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	return
}

type basicCancel struct {
	ConsumerTag string // 取消消费
	NoWait      bool
}

func (msg *basicCancel) id() (uint16, uint16) {
	return 60, 30
}

func (msg *basicCancel) wait() bool {
	return true && !msg.NoWait
}

func (msg *basicCancel) write(w io.Writer) (err error) {
	var bits byte

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicCancel) read(r io.Reader) (err error) {
	var bits byte

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	return
}

type basicCancelOk struct {
	ConsumerTag string // 返回取消消费成功的tag
}

func (msg *basicCancelOk) id() (uint16, uint16) {
	return 60, 31
}

func (msg *basicCancelOk) wait() bool {
	return true
}

func (msg *basicCancelOk) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	return
}

func (msg *basicCancelOk) read(r io.Reader) (err error) {

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	return
}

type basicPublish struct {
	reserved1  uint16
	Exchange   string
	RoutingKey string
	Mandatory  bool // 立即强制路由, 如果消息不能路由, 以Return返回, false,则直接丢弃
	Immediate  bool // 减少延迟, 请求立刻分发, 如果消息不能立即交给消费者, 以Return返回, false,则直接排队
	Properties properties
	Body       []byte
}

func (msg *basicPublish) id() (uint16, uint16) {
	return 60, 40
}

func (msg *basicPublish) wait() bool {
	return false
}

func (msg *basicPublish) getContent() (properties, []byte) {
	return msg.Properties, msg.Body
}

func (msg *basicPublish) setContent(props properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

func (msg *basicPublish) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.Mandatory {
		bits |= 1 << 0
	}

	if msg.Immediate {
		bits |= 1 << 1
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicPublish) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Mandatory = (bits&(1<<0) > 0)
	msg.Immediate = (bits&(1<<1) > 0)

	return
}

type basicReturn struct {
	ReplyCode  uint16
	ReplyText  string
	Exchange   string
	RoutingKey string
	Properties properties // 属性也会一并return
	Body       []byte     // 消息
}

func (msg *basicReturn) id() (uint16, uint16) {
	return 60, 50
}

func (msg *basicReturn) wait() bool {
	return false
}

func (msg *basicReturn) getContent() (properties, []byte) {
	return msg.Properties, msg.Body
}

func (msg *basicReturn) setContent(props properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

func (msg *basicReturn) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ReplyCode); err != nil {
		return
	}

	if err = writeShortstr(w, msg.ReplyText); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	return
}

func (msg *basicReturn) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ReplyCode); err != nil {
		return
	}

	if msg.ReplyText, err = readShortstr(r); err != nil {
		return
	}
	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	return
}

type basicDeliver struct {
	ConsumerTag string
	DeliveryTag uint64 // 服务器指定的, channel特定的, 不能用0值, 0保留客户端ack,reject,nack使用, 表示到目前为止收到的所有消息
	Redelivered bool   // 之前是否会发送到这个或者其他客户端
	Exchange    string
	RoutingKey  string
	Properties  properties
	Body        []byte // 消息体
}

func (msg *basicDeliver) id() (uint16, uint16) {
	return 60, 60
}

func (msg *basicDeliver) wait() bool {
	return false
}

func (msg *basicDeliver) getContent() (properties, []byte) {
	return msg.Properties, msg.Body
}

func (msg *basicDeliver) setContent(props properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

func (msg *basicDeliver) write(w io.Writer) (err error) {
	var bits byte

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Redelivered {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	return
}

func (msg *basicDeliver) read(r io.Reader) (err error) {
	var bits byte

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Redelivered = (bits&(1<<0) > 0)

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	return
}

type basicGet struct {
	reserved1 uint16
	Queue     string
	NoAck     bool // 不进行ack流程 告诉服务器不需要等待ack,reject,nack 之类的回复
}

func (msg *basicGet) id() (uint16, uint16) {
	return 60, 70
}

func (msg *basicGet) wait() bool {
	return true
}

func (msg *basicGet) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.NoAck {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicGet) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoAck = (bits&(1<<0) > 0)

	return
}

type basicGetOk struct {
	DeliveryTag  uint64 // 唯一id
	Redelivered  bool   // 重递送
	Exchange     string
	RoutingKey   string
	MessageCount uint32     // 队列里的消息数量
	Properties   properties // 消息属性
	Body         []byte
}

func (msg *basicGetOk) id() (uint16, uint16) {
	return 60, 71
}

func (msg *basicGetOk) wait() bool {
	return true
}

func (msg *basicGetOk) getContent() (properties, []byte) {
	return msg.Properties, msg.Body
}

func (msg *basicGetOk) setContent(props properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

func (msg *basicGetOk) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Redelivered {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}

	return
}

func (msg *basicGetOk) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Redelivered = (bits&(1<<0) > 0)

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}

	return
}

type basicGetEmpty struct {
	reserved1 string
}

func (msg *basicGetEmpty) id() (uint16, uint16) {
	return 60, 72
}

func (msg *basicGetEmpty) wait() bool {
	return true
}

func (msg *basicGetEmpty) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

func (msg *basicGetEmpty) read(r io.Reader) (err error) {

	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	return
}

type basicAck struct {
	DeliveryTag uint64 // 0表示到目前为止收到的所有消息
	Multiple    bool   // 确认之前到现在的多个消息
}

func (msg *basicAck) id() (uint16, uint16) {
	return 60, 80
}

func (msg *basicAck) wait() bool {
	return false
}

func (msg *basicAck) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Multiple {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicAck) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Multiple = (bits&(1<<0) > 0)

	return
}

type basicReject struct {
	DeliveryTag uint64
	Requeue     bool // 重新入队, false, 或者入队失败, 就会丢弃或者到死信队列
}

func (msg *basicReject) id() (uint16, uint16) {
	return 60, 90
}

func (msg *basicReject) wait() bool {
	return false
}

func (msg *basicReject) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Requeue {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicReject) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Requeue = (bits&(1<<0) > 0)

	return
}

type basicRecoverAsync struct {
	Requeue bool // false, 则重新发送给原来的消费者, true, 则会入队, 会发给其他消费者
}

func (msg *basicRecoverAsync) id() (uint16, uint16) {
	return 60, 100
}

func (msg *basicRecoverAsync) wait() bool {
	return false
}

func (msg *basicRecoverAsync) write(w io.Writer) (err error) {
	var bits byte

	if msg.Requeue {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicRecoverAsync) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Requeue = (bits&(1<<0) > 0)

	return
}

type basicRecover struct {
	Requeue bool // false, 则重新发送给原来的消费者, true, 则会入队, 会发给其他消费者
}

func (msg *basicRecover) id() (uint16, uint16) {
	return 60, 110
}

func (msg *basicRecover) wait() bool {
	return true
}

func (msg *basicRecover) write(w io.Writer) (err error) {
	var bits byte

	if msg.Requeue {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicRecover) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Requeue = (bits&(1<<0) > 0)

	return
}

type basicRecoverOk struct {
}

func (msg *basicRecoverOk) id() (uint16, uint16) {
	return 60, 111
}

func (msg *basicRecoverOk) wait() bool {
	return true
}

func (msg *basicRecoverOk) write(w io.Writer) (err error) {

	return
}

func (msg *basicRecoverOk) read(r io.Reader) (err error) {

	return
}

type basicNack struct {
	DeliveryTag uint64
	Multiple    bool //
	Requeue     bool // 重新入队, 以便投递给其他消费者, 如果, 入队失败或者是false, 会到死信队列, or 直接丢弃
}

func (msg *basicNack) id() (uint16, uint16) {
	return 60, 120
}

func (msg *basicNack) wait() bool {
	return false
}

func (msg *basicNack) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Multiple {
		bits |= 1 << 0
	}

	if msg.Requeue {
		bits |= 1 << 1
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicNack) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Multiple = (bits&(1<<0) > 0)
	msg.Requeue = (bits&(1<<1) > 0)

	return
}

type txSelect struct {
}

func (msg *txSelect) id() (uint16, uint16) {
	return 90, 10
}

func (msg *txSelect) wait() bool {
	return true
}

func (msg *txSelect) write(w io.Writer) (err error) {

	return
}

func (msg *txSelect) read(r io.Reader) (err error) {

	return
}

type txSelectOk struct {
}

func (msg *txSelectOk) id() (uint16, uint16) {
	return 90, 11
}

func (msg *txSelectOk) wait() bool {
	return true
}

func (msg *txSelectOk) write(w io.Writer) (err error) {

	return
}

func (msg *txSelectOk) read(r io.Reader) (err error) {

	return
}

type txCommit struct {
}

func (msg *txCommit) id() (uint16, uint16) {
	return 90, 20
}

func (msg *txCommit) wait() bool {
	return true
}

func (msg *txCommit) write(w io.Writer) (err error) {

	return
}

func (msg *txCommit) read(r io.Reader) (err error) {

	return
}

type txCommitOk struct {
}

func (msg *txCommitOk) id() (uint16, uint16) {
	return 90, 21
}

func (msg *txCommitOk) wait() bool {
	return true
}

func (msg *txCommitOk) write(w io.Writer) (err error) {

	return
}

func (msg *txCommitOk) read(r io.Reader) (err error) {

	return
}

type txRollback struct {
}

func (msg *txRollback) id() (uint16, uint16) {
	return 90, 30
}

func (msg *txRollback) wait() bool {
	return true
}

func (msg *txRollback) write(w io.Writer) (err error) {

	return
}

func (msg *txRollback) read(r io.Reader) (err error) {

	return
}

type txRollbackOk struct {
}

func (msg *txRollbackOk) id() (uint16, uint16) {
	return 90, 31
}

func (msg *txRollbackOk) wait() bool {
	return true
}

func (msg *txRollbackOk) write(w io.Writer) (err error) {

	return
}

func (msg *txRollbackOk) read(r io.Reader) (err error) {

	return
}

type confirmSelect struct {
	Nowait bool // 不等待服务器响应
}

func (msg *confirmSelect) id() (uint16, uint16) {
	return 85, 10
}

func (msg *confirmSelect) wait() bool {
	return true
}

func (msg *confirmSelect) write(w io.Writer) (err error) {
	var bits byte

	if msg.Nowait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *confirmSelect) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Nowait = (bits&(1<<0) > 0)

	return
}

type confirmSelectOk struct {
}

func (msg *confirmSelectOk) id() (uint16, uint16) {
	return 85, 11
}

func (msg *confirmSelectOk) wait() bool {
	return true
}

func (msg *confirmSelectOk) write(w io.Writer) (err error) {

	return
}

func (msg *confirmSelectOk) read(r io.Reader) (err error) {

	return
}

// 解析方法帧
func (r *reader) parseMethodFrame(channel uint16, size uint32) (f frame, err error) {
	mf := &methodFrame{
		ChannelId: channel,
	}

	if err = binary.Read(r.r, binary.BigEndian, &mf.ClassId); err != nil {
		return
	}

	if err = binary.Read(r.r, binary.BigEndian, &mf.MethodId); err != nil {
		return
	}

	switch mf.ClassId {

	case 10: // connection
		switch mf.MethodId {

		case 10: // connection start
			// fmt.Println("NextMethod: class:10 method:10")
			method := &connectionStart{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // connection start-ok
			// fmt.Println("NextMethod: class:10 method:11")
			method := &connectionStartOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // connection secure
			// fmt.Println("NextMethod: class:10 method:20")
			method := &connectionSecure{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // connection secure-ok
			// fmt.Println("NextMethod: class:10 method:21")
			method := &connectionSecureOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // connection tune
			// fmt.Println("NextMethod: class:10 method:30")
			method := &connectionTune{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // connection tune-ok
			// fmt.Println("NextMethod: class:10 method:31")
			method := &connectionTuneOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // connection open
			// fmt.Println("NextMethod: class:10 method:40")
			method := &connectionOpen{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 41: // connection open-ok
			// fmt.Println("NextMethod: class:10 method:41")
			method := &connectionOpenOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 50: // connection close
			// fmt.Println("NextMethod: class:10 method:50")
			method := &connectionClose{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 51: // connection close-ok
			// fmt.Println("NextMethod: class:10 method:51")
			method := &connectionCloseOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 60: // connection blocked
			// fmt.Println("NextMethod: class:10 method:60")
			method := &connectionBlocked{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 61: // connection unblocked
			// fmt.Println("NextMethod: class:10 method:61")
			method := &connectionUnblocked{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 20: // channel
		switch mf.MethodId {

		case 10: // channel open
			// fmt.Println("NextMethod: class:20 method:10")
			method := &channelOpen{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // channel open-ok
			// fmt.Println("NextMethod: class:20 method:11")
			method := &channelOpenOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // channel flow
			// fmt.Println("NextMethod: class:20 method:20")
			method := &channelFlow{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // channel flow-ok
			// fmt.Println("NextMethod: class:20 method:21")
			method := &channelFlowOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // channel close
			// fmt.Println("NextMethod: class:20 method:40")
			method := &channelClose{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 41: // channel close-ok
			// fmt.Println("NextMethod: class:20 method:41")
			method := &channelCloseOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 40: // exchange
		switch mf.MethodId {

		case 10: // exchange declare
			// fmt.Println("NextMethod: class:40 method:10")
			method := &exchangeDeclare{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // exchange declare-ok
			// fmt.Println("NextMethod: class:40 method:11")
			method := &exchangeDeclareOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // exchange delete
			// fmt.Println("NextMethod: class:40 method:20")
			method := &exchangeDelete{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // exchange delete-ok
			// fmt.Println("NextMethod: class:40 method:21")
			method := &exchangeDeleteOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // exchange bind
			// fmt.Println("NextMethod: class:40 method:30")
			method := &exchangeBind{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // exchange bind-ok
			// fmt.Println("NextMethod: class:40 method:31")
			method := &exchangeBindOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // exchange unbind
			// fmt.Println("NextMethod: class:40 method:40")
			method := &exchangeUnbind{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 51: // exchange unbind-ok
			// fmt.Println("NextMethod: class:40 method:51")
			method := &exchangeUnbindOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 50: // queue
		switch mf.MethodId {

		case 10: // queue declare
			// fmt.Println("NextMethod: class:50 method:10")
			method := &queueDeclare{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // queue declare-ok
			// fmt.Println("NextMethod: class:50 method:11")
			method := &queueDeclareOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // queue bind
			// fmt.Println("NextMethod: class:50 method:20")
			method := &queueBind{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // queue bind-ok
			// fmt.Println("NextMethod: class:50 method:21")
			method := &queueBindOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 50: // queue unbind
			// fmt.Println("NextMethod: class:50 method:50")
			method := &queueUnbind{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 51: // queue unbind-ok
			// fmt.Println("NextMethod: class:50 method:51")
			method := &queueUnbindOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // queue purge
			// fmt.Println("NextMethod: class:50 method:30")
			method := &queuePurge{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // queue purge-ok
			// fmt.Println("NextMethod: class:50 method:31")
			method := &queuePurgeOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // queue delete
			// fmt.Println("NextMethod: class:50 method:40")
			method := &queueDelete{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 41: // queue delete-ok
			// fmt.Println("NextMethod: class:50 method:41")
			method := &queueDeleteOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 60: // basic
		switch mf.MethodId {

		case 10: // basic qos
			// fmt.Println("NextMethod: class:60 method:10")
			method := &basicQos{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // basic qos-ok
			// fmt.Println("NextMethod: class:60 method:11")
			method := &basicQosOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // basic consume
			// fmt.Println("NextMethod: class:60 method:20")
			method := &basicConsume{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // basic consume-ok
			// fmt.Println("NextMethod: class:60 method:21")
			method := &basicConsumeOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // basic cancel
			// fmt.Println("NextMethod: class:60 method:30")
			method := &basicCancel{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // basic cancel-ok
			// fmt.Println("NextMethod: class:60 method:31")
			method := &basicCancelOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // basic publish
			// fmt.Println("NextMethod: class:60 method:40")
			method := &basicPublish{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 50: // basic return
			// fmt.Println("NextMethod: class:60 method:50")
			method := &basicReturn{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 60: // basic deliver
			// fmt.Println("NextMethod: class:60 method:60")
			method := &basicDeliver{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 70: // basic get
			// fmt.Println("NextMethod: class:60 method:70")
			method := &basicGet{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 71: // basic get-ok
			// fmt.Println("NextMethod: class:60 method:71")
			method := &basicGetOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 72: // basic get-empty
			// fmt.Println("NextMethod: class:60 method:72")
			method := &basicGetEmpty{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 80: // basic ack
			// fmt.Println("NextMethod: class:60 method:80")
			method := &basicAck{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 90: // basic reject
			// fmt.Println("NextMethod: class:60 method:90")
			method := &basicReject{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 100: // basic recover-async
			// fmt.Println("NextMethod: class:60 method:100")
			method := &basicRecoverAsync{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 110: // basic recover
			// fmt.Println("NextMethod: class:60 method:110")
			method := &basicRecover{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 111: // basic recover-ok
			// fmt.Println("NextMethod: class:60 method:111")
			method := &basicRecoverOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 120: // basic nack
			// fmt.Println("NextMethod: class:60 method:120")
			method := &basicNack{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 90: // tx
		switch mf.MethodId {

		case 10: // tx select
			// fmt.Println("NextMethod: class:90 method:10")
			method := &txSelect{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // tx select-ok
			// fmt.Println("NextMethod: class:90 method:11")
			method := &txSelectOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // tx commit
			// fmt.Println("NextMethod: class:90 method:20")
			method := &txCommit{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // tx commit-ok
			// fmt.Println("NextMethod: class:90 method:21")
			method := &txCommitOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // tx rollback
			// fmt.Println("NextMethod: class:90 method:30")
			method := &txRollback{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // tx rollback-ok
			// fmt.Println("NextMethod: class:90 method:31")
			method := &txRollbackOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 85: // confirm
		switch mf.MethodId {

		case 10: // confirm select
			// fmt.Println("NextMethod: class:85 method:10")
			method := &confirmSelect{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // confirm select-ok
			// fmt.Println("NextMethod: class:85 method:11")
			method := &confirmSelectOk{}
			if err = method.read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	default:
		return nil, fmt.Errorf("bad method frame, unknown class %d", mf.ClassId)
	}

	return mf, nil
}
