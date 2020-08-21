// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"time"
)

func printBytes(prefix string, p []byte) {
	if prefix == "<-read" {
		p = trimBytes(p)
	}
	s := fmt.Sprintf(prefix+"(%d) [", len(p))
	for i := range p {
		s += fmt.Sprintf("%02X ", p[i])
	}
	if len(p) > 0 {
		s = s[:len(s)-1] + "]"
	}
	log.Printf(s)
}

func printBytess(prefix string, ps ...[]byte) {
	s := prefix
	for _, p := range ps {
		if prefix == "<-read" {
			p = trimBytes(p)
		}
		s += fmt.Sprintf(" (%d)[", len(p))
		for i := range p {
			s += fmt.Sprintf("%02X ", p[i])
		}
		if len(p) > 0 {
			s = s[:len(s)-1] + "]"
		}
	}
	log.Printf(s)
}

func printMethodFrame(mf *methodFrame) {
	mf.ClassId, mf.MethodId = mf.Method.id()
	bts, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("printMethodFrame %+v %s", mf, bts)
}

func printFrame(prefix string, mf frame) {
	bts, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s printFrame %+v %s", prefix, mf, bts)
}

func printHeaderFrame(mf *headerFrame) {
	bts, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("printHeaderFrame %+v %s", *mf, bts)
}

func printBodyFrame(mf *bodyFrame) {
	bts, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// 为什么 marshal之后 mf.Body, 不同于%s Body
	log.Printf("printBodyFrame %+v %s, %s", *mf, bts, mf.Body)
}

func trimBytes(p []byte) []byte {
	i := len(p) - 1
	for ; i >= 0; i-- {
		if p[i] != 0 {
			break
		}
	}
	return p[:i+1]
}

type FakerWriter struct {
	innerWriter io.Writer
}

func (f *FakerWriter) Write(p []byte) (n int, err error) {
	// if true || string(p) == "sss" {
	// 	debug.PrintStack()
	// }
	// log.Printf("send->`%s`\t%X", p, p)

	n, err = f.innerWriter.Write(p)
	fakerBytes := make([]byte, len(p))
	copy(fakerBytes, p)
	printBytes("->send", fakerBytes)
	return
}

type FakerBufWriter struct {
	*bufio.Writer
}

// 会产生死循环?
// func (b *FakerBufWriter) Flush() error {
// 	return b.Flush()
// }
//
// func (b *FakerBufWriter) Write(p []byte) (nn int, err error) {
// 	nn, err = b.Write(p)
// 	return
// }

// 写到buf, 然后刷新buf到内核缓冲区
func (w *writer) WriteFrame(frame frame) (err error) {
	fw := &FakerWriter{w.w}

	if err = frame.write(fw); err != nil {
		return
	}

	// 这是个struct, 不是接口
	if buf, ok := w.w.(*bufio.Writer); ok {
		fbw := &FakerBufWriter{buf}
		err = fbw.Flush()
	}

	return
}

func (f *methodFrame) write(w io.Writer) (err error) {
	var payload bytes.Buffer

	if f.Method == nil {
		return errors.New("malformed frame: missing method")
	}

	class, method := f.Method.id()

	if err = binary.Write(&payload, binary.BigEndian, class); err != nil {
		return
	}

	if err = binary.Write(&payload, binary.BigEndian, method); err != nil {
		return
	}

	if err = f.Method.write(&payload); err != nil {
		return
	}

	return writeFrame(w, frameMethod, f.ChannelId, payload.Bytes())
}

// Heartbeat
//
// Payload is empty
func (f *heartbeatFrame) write(w io.Writer) (err error) {
	return writeFrame(w, frameHeartbeat, f.ChannelId, []byte{})
}

// CONTENT HEADER
// 0          2        4           12               14
// +----------+--------+-----------+----------------+------------- - -
// | class-id | weight | body size | property flags | property list...
// +----------+--------+-----------+----------------+------------- - -
//    short     short    long long       short        remainder...
//
func (f *headerFrame) write(w io.Writer) (err error) {
	var payload bytes.Buffer
	var zeroTime time.Time

	if err = binary.Write(&payload, binary.BigEndian, f.ClassId); err != nil {
		return
	}

	if err = binary.Write(&payload, binary.BigEndian, f.weight); err != nil {
		return
	}

	if err = binary.Write(&payload, binary.BigEndian, f.Size); err != nil {
		return
	}

	// First pass will build the mask to be serialized, second pass will serialize
	// each of the fields that appear in the mask.

	var mask uint16

	if len(f.Properties.ContentType) > 0 {
		mask = mask | flagContentType
	}
	if len(f.Properties.ContentEncoding) > 0 {
		mask = mask | flagContentEncoding
	}
	if f.Properties.Headers != nil && len(f.Properties.Headers) > 0 {
		mask = mask | flagHeaders
	}
	if f.Properties.DeliveryMode > 0 {
		mask = mask | flagDeliveryMode
	}
	if f.Properties.Priority > 0 {
		mask = mask | flagPriority
	}
	if len(f.Properties.CorrelationId) > 0 {
		mask = mask | flagCorrelationId
	}
	if len(f.Properties.ReplyTo) > 0 {
		mask = mask | flagReplyTo
	}
	if len(f.Properties.Expiration) > 0 {
		mask = mask | flagExpiration
	}
	if len(f.Properties.MessageId) > 0 {
		mask = mask | flagMessageId
	}
	if f.Properties.Timestamp != zeroTime {
		mask = mask | flagTimestamp
	}
	if len(f.Properties.Type) > 0 {
		mask = mask | flagType
	}
	if len(f.Properties.UserId) > 0 {
		mask = mask | flagUserId
	}
	if len(f.Properties.AppId) > 0 {
		mask = mask | flagAppId
	}

	// flags
	if err = binary.Write(&payload, binary.BigEndian, mask); err != nil {
		return
	}

	if hasProperty(mask, flagContentType) {
		if err = writeShortstr(&payload, f.Properties.ContentType); err != nil {
			return
		}
	}
	if hasProperty(mask, flagContentEncoding) {
		if err = writeShortstr(&payload, f.Properties.ContentEncoding); err != nil {
			return
		}
	}
	if hasProperty(mask, flagHeaders) {
		if err = writeTable(&payload, f.Properties.Headers); err != nil {
			return
		}
	}
	if hasProperty(mask, flagDeliveryMode) {
		if err = binary.Write(&payload, binary.BigEndian, f.Properties.DeliveryMode); err != nil {
			return
		}
	}
	if hasProperty(mask, flagPriority) {
		if err = binary.Write(&payload, binary.BigEndian, f.Properties.Priority); err != nil {
			return
		}
	}
	if hasProperty(mask, flagCorrelationId) {
		if err = writeShortstr(&payload, f.Properties.CorrelationId); err != nil {
			return
		}
	}
	if hasProperty(mask, flagReplyTo) {
		if err = writeShortstr(&payload, f.Properties.ReplyTo); err != nil {
			return
		}
	}
	if hasProperty(mask, flagExpiration) {
		if err = writeShortstr(&payload, f.Properties.Expiration); err != nil {
			return
		}
	}
	if hasProperty(mask, flagMessageId) {
		if err = writeShortstr(&payload, f.Properties.MessageId); err != nil {
			return
		}
	}
	if hasProperty(mask, flagTimestamp) {
		if err = binary.Write(&payload, binary.BigEndian, uint64(f.Properties.Timestamp.Unix())); err != nil {
			return
		}
	}
	if hasProperty(mask, flagType) {
		if err = writeShortstr(&payload, f.Properties.Type); err != nil {
			return
		}
	}
	if hasProperty(mask, flagUserId) {
		if err = writeShortstr(&payload, f.Properties.UserId); err != nil {
			return
		}
	}
	if hasProperty(mask, flagAppId) {
		if err = writeShortstr(&payload, f.Properties.AppId); err != nil {
			return
		}
	}

	return writeFrame(w, frameHeader, f.ChannelId, payload.Bytes())
}

// Body
//
// Payload is one byterange from the full body who's size is declared in the
// Header frame
func (f *bodyFrame) write(w io.Writer) (err error) {
	return writeFrame(w, frameBody, f.ChannelId, f.Body)
}

func writeFrame(w io.Writer, typ uint8, channel uint16, payload []byte) (err error) {
	end := []byte{frameEnd}
	size := uint(len(payload))

	// head 7B type1B+channel2B+length4B 大端序 网络序
	head := []byte{
		byte(typ), // type 1B
		byte((channel & 0xff00) >> 8),
		byte((channel & 0x00ff) >> 0), // channel 2B
		byte((size & 0xff000000) >> 24),
		byte((size & 0x00ff0000) >> 16),
		byte((size & 0x0000ff00) >> 8),
		byte((size & 0x000000ff) >> 0), // length 4B
	}

	printBytess("->send\n Frame", head, payload, end)
	// debug.PrintStack()
	_, err = w.Write(head)

	if err != nil {
		return
	}

	if _, err = w.Write(payload); err != nil {
		return
	}

	// 补上ce
	if _, err = w.Write(end); err != nil {
		return
	}

	return
}

func writeShortstr(w io.Writer, s string) (err error) {
	b := []byte(s)

	var length = uint8(len(b))

	if err = binary.Write(w, binary.BigEndian, length); err != nil {
		return
	}

	if _, err = w.Write(b[:length]); err != nil {
		return
	}

	return
}

func writeLongstr(w io.Writer, s string) (err error) {
	b := []byte(s)

	var length = uint32(len(b))

	if err = binary.Write(w, binary.BigEndian, length); err != nil {
		return
	}

	if _, err = w.Write(b[:length]); err != nil {
		return
	}

	return
}

/*
'A': []interface{}
'D': Decimal
'F': Table
'I': int32
'S': string
'T': time.Time
'V': nil
'b': byte
'd': float64
'f': float32
'l': int64
's': int16
't': bool
'x': []byte
*/
func writeField(w io.Writer, value interface{}) (err error) {
	var buf [9]byte
	var enc []byte

	switch v := value.(type) {
	case bool:
		buf[0] = 't'
		if v {
			buf[1] = byte(1)
		} else {
			buf[1] = byte(0)
		}
		enc = buf[:2]

	case byte:
		buf[0] = 'b'
		buf[1] = byte(v)
		enc = buf[:2]

	case int16:
		buf[0] = 's'
		binary.BigEndian.PutUint16(buf[1:3], uint16(v))
		enc = buf[:3]

	case int:
		buf[0] = 'I'
		binary.BigEndian.PutUint32(buf[1:5], uint32(v))
		enc = buf[:5]

	case int32:
		buf[0] = 'I'
		binary.BigEndian.PutUint32(buf[1:5], uint32(v))
		enc = buf[:5]

	case int64:
		buf[0] = 'l'
		binary.BigEndian.PutUint64(buf[1:9], uint64(v))
		enc = buf[:9]

	case float32:
		buf[0] = 'f'
		binary.BigEndian.PutUint32(buf[1:5], math.Float32bits(v))
		enc = buf[:5]

	case float64:
		buf[0] = 'd'
		binary.BigEndian.PutUint64(buf[1:9], math.Float64bits(v))
		enc = buf[:9]

	case Decimal:
		buf[0] = 'D'
		buf[1] = byte(v.Scale)
		binary.BigEndian.PutUint32(buf[2:6], uint32(v.Value))
		enc = buf[:6]

	case string:
		buf[0] = 'S'
		binary.BigEndian.PutUint32(buf[1:5], uint32(len(v)))
		enc = append(buf[:5], []byte(v)...)

	case []interface{}: // field-array
		buf[0] = 'A'

		sec := new(bytes.Buffer)
		for _, val := range v {
			// 递归
			if err = writeField(sec, val); err != nil {
				return
			}
		}

		binary.BigEndian.PutUint32(buf[1:5], uint32(sec.Len()))
		if _, err = w.Write(buf[:5]); err != nil {
			return
		}

		if _, err = w.Write(sec.Bytes()); err != nil {
			return
		}

		return

	case time.Time:
		buf[0] = 'T'
		binary.BigEndian.PutUint64(buf[1:9], uint64(v.Unix()))
		enc = buf[:9]

	case Table:
		if _, err = w.Write([]byte{'F'}); err != nil {
			return
		}
		return writeTable(w, v)

	case []byte:
		buf[0] = 'x'
		binary.BigEndian.PutUint32(buf[1:5], uint32(len(v)))
		if _, err = w.Write(buf[0:5]); err != nil {
			return
		}
		if _, err = w.Write(v); err != nil {
			return
		}
		return

	case nil:
		buf[0] = 'V'
		enc = buf[:1]

	default:
		return ErrFieldType
	}

	_, err = w.Write(enc)

	return
}

// 帮助函数
func writeTable(w io.Writer, table Table) (err error) {
	var buf bytes.Buffer

	for key, val := range table {
		if err = writeShortstr(&buf, key); err != nil {
			return
		}
		if err = writeField(&buf, val); err != nil {
			return
		}
	}

	return writeLongstr(w, string(buf.Bytes()))
}
