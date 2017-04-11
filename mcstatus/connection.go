package mcstatus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"time"
	"unicode/utf8"
)

// Connection buffer

func NewConnection() Connection {
	return Connection{[]byte{}, []byte{}}
}

type Connection struct {
	sent     []byte
	received []byte
}

func (c *Connection) Read(length int) ([]byte, error) {
	end := int(math.Min(float64(length), float64(len(c.received))))
	result := c.received[:end]
	c.received = c.received[end:]
	return result, nil
}

func (c *Connection) Write(data []byte) {
	c.sent = append(c.sent, data...)
}

func (c *Connection) Receive(data []byte) {
	c.received = append(c.received, data...)
}

func (c *Connection) Remaining() int {
	return len(c.received)
}

func (c *Connection) Flush() []byte {
	result := c.sent
	c.sent = []byte{}
	return result
}

func (c *Connection) ReadVarInt() (int, error) {
	result := 0
	for i := 0; i < 5; i++ {
		data, err := c.Read(1)
		if err != nil {
			return 0, err
		}
		if len(data) == 0 {
			return 0, fmt.Errorf("cannot parse, incomplete data")
		}
		part := data[0]
		result |= (int(part) & 0x7F) << uint(7*i)
		if part&0x80 == 0 {
			return result, nil
		}
	}
	return 0, fmt.Errorf("server sent a varint that was too big")
}

func (c *Connection) WriteVarInt(value int) error {
	remaining := value
	for i := 0; i < 5; i++ {
		if remaining & ^0x7F == 0 {
			c.Write([]byte{byte(remaining)})
			return nil
		}
		c.Write([]byte{byte(remaining&0x7F | 0x80)})
		remaining >>= 7
	}
	return fmt.Errorf("value is too big to send in a varint")
}

func (c *Connection) ReadVarLong() (int, error) {
	result := 0
	for i := 0; i < 10; i++ {
		data, err := c.Read(1)
		if err != nil {
			return 0, err
		}
		part := data[0]
		result |= (int(part) & 0x7F) << uint(7*i)
		if part&0x80 == 0 {
			return result, nil
		}
	}
	return 0, fmt.Errorf("server sent a varlong that was too big")
}

func (c *Connection) WriteVarLong(value int) error {
	remaining := value
	for i := 0; i < 10; i++ {
		if remaining & ^0x7F == 0 {
			c.Write([]byte{byte(remaining)})
			return nil
		}
		c.Write([]byte{byte(remaining&0x7F | 0x80)})
		remaining >>= 7
	}
	return fmt.Errorf("the value %d is too big to send in a varlong", value)
}

//TODO: Deal with invalid Unicode strings?
func (c *Connection) ReadUTF() (string, error) {
	length, err := c.ReadVarInt()
	if err != nil {
		return "", err
	}

	data, err := c.Read(length)
	if err != nil {
		return "", err
	}
	str := ""
	for len(data) > 0 {
		char, size := utf8.DecodeRune(data)
		data = data[size:]
		str = str + string(char)
	}
	return str, nil
}

func (c *Connection) WriteUTF(str string) {
	data := []byte(str)
	c.WriteVarInt(len(data))
	c.Write(data)
}

func (c *Connection) ReadASCII() (string, error) {
	result := []byte{}
	for (len(result) == 0) || (result[len(result)-1] != byte(0)) {
		char, err := c.Read(1)
		if err != nil {
			return "", err
		}
		if len(char) < 1 {
			return "", fmt.Errorf("cannot parse, incomplete data")
		}
		result = append(result, char[0])
	}
	return string(result[:len(result)-1]), nil
}

func (c *Connection) WriteASCII(str string) {
	data := []byte(str)
	data = append(data, byte(0x00))
	c.Write(data)
}

func (c *Connection) ReadShort() (int16, error) {
	var i int16
	data, err := c.Read(2)
	if err != nil {
		return 0, err
	}
	err = binary.Read(bytes.NewReader(data), binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (c *Connection) WriteShort(i int16) {
	data := bytes.NewBuffer(make([]byte, 0, 2))
	binary.Write(data, binary.BigEndian, i)
	c.Write(data.Bytes())
}

func (c *Connection) ReadUshort() (uint16, error) {
	var i uint16
	data, err := c.Read(2)
	if err != nil {
		return 0, err
	}
	err = binary.Read(bytes.NewReader(data), binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (c *Connection) WriteUshort(i uint16) {
	data := bytes.NewBuffer(make([]byte, 0, 2))
	binary.Write(data, binary.BigEndian, i)
	c.Write(data.Bytes())
}

func (c *Connection) ReadInt() (int32, error) {
	var i int32
	data, err := c.Read(4)
	if err != nil {
		return 0, err
	}
	err = binary.Read(bytes.NewReader(data), binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (c *Connection) WriteInt(i int32) {
	data := bytes.NewBuffer(make([]byte, 0, 4))
	binary.Write(data, binary.BigEndian, i)
	c.Write(data.Bytes())
}

func (c *Connection) ReadUint() (uint32, error) {
	var i uint32
	data, err := c.Read(4)
	if err != nil {
		return 0, err
	}
	err = binary.Read(bytes.NewReader(data), binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (c *Connection) WriteUint(i uint32) {
	data := bytes.NewBuffer(make([]byte, 0, 4))
	binary.Write(data, binary.BigEndian, i)
	c.Write(data.Bytes())
}

func (c *Connection) ReadLong() (int64, error) {
	var i int64
	data, err := c.Read(8)
	if err != nil {
		return 0, err
	}
	err = binary.Read(bytes.NewReader(data), binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (c *Connection) WriteLong(i int64) {
	data := bytes.NewBuffer(make([]byte, 0, 8))
	binary.Write(data, binary.BigEndian, i)
	c.Write(data.Bytes())
}

func (c *Connection) ReadULong() (uint64, error) {
	var i uint64
	data, err := c.Read(8)
	if err != nil {
		return 0, err
	}
	err = binary.Read(bytes.NewReader(data), binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (c *Connection) WriteULong(i uint64) {
	data := bytes.NewBuffer(make([]byte, 0, 8))
	binary.Write(data, binary.BigEndian, i)
	c.Write(data.Bytes())
}

func (c *Connection) ReadBuffer() (*Connection, error) {
	length, err := c.ReadVarInt()
	if err != nil {
		return nil, err
	}
	var result Connection
	data, err := c.Read(length)
	if err != nil {
		return nil, err
	}
	result.Receive(data)
	return &result, nil
}

func (c *Connection) WriteBuffer(buffer Connection) {
	data := buffer.Flush()
	c.WriteVarInt(len(data))
	c.Write(data)
}

// TCP

func NewTCPSocketConnection(addr string, timeout int) (*TCPSocketConnection, error) {
	sock, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPSocketConnection{NewConnection(), addr, sock, timeout}, nil
}

type TCPSocketConnection struct {
	conn    Connection
	addr    string
	sock    net.Conn
	timeout int
}

//TODO: Implement timeout
func (t *TCPSocketConnection) Read(length int) ([]byte, error) {
	var result []byte
	for len(result) < length {
		chunk := make([]byte, length-len(result))
		t.sock.SetDeadline(time.Now().Add(time.Duration(t.timeout) * time.Millisecond))
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

func (t *TCPSocketConnection) Write(data []byte) {
	t.sock.SetDeadline(time.Now().Add(time.Duration(t.timeout) * time.Millisecond))
	t.sock.Write(data)
}

// UDP

func NewUDPSocketConnection(addr string, timeout int) (*UDPSocketConnection, error) {
	UDPAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	sock, err := net.DialUDP("udp", nil, UDPAddr)
	if err != nil {
		return nil, err
	}
	return &UDPSocketConnection{NewConnection(), addr, *sock, timeout}, nil
}

type UDPSocketConnection struct {
	conn    Connection
	addr    string
	sock    net.UDPConn
	timeout int
}

//TODO: Implement timeout with UDPConn.SetDeadline()
func (u *UDPSocketConnection) Read(length int) ([]byte, error) {
	fmt.Printf("UDP read with timeout %d\n", u.timeout)
	result := make([]byte, 65535)
	i := 0
	var err error
	for i == 0 {
		u.sock.SetDeadline(time.Now().Add(time.Duration(u.timeout) * time.Millisecond))
		i, _, err = u.sock.ReadFromUDP(result)
		if err != nil {
			return []byte{}, err
		}
	}
	return result[:i], nil
}

func (u *UDPSocketConnection) Write(data []byte) {
	u.sock.SetDeadline(time.Now().Add(time.Duration(u.timeout) * time.Millisecond))
	u.sock.Write(data)
}

func (u *UDPSocketConnection) Remaining() int {
	return 65535
}
