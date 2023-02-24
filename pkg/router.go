package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/AbandonTech/minecraftrouter/pkg/resolver"
	"net"
	"strconv"
	"strings"
)

func SplitAddr(addr net.Addr) (string, uint16, error) {
	split := strings.Split(addr.String(), ":")
	address := split[0]
	port, err := strconv.ParseUint(split[1], 10, 16)

	if err != nil {
		return "", 0, err
	}

	return address, uint16(port), nil
}

// CreateProxyProtocolHeader for Minecraft proxy protocol
// "PROXY <TCP IPv4 or IPv6> <source IP> <destination IP> <source port> <destination port>"
// https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt
func CreateProxyProtocolHeader(source net.Addr, dest net.Addr) ([]byte, error) {
	sourceAddr, sourcePort, err := SplitAddr(source)
	if err != nil {
		return nil, err
	}

	destAddr, destPort, err := SplitAddr(dest)
	if err != nil {
		return nil, err
	}

	header := fmt.Sprintf("PROXY TCP4 %s %s %d %d\r\n", sourceAddr, destAddr, sourcePort, destPort)
	return []byte(header), nil
}

type Router struct {
	resolver resolver.Resolver
	address  string
}

func (r Router) Run() error {
	fmt.Printf("Listening on %s\n", r.address)
	listener, err := net.Listen("tcp", r.address)
	if err != nil {
		return err
	}

	defer listener.Close()

	for {
		client, err := listener.Accept()
		if err != nil {
			return err
		}

		packet := make([]byte, 1024)
		_, err = client.Read(packet)
		if err != nil {
			return err
		}

		packetReader := bytes.NewReader(packet)

		// Walk through buffer, these values are not required tho
		_, err = binary.ReadUvarint(packetReader)
		if err != nil {
			return err
		}
		_, err = binary.ReadUvarint(packetReader)
		if err != nil {
			return err
		}
		_, err = binary.ReadUvarint(packetReader)
		if err != nil {
			return err
		}

		serverAddressLength, _ := binary.ReadUvarint(packetReader)
		serverAddressRaw := make([]byte, serverAddressLength)
		_, err = packetReader.Read(serverAddressRaw)
		if err != nil {
			return err
		}
		serverAddress := string(serverAddressRaw)

		serverPortRaw := make([]byte, 2)
		_, err = packetReader.Read(serverPortRaw)
		if err != nil {
			return err
		}
		serverPort := binary.BigEndian.Uint16(serverPortRaw)

		requestedAddress := fmt.Sprintf("%s:%d", serverAddress, serverPort)
		resolvedAddress, ok := r.resolver.ResolveHostname(requestedAddress)
		if !ok {
			fmt.Printf("could not resolve hostname %s\n", requestedAddress)
			client.Close()
			continue
		}

		server, err := net.Dial("tcp", resolvedAddress)
		if err != nil {
			return err
		}

		header, err := CreateProxyProtocolHeader(client.RemoteAddr(), server.RemoteAddr())
		if err != nil {
			return err
		}

		_, err = server.Write(header)
		if err != nil {
			return err
		}

		_, err = server.Write(packet)
		if err != nil {
			return err
		}

		go ProxyForever(client, server)
	}
}

func NewRouter(address string, resolver resolver.Resolver) Router {
	return Router{
		resolver: resolver,
		address:  address,
	}
}
