package mcstatus

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

func NewServerPinger(connection Connection, host string, port int, version int) ServerPinger {
	pingToken := int64(rand.Intn((1 << 63) - 1))
	s := ServerPinger{pingToken, version, connection, host, port}
	return s
}

type ServerPinger struct {
	pingToken  int64
	version    int
	connection Connection
	host       string
	port       int
}

func (s ServerPinger) handshake() {
	packet := NewConnection()
	packet.writeVarint(0)
	packet.writeVarint(s.version)
	packet.writeUTF(s.host)
	packet.writeUshort(uint16(s.port))
	packet.writeVarint(1)

	s.connection.writeBuffer(packet)
}

func (s ServerPinger) readStatus() (*PingResponse, error) {
	request := NewConnection()
	request.writeVarint(0)
	s.connection.writeBuffer(request)

	response, err := s.connection.readBuffer()
	if err != nil {
		return nil, err
	}

	v, err := response.readVarint()
	if err != nil {
		return nil, err
	}
	if v != 0 {
		return nil, fmt.Errorf("received invalid status response packet")
	}

	//Decode JSON
	rawStr, err := response.readUTF()
	if err != nil {
		return nil, err
	}
	var raw PingResponse
	err = json.Unmarshal([]byte(rawStr), &raw)
	if err != nil {
		return nil, fmt.Errorf("received invalid JSON")
	}
	return &raw, nil
}

func (s ServerPinger) testPing() (float64, error) {
	request := NewConnection()
	request.writeVarint(1)
	request.writeLong(int64(s.pingToken))
	sent := time.Now().UnixNano()
	s.connection.writeBuffer(request)

	response, err := s.connection.readBuffer()
	if err != nil {
		return -1, err
	}
	received := time.Now().UnixNano()
	i, err := response.readVarint()
	if err != nil {
		return -1, err
	}
	if i != 1 {
		return -1, fmt.Errorf("received invalid ping response packet")
	}
	receivedToken := response.readLong()
	if receivedToken != s.pingToken {
		return -1, fmt.Errorf("received mangled ping response packet (expected token %d, received %d)", s.pingToken, receivedToken)
	}
	return float64(received-sent) / 1000000000, nil
}

type PingResponse struct {
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
	Favicon string `json:"favicon"`
	Players struct {
		Max    int64 `json:"max"`
		Online int64 `json:"online"`
		Sample []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"sample"`
	} `json:"players"`
	Version struct {
		Name     string `json:"name"`
		Protocol int64  `json:"protocol"`
	} `json:"version"`
}
