package mcstatus

import (
	"strconv"
	"strings"
)

func NewServerQuerier(connection UDPSocketConnection) ServerQuerier {
	return ServerQuerier{connection, 0}
}

type ServerQuerier struct {
	connection UDPSocketConnection
	challenge  int
}

func (s *ServerQuerier) createPacket(id int) Connection {
	packet := NewConnection()
	packet.Write([]byte{0xFE, 0xFD})
	packet.Write([]byte{byte(id)})
	packet.WriteUint(0)
	packet.WriteInt(int32(s.challenge))
	return packet
}

func (s *ServerQuerier) readPacket() (*Connection, error) {
	packet := NewConnection()
	data, err := s.connection.Read(s.connection.Remaining())
	if err != nil {
		return nil, err
	}
	packet.Receive(data)
	_, err = packet.Read(1 + 4)
	if err != nil {
		return nil, err
	}
	return &packet, nil
}

func (s *ServerQuerier) handshake() error {
	pkt := s.createPacket(9)
	s.connection.Write(pkt.Flush())
	packet, err := s.readPacket()
	if err != nil {
		return err
	}
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

func (s *ServerQuerier) readQuery() (*QueryResponse, error) {
	request := s.createPacket(0)
	request.WriteUint(0)
	s.connection.Write(request.Flush())

	response, err := s.readPacket()
	if err != nil {
		return nil, err
	}
	response.Read(len("splitnum") + 1 + 1 + 1)

	data := make(map[string]string)
	players := make([]string, 0)

	for {
		key, err := response.ReadASCII()
		if err != nil {
			return nil, err
		}
		if len(key) == 0 {
			response.Read(1)
			break
		}
		value, err := response.ReadASCII()
		if err != nil {
			return nil, err
		}
		data[key] = value
	}

	response.Read(len("player_") + 1 + 1)

	for {
		name, err := response.ReadASCII()
		if err != nil {
			return nil, err
		}
		if len(name) == 0 {
			break
		}
		players = append(players, name)
	}

	q, err := newQueryResponse(data, players)
	return q, err
}

func newQueryResponse(raw map[string]string, players []string) (*QueryResponse, error) {
	numplayers, err := strconv.Atoi(raw["numplayers"])
	if err != nil {
		return nil, err
	}
	maxplayers, err := strconv.Atoi(raw["maxplayers"])
	if err != nil {
		return nil, err
	}

	version := raw["version"]
	brand := "vanilla"
	var pluginList []string
	plugins := raw["plugins"]
	if len(plugins) > 0 {
		parts := strings.Split(plugins, ":")
		brand = strings.Trim(parts[0], " ")
		if len(parts) == 2 {
			for _, s := range strings.Split(parts[1], ";") {
				pluginList = append(pluginList, strings.Trim(s, " "))
			}
		}
	}

	q := QueryResponse{
		raw,
		raw["hostname"],
		raw["map"],
		Players{
			numplayers,
			maxplayers,
			players,
		},
		Software{version,
			brand,
			pluginList,
		},
	}
	return &q, nil
}

type QueryResponse struct {
	raw      map[string]string
	motd     string
	worldmap string
	players  Players
	software Software
}
type Players struct {
	online int
	max    int
	names  []string
}
type Software struct {
	version string
	brand   string
	plugins []string
}
