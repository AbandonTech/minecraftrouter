package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/AbandonTech/minecraftrouter/pkg/resolver"
	"github.com/rs/zerolog/log"
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
	resolver      resolver.Resolver
	address       string
	proxyProtocol bool
}

func (r Router) Run() error {
	log.Info().
		Str("Address", r.address).
		Msg("Listening for connections")

	listener, err := net.Listen("tcp", r.address)
	if err != nil {
		return err
	}

	defer func(listener net.Listener) {
		log.Info().Msg("Router exiting")
		err := listener.Close()
		if err != nil {
			log.Fatal().Msg("Could not close listener")
		}
	}(listener)

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Unable to accept connection")
		}

		go r.handleConnection(client)
	}
}

func (r Router) handleConnection(client net.Conn) {
	// Overwritting name to use this as a contextual log
	log := log.With().
		Stringer("Client", client.RemoteAddr()).
		Logger()

	log.Debug().
		Msg("Connected")

	packet := make([]byte, 1024)
	handshakeSize, err := client.Read(packet)
	if err != nil {
		log.Err(err).Msg("Unable to read packet")
		return
	}

	packetReader := bytes.NewReader(packet)

	// Walk through buffer, these values are not required tho
	for i := 0; i < 3; i++ {
		_, err = binary.ReadUvarint(packetReader)
		if err != nil {
			log.Info().
				Msg("Received erroneous handshake packet. Closing connection.")
			client.Close()
			return
		}
	}

	serverAddressLength, _ := binary.ReadUvarint(packetReader)
	serverAddressRaw := make([]byte, serverAddressLength)
	_, err = packetReader.Read(serverAddressRaw)
	if err != nil {
		log.Info().
			Msg("Could not read server address from login packet. Closing connection.")
		client.Close()
		return
	}
	serverAddress := string(serverAddressRaw)

	// Its common for proxies or mods to append extra data to the server address after a "///" separator. We ignore this for routing.
	serverAddress = strings.Split(serverAddress, "///")[0]

	// Forge's "FML" protocol appends a marker to the end of the server address prefixed with a null character.
	serverAddress = strings.Split(serverAddress, "\x00")[0]

	serverPortRaw := make([]byte, 2)
	_, err = packetReader.Read(serverPortRaw)
	if err != nil {
		log.Info().
			Msg("Could not read server port from login packet. Closing connection.")
		client.Close()
		return
	}
	serverPort := binary.BigEndian.Uint16(serverPortRaw)

	requestedAddress := fmt.Sprintf("%s:%d", serverAddress, serverPort)
	resolvedAddress, ok := r.resolver.ResolveHostname(requestedAddress)
	if !ok {
		log.Warn().
			Str("Host", serverAddress).
			Uint16("Port", serverPort).
			Msg("Could not resolve hostname. Closing connection.")
		client.Close()
		return
	}
	log.Debug().
		Str("ResolvedAddress", resolvedAddress).
		Msg("Resolved hostname to address.")

	server, err := net.Dial("tcp", resolvedAddress)
	if err != nil {
		log.Error().
			Err(err).
			Str("MinecraftServer", resolvedAddress).
			Msg("Could not connect to minecraft server. Closing connection.")
		client.Close()
		return
	}

	if r.proxyProtocol {
		header, err := CreateProxyProtocolHeader(client.RemoteAddr(), server.RemoteAddr())
		if err != nil {
			log.Error().
				Err(err).
				Stringer("Client", client.RemoteAddr()).
				Stringer("Server", server.RemoteAddr()).
				Msg("Unable to create proxy protocol header. Closing connection.")
			client.Close()
			return
		}

		_, err = server.Write(header)
		if err != nil {
			log.Error().
				Err(err).
				Stringer("Server", server.RemoteAddr()).
				Msg("Unable to write proxy protocol header to server. Closing connections.")
			server.Close()
			client.Close()
			return
		}
	}

	_, err = server.Write(packet[:handshakeSize])
	if err != nil {
		log.Error().
			Err(err).
			Stringer("Server", server.RemoteAddr()).
			Msg("Unable to write handshake packet to server. Closing connections.")
		server.Close()
		client.Close()
		return
	}

	go ProxyForever(client, server)
}

func NewRouter(address string, resolver resolver.Resolver, proxyProtocol bool) Router {
	return Router{
		resolver:      resolver,
		address:       address,
		proxyProtocol: proxyProtocol,
	}
}
