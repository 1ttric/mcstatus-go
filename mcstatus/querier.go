package mcstatus

import (
	"encoding/json"
	"strconv"
)

func NewServerQuerier() ServerQuerier {
	return ServerQuerier{NewConnection(), 0}
}

type ServerQuerier struct {
	connection Connection
	challenge  int
}

func (s ServerQuerier) createPacket(id int) Connection {
	packet := NewConnection()
	packet.Write([]byte{0xFE, 0xFD})
	packet.Write([]byte{byte(id)})
	packet.WriteUint(0)
	packet.WriteInt(int32(s.challenge))
	return packet
}

func (s ServerQuerier) readPacket() Connection {
	packet := NewConnection()
	packet.Receive(s.connection.Read(s.connection.Remaining()))
	packet.Read(1 + 4)
	return packet
}

func (s ServerQuerier) handshake() error {
	s.connection.Write(s.createPacket(9).Flush())

	packet := s.readPacket()
	str, err := packet.ReadASCII()
	if err != nil {
		return err
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	s.challenge = i
	return nil
}

func (s ServerQuerier) readQuery() QueryResponse {
	request := s.createPacket(0)
	request.WriteUint(0)
	s.connection.Write(request.Flush())

	response := s.readPacket()
	response.Read(len("splitnum") + 1 + 1 + 1)

	var data QueryResponse
	json.Unmarshal(response.Flush(), data)
	return data
}

type QueryResponse struct {
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
		Name       string `json:"name"`
		versioncol int64  `json:"versioncol"`
	} `json:"version"`
}
