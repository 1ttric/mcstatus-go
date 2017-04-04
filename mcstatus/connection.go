package mcstatus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

// Connection buffer

func NewConnection() Connection {
	return Connection{[]byte{}, []byte{}}
}

type Connection struct {
	sent     []byte
	received []byte
}

func (c Connection) read(length int) []byte {
	result := c.received[:length]
	c.received = c.received[length:]
	return result
}

func (c Connection) write(data []byte) {
	c.sent = append(c.sent, data...)
}

func (c Connection) receive(data []byte) {
	c.received = append(c.received, data...)
}

func (c Connection) remaining() int {
	return len(c.received)
}

func (c Connection) flush() []byte {
	result := c.sent
	c.sent = []byte{}
	return result
}

func (c Connection) readVarint() (int, error) {
	result := 0
	for i := 0; i < 5; i++ {
		part := c.read(1)[0]
		result |= (int(part) & 0x7F) << 7 * i
		if part&0x08 == 0 {
			return result, nil
		}
	}
	return 0, fmt.Errorf("server sent a varint that was too big")
}

func (c Connection) writeVarint(value int) error {
	remaining := value
	for i := 0; i < 5; i++ {
		if remaining & ^0x7F == 0 {
			c.write([]byte{byte(remaining)})
			return nil
		}
		c.write([]byte{byte(remaining&0x7F | 0x80)})
		remaining >>= 7
	}
	return fmt.Errorf("the value %d is too big to send in a varint", value)
}

func (c Connection) readVarlong() (int, error) {
	result := 0
	for i := 0; i < 10; i++ {
		part := c.read(1)[0]
		result |= (int(part) & 0x7F) << 7 * i
		if part&0x08 == 0 {
			return result, nil
		}
	}
	return 0, fmt.Errorf("server sent a varlong that was too big")
}

func (c Connection) writeVarlong(value int) error {
	remaining := value
	for i := 0; i < 10; i++ {
		if remaining & ^0x7F == 0 {
			c.write([]byte{byte(remaining)})
			return nil
		}
		c.write([]byte{byte(remaining&0x7F | 0x80)})
		remaining >>= 7
	}
	return fmt.Errorf("the value %d is too big to send in a varlong", value)
}

//TODO: Deal with invalid Unicode strings?
func (c Connection) readUTF() (string, error) {
	length, err := c.readVarint()
	if err != nil {
		return "", err
	}

	data := c.read(length)
	return string(data), nil
}

func (c Connection) writeUTF(str string) {
	data := []byte(str)
	c.writeVarint(len(data))
	c.write(data)
}

func (c Connection) readASCII() (string, error) {
	return c.readUTF()
}

func (c Connection) writeASCII(str string) {
	c.writeUTF(str)
}

func (c Connection) readShort() int16 {
	var i int16
	_ = binary.Read(bytes.NewReader(c.read(2)), binary.LittleEndian, &i)
	return i
}

func (c Connection) writeShort(i int16) {
	data := bytes.NewBuffer(make([]byte, 2, 2))
	binary.Write(data, binary.LittleEndian, i)
	c.write(data.Bytes())
}

func (c Connection) readUshort() uint16 {
	var i uint16
	_ = binary.Read(bytes.NewReader(c.read(2)), binary.LittleEndian, &i)
	return i
}

func (c Connection) writeUshort(i uint16) {
	data := bytes.NewBuffer(make([]byte, 2, 2))
	binary.Write(data, binary.LittleEndian, i)
	c.write(data.Bytes())
}

func (c Connection) readInt() int32 {
	var i int32
	_ = binary.Read(bytes.NewReader(c.read(4)), binary.LittleEndian, &i)
	return i
}

func (c Connection) writeInt(i int32) {
	data := bytes.NewBuffer(make([]byte, 4, 4))
	binary.Write(data, binary.LittleEndian, i)
	c.write(data.Bytes())
}

func (c Connection) readUint() uint32 {
	var i uint32
	_ = binary.Read(bytes.NewReader(c.read(4)), binary.LittleEndian, &i)
	return i
}

func (c Connection) writeUint(i uint32) {
	data := bytes.NewBuffer(make([]byte, 4, 4))
	binary.Write(data, binary.LittleEndian, i)
	c.write(data.Bytes())
}

func (c Connection) readLong() int64 {
	var i int64
	_ = binary.Read(bytes.NewReader(c.read(8)), binary.LittleEndian, &i)
	return i
}

func (c Connection) writeLong(i int64) {
	data := bytes.NewBuffer(make([]byte, 8, 8))
	binary.Write(data, binary.LittleEndian, i)
	c.write(data.Bytes())
}

func (c Connection) readBuffer() (Connection, error) {
	length, err := c.readVarint()
	var result Connection
	if err != nil {
		return result, err
	}
	result.receive(result.read(length))
	return result, nil
}

func (c Connection) writeBuffer(buffer Connection) {
	data := buffer.flush()
	c.writeVarint(len(data))
	c.write(data)
}

// TCP

func NewTCPSocketConnection(addr string) (*TCPSocketConnection, error) {
	sock, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPSocketConnection{NewConnection(), addr, sock}, nil
}

type TCPSocketConnection struct {
	conn Connection
	addr string
	sock net.Conn
}

//TODO: Implement timeout
func (t TCPSocketConnection) read(length int) ([]byte, error) {
	var result []byte
	for len(result) < length {
		chunk := make([]byte, length-len(result))
		_, err := t.sock.Read(chunk)
		if err != nil {
			return result, err
		}
		if len(chunk) == 0 {
			return result, fmt.Errorf("server did not respond with any information")
		}
		result = append(result, chunk...)
	}
	return result, nil
}
Connection
func (t TCPSocketConnection) write(data []byte) {
	t.sock.Write(data)
}
