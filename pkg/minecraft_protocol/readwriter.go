package minecraft_protocol

import "net"

type Packet = []byte

type ReadWriter struct {
	Writer
	Reader
	con net.Conn
}

func NewReadWriter(conn net.Conn) ReadWriter {
	return ReadWriter{
		Writer: Writer{},
		Reader: Reader{},
		con:    conn,
	}
}
