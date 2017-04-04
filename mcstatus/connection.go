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

func (c Connection) Read(length int) []byte {
	result := c.received[:length]
	c.received = c.received[length:]
	return result
}

func (c Connection) Write(data []byte) {
	c.sent = append(c.sent, data...)
}

func (c Connection) Receive(data []byte) {
	c.received = append(c.received, data...)
}

func (c Connection) Remaining() int {
	return len(c.received)
}

func (c Connection) Flush() []byte {
	result := c.sent
	c.sent = []byte{}
	return result
}

func (c Connection) ReadVarint() (int, error) {
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

func (c Connection) WriteVarint(value int) error {
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

func (c Connection) ReadVarlong() (int, error) {
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

func (c Connection) WriteVarlong(value int) error {
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
func (c Connection) ReadUTF() (string, error) {
	length, err := c.readVarint()
	if err != nil {
		return "", err
	}

	data := c.read(length)
	return string(data), nil
}

func (c Connection) WriteUTF(str string) {
	data := []byte(str)
	c.writeVarint(len(data))
	c.write(data)
}

func (c Connection) ReadASCII() (string, error) {
	return c.readUTF()
}

func (c Connection) WriteASCII(str string) {
	c.writeUTF(str)
}

func (c Connection) ReadShort() int16 {
	var i int16
	_ = binary.Read(bytes.NewReader(c.read(2)), binary.LittleEndian, &i)
	return i
}

func (c Connection) WriteShort(i int16) {
	data := bytes.NewBuffer(make([]byte, 2, 2))
	binary.Write(data, binary.LittleEndian, i)
	c.write(data.Bytes())
}

func (c Connection) ReadUshort() uint16 {
	var i uint16
	_ = binary.Read(bytes.NewReader(c.read(2)), binary.LittleEndian, &i)
	return i
}

func (c Connection) WriteUshort(i uint16) {
	data := bytes.NewBuffer(make([]byte, 2, 2))
	binary.Write(data, binary.LittleEndian, i)
	c.write(data.Bytes())
}

func (c Connection) ReadInt() int32 {
	var i int32
	_ = binary.Read(bytes.NewReader(c.read(4)), binary.LittleEndian, &i)
	return i
}

func (c Connection) WriteInt(i int32) {
	data := bytes.NewBuffer(make([]byte, 4, 4))
	binary.Write(data, binary.LittleEndian, i)
	c.write(data.Bytes())
}

func (c Connection) ReadUint() uint32 {
	var i uint32
	_ = binary.Read(bytes.NewReader(c.read(4)), binary.LittleEndian, &i)
	return i
}

func (c Connection) WriteUint(i uint32) {
	data := bytes.NewBuffer(make([]byte, 4, 4))
	binary.Write(data, binary.LittleEndian, i)
	c.write(data.Bytes())
}

func (c Connection) ReadLong() int64 {
	var i int64
	_ = binary.Read(bytes.NewReader(c.read(8)), binary.LittleEndian, &i)
	return i
}

func (c Connection) WriteLong(i int64) {
	data := bytes.NewBuffer(make([]byte, 8, 8))
	binary.Write(data, binary.LittleEndian, i)
	c.write(data.Bytes())
}

func (c Connection) ReadBuffer() (Connection, error) {
	length, err := c.readVarint()
	var result Connection
	if err != nil {
		return result, err
	}
	result.receive(result.read(length))
	return result, nil
}

func (c Connection) WriteBuffer(buffer Connection) {
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

func (t TCPSocketConnection) write(data []byte) {
	t.sock.Write(data)
}
