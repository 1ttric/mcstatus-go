package mcstatus

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func NewMinecraftServer(addr string, timeout int) (*MinecraftServer, error) {
	host, port, err := Lookup(addr)
	if err != nil {
		return nil, err
	}
	return &MinecraftServer{host, port, timeout}, nil
}

type MinecraftServer struct {
	host    string
	port    int
	timeout int
}

func (m MinecraftServer) Query() (*QueryResponse, error) {
	host := m.host
	ips, err := net.LookupHost(m.host)
	if err != nil {
		return nil, err
	}
	if len(ips) > 0 {
		host = ips[0]
	}
	connection, err := NewUDPSocketConnection(fmt.Sprintf("%s:%d", host, m.port), m.timeout)
	if err != nil {
		return nil, err
	}
	querier := NewServerQuerier(*connection)
	err = querier.handshake()
	if err != nil {
		return nil, err
	}
	response, err := querier.readQuery()
	if err != nil {
		return nil, err
	}
	return response, nil
}

func Lookup(address string) (string, int, error) {
	host := address
	port := -1
	if strings.Contains(address, ":") {
		parts := strings.Split(address, ":")
		if len(parts) > 2 {
			return "", 0, fmt.Errorf("invalid address '%s'", address)
		}
		host = parts[0]
		var err error
		port, err = strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, fmt.Errorf("invalid address '%s'", address)
		}
	}
	if port == -1 {
		port = 25665
		_, addrs, _ := net.LookupSRV("minecraft", "tcp", host)
		if len(addrs) > 0 {
			answer := *addrs[0]
			host = answer.Target
			port = int(answer.Port)
		}
	}
	return host, port, nil
}
