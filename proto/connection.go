package connection

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

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

func (c Connection) readShort() int {
	var i int16
	_ = binary.Read(bytes.NewReader(c.read(2)), binary.LittleEndian, &i)
	return int(i)
}

func (c Connection) writeShort(i int) {
	data := make([]byte, 2)
	x := int16(i)
	binary.Write(bytes.NewBuffer(data), binary.LittleEndian, &x)
	c.write(data)
}
